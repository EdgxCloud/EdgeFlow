package dashboard

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// GaugeNode displays numeric values as gauges
type GaugeNode struct {
	*BaseWidget
	min        float64
	max        float64
	units      string
	sectors    []GaugeSector
	showValue  bool
	showMinMax bool
}

// GaugeSector defines a colored sector on the gauge
type GaugeSector struct {
	From  float64 `json:"from"`
	To    float64 `json:"to"`
	Color string  `json:"color"`
	Label string  `json:"label,omitempty"`
}

// NewGaugeNode creates a new gauge widget
func NewGaugeNode() *GaugeNode {
	return &GaugeNode{
		BaseWidget: NewBaseWidget(WidgetTypeGauge),
		min:        0,
		max:        100,
		showValue:  true,
		showMinMax: true,
		sectors:    make([]GaugeSector, 0),
	}
}

// Init initializes the gauge node
func (n *GaugeNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	// Parse gauge-specific configuration
	if min, ok := config["min"].(float64); ok {
		n.min = min
	}
	if max, ok := config["max"].(float64); ok {
		n.max = max
	}
	if units, ok := config["units"].(string); ok {
		n.units = units
	}
	if showValue, ok := config["showValue"].(bool); ok {
		n.showValue = showValue
	}
	if showMinMax, ok := config["showMinMax"].(bool); ok {
		n.showMinMax = showMinMax
	}

	// Parse sectors
	if sectors, ok := config["sectors"].([]interface{}); ok {
		for _, s := range sectors {
			if sectorMap, ok := s.(map[string]interface{}); ok {
				sector := GaugeSector{}
				if from, ok := sectorMap["from"].(float64); ok {
					sector.From = from
				}
				if to, ok := sectorMap["to"].(float64); ok {
					sector.To = to
				}
				if color, ok := sectorMap["color"].(string); ok {
					sector.Color = color
				}
				if label, ok := sectorMap["label"].(string); ok {
					sector.Label = label
				}
				n.sectors = append(n.sectors, sector)
			}
		}
	}

	return nil
}

// Execute processes incoming messages and updates the gauge
func (n *GaugeNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract numeric value from payload
	var value float64

	if v, ok := msg.Payload["value"].(float64); ok {
		value = v
	} else if v, ok := msg.Payload["value"].(int); ok {
		value = float64(v)
	} else {
		return node.Message{}, fmt.Errorf("gauge requires numeric value")
	}

	// Clamp value to min/max
	if value < n.min {
		value = n.min
	}
	if value > n.max {
		value = n.max
	}

	// Update dashboard
	if n.manager != nil {
		gaugeData := map[string]interface{}{
			"value":      value,
			"min":        n.min,
			"max":        n.max,
			"units":      n.units,
			"sectors":    n.sectors,
			"showValue":  n.showValue,
			"showMinMax": n.showMinMax,
		}
		n.manager.UpdateWidget(n.id, gaugeData)
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *GaugeNode) SetManager(manager *Manager) {
	n.manager = manager
}
