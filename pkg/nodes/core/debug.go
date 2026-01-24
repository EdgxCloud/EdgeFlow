package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// DebugNode outputs messages for debugging
type DebugNode struct {
	outputTo   string // "console" or "log"
	complete   bool   // output complete message or just payload
	outputFunc func(string)
}

// NewDebugNode creates a new debug node
func NewDebugNode() *DebugNode {
	return &DebugNode{
		outputTo: "console",
		complete: false,
		outputFunc: func(s string) {
			fmt.Println(s)
		},
	}
}

// Init initializes the debug node with configuration
func (n *DebugNode) Init(config map[string]interface{}) error {
	// Parse output destination
	if outputTo, ok := config["output_to"].(string); ok {
		n.outputTo = outputTo
	}

	// Parse complete flag
	if complete, ok := config["complete"].(bool); ok {
		n.complete = complete
	}

	return nil
}

// Execute processes incoming messages and outputs them
func (n *DebugNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var output string

	if n.complete {
		// Output entire message
		msgJSON, err := json.MarshalIndent(msg, "", "  ")
		if err != nil {
			return msg, fmt.Errorf("failed to marshal message: %w", err)
		}
		output = string(msgJSON)
	} else {
		// Output just payload
		payloadJSON, err := json.MarshalIndent(msg.Payload, "", "  ")
		if err != nil {
			return msg, fmt.Errorf("failed to marshal payload: %w", err)
		}
		output = string(payloadJSON)
	}

	// Output based on configuration
	n.outputFunc(fmt.Sprintf("[DEBUG] %s", output))

	// Pass message through unchanged
	return msg, nil
}

// Cleanup stops the debug node
func (n *DebugNode) Cleanup() error {
	return nil
}

// SetOutputFunc allows custom output function (useful for testing)
func (n *DebugNode) SetOutputFunc(f func(string)) {
	n.outputFunc = f
}
