//go:build linux
// +build linux

package gpio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// PIRNode handles PIR motion sensor input via the HAL GPIO provider.
type PIRNode struct {
	pinNumber     int
	debounce      time.Duration
	sensitivity   string
	lastTrigger   time.Time
	triggered     bool
	outputChan    chan node.Message
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	mu            sync.RWMutex
	pullMode      string
	triggerMode   string // "rising", "falling", "both"
	retriggerTime time.Duration
	halInstance   hal.HAL
}

// NewPIRNode creates a new PIR sensor node
func NewPIRNode() *PIRNode {
	return &PIRNode{
		debounce:      50 * time.Millisecond,
		sensitivity:   "normal",
		outputChan:    make(chan node.Message, 10),
		pullMode:      "down",
		triggerMode:   "rising",
		retriggerTime: 2 * time.Second,
	}
}

// Init initializes the PIR sensor node
func (n *PIRNode) Init(config map[string]interface{}) error {
	// Parse pin number
	if pin, ok := config["pin"].(float64); ok {
		n.pinNumber = int(pin)
	} else if pin, ok := config["pin"].(int); ok {
		n.pinNumber = pin
	} else {
		return fmt.Errorf("pin number is required")
	}

	// Parse debounce
	if debounceStr, ok := config["debounce"].(string); ok {
		duration, err := time.ParseDuration(debounceStr)
		if err != nil {
			return fmt.Errorf("invalid debounce duration: %w", err)
		}
		n.debounce = duration
	}

	// Parse sensitivity
	if sensitivity, ok := config["sensitivity"].(string); ok {
		n.sensitivity = sensitivity
	}

	// Parse pull mode
	if pullMode, ok := config["pullMode"].(string); ok {
		n.pullMode = pullMode
	}

	// Parse trigger mode
	if triggerMode, ok := config["triggerMode"].(string); ok {
		n.triggerMode = triggerMode
	}

	// Parse retrigger time
	if retriggerStr, ok := config["retriggerTime"].(string); ok {
		duration, err := time.ParseDuration(retriggerStr)
		if err != nil {
			return fmt.Errorf("invalid retrigger time: %w", err)
		}
		n.retriggerTime = duration
	}

	// Get HAL and configure GPIO pin
	h, err := hal.GetGlobalHAL()
	if err != nil {
		return fmt.Errorf("failed to get HAL: %w", err)
	}
	n.halInstance = h

	gpio := h.GPIO()
	if err := gpio.SetMode(n.pinNumber, hal.Input); err != nil {
		return fmt.Errorf("failed to set pin %d as input: %w", n.pinNumber, err)
	}

	// Set pull mode
	var pull hal.PullMode
	switch n.pullMode {
	case "up":
		pull = hal.PullUp
	case "down":
		pull = hal.PullDown
	case "off":
		pull = hal.PullNone
	default:
		pull = hal.PullDown
	}
	if err := gpio.SetPull(n.pinNumber, pull); err != nil {
		return fmt.Errorf("failed to set pull on pin %d: %w", n.pinNumber, err)
	}

	return nil
}

// Start begins monitoring the PIR sensor
func (n *PIRNode) Start(ctx context.Context) error {
	n.mu.Lock()
	if n.ctx != nil {
		n.mu.Unlock()
		return fmt.Errorf("PIR sensor already started")
	}

	n.ctx, n.cancel = context.WithCancel(ctx)
	n.mu.Unlock()

	n.wg.Add(1)
	go n.monitorMotion()

	return nil
}

// Execute processes incoming messages (for manual trigger/query)
func (n *PIRNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// PIR node primarily outputs data asynchronously
	// But can respond to queries about current state
	if msgPayload := msg.Payload; msgPayload != nil {
		if cmd, ok := msgPayload["command"].(string); ok {
			switch cmd {
			case "status":
				n.mu.RLock()
				triggered := n.triggered
				lastTrigger := n.lastTrigger
				n.mu.RUnlock()

				return node.Message{
					Type: node.MessageTypeData,
					Payload: map[string]interface{}{
						"motion":      triggered,
						"lastTrigger": lastTrigger.Unix(),
						"pin":         n.pinNumber,
					},
					Topic: msg.Topic,
				}, nil

			case "reset":
				n.mu.Lock()
				n.triggered = false
				n.lastTrigger = time.Time{}
				n.mu.Unlock()

				return node.Message{
					Type: node.MessageTypeData,
					Payload: map[string]interface{}{
						"reset": true,
					},
					Topic: msg.Topic,
				}, nil
			}
		}
	}

	return node.Message{}, nil
}

// monitorMotion continuously monitors the PIR sensor via HAL
func (n *PIRNode) monitorMotion() {
	defer n.wg.Done()

	gpio := n.halInstance.GPIO()

	ticker := time.NewTicker(n.debounce)
	defer ticker.Stop()

	var lastState bool
	var stateChangeTime time.Time

	for {
		select {
		case <-n.ctx.Done():
			return

		case <-ticker.C:
			currentVal, err := gpio.DigitalRead(n.pinNumber)
			if err != nil {
				continue
			}

			// Detect state change
			if currentVal != lastState {
				stateChangeTime = time.Now()
				lastState = currentVal

				// Check if enough time has passed since last trigger
				n.mu.RLock()
				timeSinceLastTrigger := time.Since(n.lastTrigger)
				n.mu.RUnlock()

				// Determine if this state change should trigger an event
				shouldTrigger := false
				motionDetected := false

				switch n.triggerMode {
				case "rising":
					if currentVal && timeSinceLastTrigger >= n.retriggerTime {
						shouldTrigger = true
						motionDetected = true
					}
				case "falling":
					if !currentVal && timeSinceLastTrigger >= n.retriggerTime {
						shouldTrigger = true
						motionDetected = false
					}
				case "both":
					if timeSinceLastTrigger >= n.retriggerTime {
						shouldTrigger = true
						motionDetected = currentVal
					}
				}

				if shouldTrigger {
					n.mu.Lock()
					n.triggered = motionDetected
					n.lastTrigger = stateChangeTime
					n.mu.Unlock()

					// Send motion event
					msg := node.Message{
						Type: node.MessageTypeData,
						Payload: map[string]interface{}{
							"motion":    motionDetected,
							"timestamp": stateChangeTime.Unix(),
							"pin":       n.pinNumber,
							"state":     currentVal,
						},
						Topic: "pir/motion",
					}

					select {
					case n.outputChan <- msg:
					default:
						// Channel full, skip
					}
				}
			}
		}
	}
}

// GetOutputChannel returns the channel for motion events
func (n *PIRNode) GetOutputChannel() <-chan node.Message {
	return n.outputChan
}

// Cleanup stops monitoring and releases resources
func (n *PIRNode) Cleanup() error {
	if n.cancel != nil {
		n.cancel()
	}

	n.wg.Wait()

	if n.outputChan != nil {
		close(n.outputChan)
	}

	return nil
}

// NewPIRExecutor creates a new PIR executor
func NewPIRExecutor(config map[string]interface{}) (node.Executor, error) {
	executor := NewPIRNode()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
