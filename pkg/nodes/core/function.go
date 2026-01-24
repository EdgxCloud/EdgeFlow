package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// FunctionNode executes custom JavaScript code (simplified version)
// TODO: Integrate with goja for actual JavaScript execution
type FunctionNode struct {
	code       string
	outputKey  string
}

// NewFunctionNode creates a new function node
func NewFunctionNode() *FunctionNode {
	return &FunctionNode{
		outputKey: "result",
	}
}

// Init initializes the function node with configuration
func (n *FunctionNode) Init(config map[string]interface{}) error {
	if code, ok := config["code"].(string); ok {
		n.code = code
	}

	if outputKey, ok := config["output_key"].(string); ok {
		n.outputKey = outputKey
	}

	if n.code == "" {
		return fmt.Errorf("function code cannot be empty")
	}

	return nil
}

// Execute runs the function code (simplified implementation)
func (n *FunctionNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// This is a simplified implementation
	// In a complete version, this would use goja or another JavaScript engine

	// For now, we'll support some basic operations via JSON templates
	result, err := n.executeSimpleFunction(msg)
	if err != nil {
		return msg, fmt.Errorf("function execution failed: %w", err)
	}

	// Add result to payload
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}
	msg.Payload[n.outputKey] = result

	return msg, nil
}

// Cleanup stops the function node
func (n *FunctionNode) Cleanup() error {
	return nil
}

// executeSimpleFunction provides basic function execution
// TODO: Replace with actual JavaScript engine (goja)
func (n *FunctionNode) executeSimpleFunction(msg node.Message) (interface{}, error) {
	// Simple template-based execution
	// This is a placeholder for actual JS execution

	// Example: Support simple operations like:
	// - "return msg.payload.temperature * 1.8 + 32" (F to C)
	// - "return msg.payload.value + 10"

	// For now, just return the payload
	// In production, integrate with goja JavaScript engine

	return map[string]interface{}{
		"code_executed": true,
		"original_payload": msg.Payload,
		"note": "JavaScript execution requires goja integration",
	}, nil
}

// Helper function to safely get values from payload
func getPayloadValue(payload map[string]interface{}, key string) (interface{}, error) {
	if value, ok := payload[key]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("key %s not found in payload", key)
}

// Helper function to convert to JSON string
func toJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
