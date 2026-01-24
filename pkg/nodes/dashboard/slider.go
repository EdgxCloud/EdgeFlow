package dashboard

import (
	"context"

	"github.com/edgeflow/edgeflow/internal/node"
)

// SliderNode provides interactive slider input
type SliderNode struct {
	*BaseWidget
	min   float64
	max   float64
	step  float64
	value float64
	units string
}

// NewSliderNode creates a new slider widget
func NewSliderNode() *SliderNode {
	return &SliderNode{
		BaseWidget: NewBaseWidget(WidgetTypeSlider),
		min:        0,
		max:        100,
		step:       1,
		value:      50,
	}
}

// Init initializes the slider node
func (n *SliderNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if min, ok := config["min"].(float64); ok {
		n.min = min
	}
	if max, ok := config["max"].(float64); ok {
		n.max = max
	}
	if step, ok := config["step"].(float64); ok {
		n.step = step
	}
	if value, ok := config["value"].(float64); ok {
		n.value = value
	}
	if units, ok := config["units"].(string); ok {
		n.units = units
	}

	return nil
}

// Execute handles slider value changes
func (n *SliderNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract value from user interaction
	if value, ok := msg.Payload["value"].(float64); ok {
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
func (n *SliderNode) SetManager(manager *Manager) {
	n.manager = manager
}
