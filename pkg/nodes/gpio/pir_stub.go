//go:build !linux
// +build !linux

package gpio

import (
	"context"
	"fmt"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// PIRNode handles PIR motion sensor input (stub for non-Linux platforms)
type PIRNode struct {
	pinNumber     int
	debounce      time.Duration
	sensitivity   string
	pullMode      string
	triggerMode   string
	retriggerTime time.Duration
}

// NewPIRNode creates a new PIR sensor node
func NewPIRNode() *PIRNode {
	return &PIRNode{
		debounce:      50 * time.Millisecond,
		sensitivity:   "normal",
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

	return nil
}

// Execute processes incoming messages (stub - returns error on non-Linux)
func (n *PIRNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	return msg, fmt.Errorf("PIR sensor is only supported on Linux platforms")
}

// Cleanup stops monitoring and closes GPIO
func (n *PIRNode) Cleanup() error {
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
