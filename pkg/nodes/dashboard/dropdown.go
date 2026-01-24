package dashboard

import (
	"context"

	"github.com/edgeflow/edgeflow/internal/node"
)

// DropdownOption represents a dropdown option
type DropdownOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

// DropdownNode provides dropdown select input
type DropdownNode struct {
	*BaseWidget
	options       []DropdownOption
	selectedValue interface{}
	placeholder   string
	multiple      bool
}

// NewDropdownNode creates a new dropdown widget
func NewDropdownNode() *DropdownNode {
	return &DropdownNode{
		BaseWidget:  NewBaseWidget(WidgetTypeDropdown),
		options:     make([]DropdownOption, 0),
		placeholder: "Select an option...",
		multiple:    false,
	}
}

// Init initializes the dropdown node
func (n *DropdownNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if placeholder, ok := config["placeholder"].(string); ok {
		n.placeholder = placeholder
	}
	if multiple, ok := config["multiple"].(bool); ok {
		n.multiple = multiple
	}

	// Parse options
	if options, ok := config["options"].([]interface{}); ok {
		for _, opt := range options {
			if optMap, ok := opt.(map[string]interface{}); ok {
				option := DropdownOption{
					Label: optMap["label"].(string),
					Value: optMap["value"],
				}
				n.options = append(n.options, option)
			}
		}
	}

	return nil
}

// Execute handles dropdown selection changes
func (n *DropdownNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract selected value from user interaction
	if value, ok := msg.Payload["value"]; ok {
		n.selectedValue = value

		// Send output message
		var outputPayload map[string]interface{}
		switch v := value.(type) {
		case map[string]interface{}:
			outputPayload = v
		default:
			outputPayload = map[string]interface{}{"value": v}
		}

		outputMsg := node.Message{
			Type:    node.MessageTypeData,
			Payload: outputPayload,
		}
		n.SendOutput(outputMsg)
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *DropdownNode) SetManager(manager *Manager) {
	n.manager = manager
}
