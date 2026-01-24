package dashboard

import (
	"context"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// DatePickerNode provides date/time picker input
type DatePickerNode struct {
	*BaseWidget
	mode          string // "date", "time", "datetime"
	format        string
	minDate       *time.Time
	maxDate       *time.Time
	selectedDate  *time.Time
	enableTime    bool
	enableSeconds bool
}

// NewDatePickerNode creates a new date picker widget
func NewDatePickerNode() *DatePickerNode {
	return &DatePickerNode{
		BaseWidget:    NewBaseWidget(WidgetTypeDatePicker),
		mode:          "datetime",
		format:        "2006-01-02 15:04:05",
		enableTime:    true,
		enableSeconds: false,
	}
}

// Init initializes the date picker node
func (n *DatePickerNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	if mode, ok := config["mode"].(string); ok {
		n.mode = mode
	}
	if format, ok := config["format"].(string); ok {
		n.format = format
	}
	if enableTime, ok := config["enableTime"].(bool); ok {
		n.enableTime = enableTime
	}
	if enableSeconds, ok := config["enableSeconds"].(bool); ok {
		n.enableSeconds = enableSeconds
	}

	return nil
}

// Execute handles date selection changes
func (n *DatePickerNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract date from user interaction
	if dateStr, ok := msg.Payload["value"].(string); ok {
		if parsedDate, err := time.Parse(n.format, dateStr); err == nil {
			n.selectedDate = &parsedDate

			// Send output message
			outputMsg := node.Message{
				Type: node.MessageTypeData,
				Payload: map[string]interface{}{
					"value":     dateStr,
					"timestamp": parsedDate.Unix(),
					"date":      parsedDate.Format("2006-01-02"),
					"time":      parsedDate.Format("15:04:05"),
				},
			}
			n.SendOutput(outputMsg)
		}
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *DatePickerNode) SetManager(manager *Manager) {
	n.manager = manager
}
