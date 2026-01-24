package dashboard

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// TextNode displays text on the dashboard
type TextNode struct {
	*BaseWidget
	format    string
	fontSize  int
	fontColor string
	layout    string
}

// NewTextNode creates a new text display widget
func NewTextNode() *TextNode {
	return &TextNode{
		BaseWidget: NewBaseWidget(WidgetTypeText),
		format:     "{{msg.payload}}",
		fontSize:   14,
		fontColor:  "#000000",
		layout:     "row-spread",
	}
}

// Init initializes the text node
func (n *TextNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	// Parse text-specific configuration
	if format, ok := config["format"].(string); ok {
		n.format = format
	}
	if fontSize, ok := config["fontSize"].(float64); ok {
		n.fontSize = int(fontSize)
	}
	if fontColor, ok := config["fontColor"].(string); ok {
		n.fontColor = fontColor
	}
	if layout, ok := config["layout"].(string); ok {
		n.layout = layout
	}

	return nil
}

// Execute processes incoming messages and updates the text display
func (n *TextNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract text from payload
	var text string

	if value, ok := msg.Payload["value"]; ok {
		text = fmt.Sprintf("%v", value)
	} else if textVal, ok := msg.Payload["text"].(string); ok {
		text = textVal
	} else {
		// Convert entire payload to string
		text = fmt.Sprintf("%v", msg.Payload)
	}

	// Update dashboard
	if n.manager != nil {
		textData := map[string]interface{}{
			"text":      text,
			"format":    n.format,
			"fontSize":  n.fontSize,
			"fontColor": n.fontColor,
			"layout":    n.layout,
		}
		n.manager.UpdateWidget(n.id, textData)
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *TextNode) SetManager(manager *Manager) {
	n.manager = manager
}
