//go:build linux

package gpio

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/hal"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// InterruptNode provides dedicated GPIO interrupt/edge detection
type InterruptNode struct {
	pin         int
	edge        string // rising, falling, both
	debounceMs  int
	pullMode    string // up, down, off
	count       int64
	halInstance hal.HAL
	outputChan  chan node.Message
	stopChan    chan struct{}
	mu          sync.RWMutex
	running     bool
	lastState   bool
	lastTrigger time.Time
}

// NewInterruptExecutor creates a new interrupt node executor
func NewInterruptExecutor() node.Executor {
	return &InterruptNode{
		edge:       "rising",
		debounceMs: 50,
		pullMode:   "off",
		outputChan: make(chan node.Message, 100),
		stopChan:   make(chan struct{}),
	}
}

// Init initializes the interrupt node
func (n *InterruptNode) Init(config map[string]interface{}) error {
	if pin, ok := config["pin"].(float64); ok {
		n.pin = int(pin)
	} else {
		return fmt.Errorf("interrupt node requires a pin number")
	}
	if edge, ok := config["edge"].(string); ok {
		n.edge = edge
	}
	if debounce, ok := config["debounceMs"].(float64); ok {
		n.debounceMs = int(debounce)
	}
	if pull, ok := config["pullMode"].(string); ok {
		n.pullMode = pull
	}

	h, err := hal.GetGlobalHAL()
	if err != nil {
		return fmt.Errorf("failed to get HAL: %w", err)
	}
	n.halInstance = h

	gpio := h.GPIO()
	gpio.SetMode(n.pin, hal.Input)

	switch n.pullMode {
	case "up":
		gpio.SetPull(n.pin, hal.PullUp)
	case "down":
		gpio.SetPull(n.pin, hal.PullDown)
	default:
		gpio.SetPull(n.pin, hal.PullNone)
	}

	return nil
}

// Execute handles interrupt events
func (n *InterruptNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Handle action commands
	if action, ok := msg.Payload["action"].(string); ok {
		switch action {
		case "reset_count":
			atomic.StoreInt64(&n.count, 0)
			return node.Message{
				Type: node.MessageTypeData,
				Payload: map[string]interface{}{
					"action": "reset_count",
					"count":  0,
				},
			}, nil
		case "get_count":
			return node.Message{
				Type: node.MessageTypeData,
				Payload: map[string]interface{}{
					"count": atomic.LoadInt64(&n.count),
					"pin":   n.pin,
				},
			}, nil
		}
	}

	// Start monitoring if not already running
	n.mu.Lock()
	if !n.running {
		n.running = true
		n.lastState, _ = n.halInstance.GPIO().DigitalRead(n.pin)
		go n.monitorInterrupts()
	}
	n.mu.Unlock()

	// Wait for interrupt event
	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case intMsg := <-n.outputChan:
		return intMsg, nil
	}
}

func (n *InterruptNode) monitorInterrupts() {
	debounce := time.Duration(n.debounceMs) * time.Millisecond
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-n.stopChan:
			return
		case <-ticker.C:
			currentState, _ := n.halInstance.GPIO().DigitalRead(n.pin)

			n.mu.RLock()
			lastState := n.lastState
			lastTrigger := n.lastTrigger
			n.mu.RUnlock()

			if currentState == lastState {
				continue
			}

			// Debounce check
			if time.Since(lastTrigger) < debounce {
				continue
			}

			// Check edge type
			edgeType := ""
			if !lastState && currentState {
				edgeType = "rising"
			} else if lastState && !currentState {
				edgeType = "falling"
			}

			shouldTrigger := false
			switch n.edge {
			case "rising":
				shouldTrigger = edgeType == "rising"
			case "falling":
				shouldTrigger = edgeType == "falling"
			case "both":
				shouldTrigger = true
			}

			n.mu.Lock()
			n.lastState = currentState
			n.lastTrigger = time.Now()
			n.mu.Unlock()

			if shouldTrigger {
				count := atomic.AddInt64(&n.count, 1)
				msg := node.Message{
					Type: node.MessageTypeData,
					Payload: map[string]interface{}{
						"pin":         n.pin,
						"state":       currentState,
						"count":       count,
						"timestamp":   time.Now().Format(time.RFC3339Nano),
						"edge_type":   edgeType,
						"debounce_ms": n.debounceMs,
					},
				}
				select {
				case n.outputChan <- msg:
				default:
				}
			}
		}
	}
}

// Cleanup releases resources
func (n *InterruptNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.running {
		close(n.stopChan)
		n.running = false
	}
	return nil
}
