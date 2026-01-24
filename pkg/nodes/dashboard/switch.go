package dashboard

import (
	"context"

	"github.com/edgeflow/edgeflow/internal/node"
)

// SwitchNode provides interactive switch/toggle input
type SwitchNode struct {
	*BaseWidget
	onValue  interface{}
	offValue interface{}
	state    bool
	onLabel  string
	offLabel string
	onColor  string
	offColor string
}

// NewSwitchNode creates a new switch widget
func NewSwitchNode() *SwitchNode {
	return &SwitchNode{
		BaseWidget: NewBaseWidget(WidgetTypeSwitch),
		onValue:    true,
		offValue:   false,
		state:      false,
		onLabel:    "On",
		offLabel:   "Off",
		onColor:    "#10b981",
		offColor:   "#6b7280",
	}
}

// Init initializes the switch node
func (n *SwitchNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if onValue, ok := config["onValue"]; ok {
		n.onValue = onValue
	}
	if offValue, ok := config["offValue"]; ok {
		n.offValue = offValue
	}
	if state, ok := config["state"].(bool); ok {
		n.state = state
	}
	if onLabel, ok := config["onLabel"].(string); ok {
		n.onLabel = onLabel
	}
	if offLabel, ok := config["offLabel"].(string); ok {
		n.offLabel = offLabel
	}
	if onColor, ok := config["onColor"].(string); ok {
		n.onColor = onColor
	}
	if offColor, ok := config["offColor"].(string); ok {
		n.offColor = offColor
	}

	return nil
}

// Execute handles switch state changes
func (n *SwitchNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract state from user interaction
	if state, ok := msg.Payload["value"].(bool); ok {
		n.state = state

		// Determine output value
		var outputValue interface{}
		if state {
			outputValue = n.onValue
		} else {
			outputValue = n.offValue
		}

		// Send output message
		var outputPayload map[string]interface{}
		switch v := outputValue.(type) {
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
func (n *SwitchNode) SetManager(manager *Manager) {
	n.manager = manager
}
