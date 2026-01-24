package dashboard

import (
	"context"

	"github.com/edgeflow/edgeflow/internal/node"
)

// TextInputNode provides text input field
type TextInputNode struct {
	*BaseWidget
	placeholder string
	mode        string // "text", "password", "email", "number"
	delay       int    // Delay in ms before sending output
	value       string
}

// NewTextInputNode creates a new text input widget
func NewTextInputNode() *TextInputNode {
	return &TextInputNode{
		BaseWidget:  NewBaseWidget(WidgetTypeTextInput),
		placeholder: "Enter text...",
		mode:        "text",
		delay:       300,
	}
}

// Init initializes the text input node
func (n *TextInputNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if placeholder, ok := config["placeholder"].(string); ok {
		n.placeholder = placeholder
	}
	if mode, ok := config["mode"].(string); ok {
		n.mode = mode
	}
	if delay, ok := config["delay"].(float64); ok {
		n.delay = int(delay)
	}

	return nil
}

// Execute handles text input changes
func (n *TextInputNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract value from user interaction
	if value, ok := msg.Payload["value"].(string); ok {
		n.value = value

		// Send output message
		outputMsg := node.Message{
			Type: node.MessageTypeData,
			Payload: map[string]interface{}{
				"value": value,
			},
		}
		n.SendOutput(outputMsg)
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *TextInputNode) SetManager(manager *Manager) {
	n.manager = manager
}
