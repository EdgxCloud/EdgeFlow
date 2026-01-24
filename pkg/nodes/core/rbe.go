package core

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/edgeflow/edgeflow/internal/node"
)

// RBENode (Report By Exception) only passes messages when value changes
type RBENode struct {
	property   string      // Property to check for changes
	mode       string      // "value", "deadband", "narrowband"
	bandgap    float64     // Tolerance for numeric comparisons
	startValue interface{} // Initial value to compare against
	invert     bool        // Invert the logic (block changes, pass same)

	mu        sync.Mutex
	lastValue interface{}
	started   bool
}

// NewRBENode creates a new RBE node
func NewRBENode() *RBENode {
	return &RBENode{
		property:   "payload",
		mode:       "value",
		bandgap:    0,
		startValue: nil,
		invert:     false,
		started:    false,
	}
}

// Init initializes the RBE node with configuration
func (n *RBENode) Init(config map[string]interface{}) error {
	if property, ok := config["property"].(string); ok {
		n.property = property
	}
	if mode, ok := config["mode"].(string); ok {
		n.mode = mode
	}
	if bandgap, ok := config["bandgap"].(float64); ok {
		n.bandgap = bandgap
	}
	if startValue, ok := config["startValue"]; ok {
		n.startValue = startValue
		n.lastValue = startValue
		n.started = true
	}
	if invert, ok := config["invert"].(bool); ok {
		n.invert = invert
	}
	return nil
}

// Execute filters messages based on value changes
func (n *RBENode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Get the value to check
	currentValue := n.getValue(msg)

	// First message always passes (unless we have a startValue)
	if !n.started {
		n.lastValue = currentValue
		n.started = true
		if n.invert {
			// Invert mode: block first message
			return node.Message{}, nil
		}
		return msg, nil
	}

	// Check if value changed based on mode
	changed := n.hasChanged(currentValue)

	// Update last value
	n.lastValue = currentValue

	// Determine if we should pass the message
	shouldPass := changed
	if n.invert {
		shouldPass = !changed
	}

	if shouldPass {
		return msg, nil
	}

	// Return empty message to indicate filtering
	return node.Message{}, nil
}

// getValue extracts the value to check from the message
func (n *RBENode) getValue(msg node.Message) interface{} {
	if msg.Payload == nil {
		return nil
	}

	if n.property == "payload" {
		// Return entire payload or value field
		if val, ok := msg.Payload["value"]; ok {
			return val
		}
		return msg.Payload
	}

	// Get nested property
	return msg.Payload[n.property]
}

// hasChanged determines if the value has changed based on mode
func (n *RBENode) hasChanged(current interface{}) bool {
	switch n.mode {
	case "value":
		return !n.valuesEqual(current, n.lastValue)

	case "deadband":
		// Only report if change exceeds bandgap
		return n.exceedsBandgap(current, n.lastValue)

	case "narrowband":
		// Only report if change is within bandgap (opposite of deadband)
		return !n.exceedsBandgap(current, n.lastValue)

	default:
		return !n.valuesEqual(current, n.lastValue)
	}
}

// valuesEqual compares two values for equality
func (n *RBENode) valuesEqual(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Try numeric comparison
	aNum, aOk := n.toFloat64(a)
	bNum, bOk := n.toFloat64(b)
	if aOk && bOk {
		return aNum == bNum
	}

	// String comparison
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

// exceedsBandgap checks if the difference exceeds the bandgap threshold
func (n *RBENode) exceedsBandgap(current, last interface{}) bool {
	currentNum, currentOk := n.toFloat64(current)
	lastNum, lastOk := n.toFloat64(last)

	if !currentOk || !lastOk {
		// Non-numeric: any change exceeds bandgap
		return !n.valuesEqual(current, last)
	}

	diff := math.Abs(currentNum - lastNum)
	return diff > n.bandgap
}

// toFloat64 converts a value to float64
func (n *RBENode) toFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case int32:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint64:
		return float64(val), true
	case uint32:
		return float64(val), true
	default:
		return 0, false
	}
}

// Cleanup cleans up resources
func (n *RBENode) Cleanup() error {
	return nil
}
