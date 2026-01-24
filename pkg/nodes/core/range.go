package core

import (
	"context"
	"fmt"
	"math"

	"github.com/edgeflow/edgeflow/internal/node"
)

// RangeNode scales numeric values between ranges
type RangeNode struct {
	action string  // "scale", "clamp", "wrap"
	minIn  float64 // Minimum input value
	maxIn  float64 // Maximum input value
	minOut float64 // Minimum output value
	maxOut float64 // Maximum output value
}

// NewRangeNode creates a new range node
func NewRangeNode() *RangeNode {
	return &RangeNode{
		action: "scale",
		minIn:  0,
		maxIn:  100,
		minOut: 0,
		maxOut: 1,
	}
}

// Init initializes the range node with configuration
func (n *RangeNode) Init(config map[string]interface{}) error {
	if action, ok := config["action"].(string); ok {
		n.action = action
	}
	if minIn, ok := config["minIn"].(float64); ok {
		n.minIn = minIn
	}
	if maxIn, ok := config["maxIn"].(float64); ok {
		n.maxIn = maxIn
	}
	if minOut, ok := config["minOut"].(float64); ok {
		n.minOut = minOut
	}
	if maxOut, ok := config["maxOut"].(float64); ok {
		n.maxOut = maxOut
	}
	return nil
}

// Execute applies the range transformation to the message
func (n *RangeNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract numeric value from payload
	value, err := n.extractNumber(msg.Payload)
	if err != nil {
		return msg, fmt.Errorf("range node: %w", err)
	}

	var result float64

	switch n.action {
	case "scale":
		result = n.scale(value)
	case "clamp":
		result = n.clamp(value)
	case "wrap":
		result = n.wrap(value)
	default:
		result = n.scale(value)
	}

	msg.Payload = map[string]interface{}{"value": result}
	return msg, nil
}

// extractNumber extracts a numeric value from various payload types
func (n *RangeNode) extractNumber(payload map[string]interface{}) (float64, error) {
	if payload == nil {
		return 0, fmt.Errorf("nil payload")
	}

	// Try to get "value" field first
	if val, ok := payload["value"]; ok {
		return n.toFloat64(val)
	}

	// Try to find any numeric value
	for _, v := range payload {
		if f, err := n.toFloat64(v); err == nil {
			return f, nil
		}
	}

	return 0, fmt.Errorf("no numeric value found in payload")
}

// toFloat64 converts various types to float64
func (n *RangeNode) toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case float32:
		return float64(val), nil
	case int:
		return float64(val), nil
	case int64:
		return float64(val), nil
	case int32:
		return float64(val), nil
	case uint:
		return float64(val), nil
	case uint64:
		return float64(val), nil
	case uint32:
		return float64(val), nil
	case string:
		var f float64
		_, err := fmt.Sscanf(val, "%f", &f)
		if err != nil {
			return 0, fmt.Errorf("cannot parse string '%s' as number", val)
		}
		return f, nil
	default:
		return 0, fmt.Errorf("unsupported type: %T", v)
	}
}

// scale maps a value from input range to output range
func (n *RangeNode) scale(value float64) float64 {
	// Handle edge case where input range is zero
	if n.maxIn == n.minIn {
		return n.minOut
	}

	// Normalize to 0-1 range
	normalized := (value - n.minIn) / (n.maxIn - n.minIn)

	// Scale to output range
	result := normalized*(n.maxOut-n.minOut) + n.minOut

	return result
}

// clamp constrains a value within the output range
func (n *RangeNode) clamp(value float64) float64 {
	// First scale the value
	scaled := n.scale(value)

	// Then clamp to output range
	minVal := math.Min(n.minOut, n.maxOut)
	maxVal := math.Max(n.minOut, n.maxOut)

	if scaled < minVal {
		return minVal
	}
	if scaled > maxVal {
		return maxVal
	}
	return scaled
}

// wrap wraps a value around within the output range
func (n *RangeNode) wrap(value float64) float64 {
	// First scale the value
	scaled := n.scale(value)

	// Get the range size
	minVal := math.Min(n.minOut, n.maxOut)
	maxVal := math.Max(n.minOut, n.maxOut)
	rangeSize := maxVal - minVal

	if rangeSize == 0 {
		return minVal
	}

	// Wrap the value
	wrapped := scaled - minVal
	wrapped = math.Mod(wrapped, rangeSize)
	if wrapped < 0 {
		wrapped += rangeSize
	}

	return wrapped + minVal
}

// Cleanup cleans up resources
func (n *RangeNode) Cleanup() error {
	return nil
}
