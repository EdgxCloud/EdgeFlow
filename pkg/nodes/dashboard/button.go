package dashboard

import (
	"context"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// ButtonNode provides interactive buttons on the dashboard
type ButtonNode struct {
	*BaseWidget
	buttonLabel string
	payload     interface{}
	topic       string
	icon        string
	bgColor     string
	fgColor     string
}

// NewButtonNode creates a new button widget
func NewButtonNode() *ButtonNode {
	return &ButtonNode{
		BaseWidget:  NewBaseWidget(WidgetTypeButton),
		buttonLabel: "Button",
		payload:     map[string]interface{}{"clicked": true},
		bgColor:     "#3b82f6",
		fgColor:     "#ffffff",
	}
}

// Init initializes the button node
func (n *ButtonNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	// Parse button-specific configuration
	if buttonLabel, ok := config["buttonLabel"].(string); ok {
		n.buttonLabel = buttonLabel
	}
	if payload, ok := config["payload"]; ok {
		n.payload = payload
	}
	if topic, ok := config["topic"].(string); ok {
		n.topic = topic
	}
	if icon, ok := config["icon"].(string); ok {
		n.icon = icon
	}
	if bgColor, ok := config["bgColor"].(string); ok {
		n.bgColor = bgColor
	}
	if fgColor, ok := config["fgColor"].(string); ok {
		n.fgColor = fgColor
	}

	return nil
}

// Execute handles button clicks from the dashboard
func (n *ButtonNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Button node typically receives click events from dashboard
	// and sends configured payload to output

	var outputPayload map[string]interface{}

	// Use configured payload
	switch p := n.payload.(type) {
	case map[string]interface{}:
		outputPayload = p
	default:
		outputPayload = map[string]interface{}{"value": p}
	}

	outputMsg := node.Message{
		Type:    node.MessageTypeData,
		Topic:   n.topic,
		Payload: outputPayload,
	}

	// Send to output channel
	n.SendOutput(outputMsg)

	return node.Message{}, nil
}

// HandleClick handles button click events
func (n *ButtonNode) HandleClick() {
	var outputPayload map[string]interface{}

	switch p := n.payload.(type) {
	case map[string]interface{}:
		outputPayload = p
	default:
		outputPayload = map[string]interface{}{"value": p}
	}

	// Add timestamp
	outputPayload["timestamp"] = time.Now().Unix()

	outputMsg := node.Message{
		Type:    node.MessageTypeData,
		Topic:   n.topic,
		Payload: outputPayload,
	}

	n.SendOutput(outputMsg)
}

// SetManager sets the dashboard manager
func (n *ButtonNode) SetManager(manager *Manager) {
	n.manager = manager
}

// GetButtonConfig returns button configuration for dashboard
func (n *ButtonNode) GetButtonConfig() map[string]interface{} {
	return map[string]interface{}{
		"label":   n.buttonLabel,
		"icon":    n.icon,
		"bgColor": n.bgColor,
		"fgColor": n.fgColor,
	}
}
