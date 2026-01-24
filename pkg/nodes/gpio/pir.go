//go:build linux
// +build linux

package gpio

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stianeikeland/go-rpio/v4"
)

// PIRNode handles PIR motion sensor input
type PIRNode struct {
	pin           rpio.Pin
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

	// Initialize GPIO
	if err := rpio.Open(); err != nil {
		return fmt.Errorf("failed to open GPIO: %w", err)
	}

	n.pin = rpio.Pin(n.pinNumber)
	n.pin.Input()

	// Set pull mode
	switch n.pullMode {
	case "up":
		n.pin.PullUp()
	case "down":
		n.pin.PullDown()
	case "off":
		n.pin.PullOff()
	default:
		n.pin.PullDown() // Default to pull down
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
	if msgPayload, ok := msg.Payload.(map[string]interface{}); ok {
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

// monitorMotion continuously monitors the PIR sensor
func (n *PIRNode) monitorMotion() {
	defer n.wg.Done()

	ticker := time.NewTicker(n.debounce)
	defer ticker.Stop()

	var lastState rpio.State
	var stateChangeTime time.Time

	for {
		select {
		case <-n.ctx.Done():
			return

		case <-ticker.C:
			currentState := n.pin.Read()

			// Detect state change
			if currentState != lastState {
				stateChangeTime = time.Now()
				lastState = currentState

				// Check if enough time has passed since last trigger (retrigger prevention)
				n.mu.RLock()
				timeSinceLastTrigger := time.Since(n.lastTrigger)
				n.mu.RUnlock()

				// Determine if this state change should trigger an event
				shouldTrigger := false
				motionDetected := false

				switch n.triggerMode {
				case "rising":
					if currentState == rpio.High && timeSinceLastTrigger >= n.retriggerTime {
						shouldTrigger = true
						motionDetected = true
					}
				case "falling":
					if currentState == rpio.Low && timeSinceLastTrigger >= n.retriggerTime {
						shouldTrigger = true
						motionDetected = false
					}
				case "both":
					if timeSinceLastTrigger >= n.retriggerTime {
						shouldTrigger = true
						motionDetected = currentState == rpio.High
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
							"state":     currentState,
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

// Cleanup stops monitoring and closes GPIO
func (n *PIRNode) Cleanup() error {
	if n.cancel != nil {
		n.cancel()
	}

	n.wg.Wait()

	if n.outputChan != nil {
		close(n.outputChan)
	}

	return rpio.Close()
}

// NewPIRExecutor creates a new PIR executor
func NewPIRExecutor(config map[string]interface{}) (node.Executor, error) {
	executor := NewPIRNode()
	if err := executor.Init(config); err != nil {
		return nil, err
	}
	return executor, nil
}
