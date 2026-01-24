package dashboard

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// ChartType defines the type of chart
type ChartType string

const (
	ChartTypeLine      ChartType = "line"
	ChartTypeBar       ChartType = "bar"
	ChartTypePie       ChartType = "pie"
	ChartTypeHistogram ChartType = "histogram"
	ChartTypeScatter   ChartType = "scatter"
)

// ChartNode displays data as charts
type ChartNode struct {
	*BaseWidget
	chartType   ChartType
	xAxisLabel  string
	yAxisLabel  string
	maxDataSize int
	legend      bool
	dataPoints  []ChartDataPoint
}

// ChartDataPoint represents a single data point
type ChartDataPoint struct {
	X     interface{} `json:"x"`
	Y     interface{} `json:"y"`
	Label string      `json:"label,omitempty"`
}

// NewChartNode creates a new chart widget
func NewChartNode() *ChartNode {
	return &ChartNode{
		BaseWidget:  NewBaseWidget(WidgetTypeChart),
		chartType:   ChartTypeLine,
		maxDataSize: 100,
		legend:      true,
		dataPoints:  make([]ChartDataPoint, 0),
	}
}

// Init initializes the chart node
func (n *ChartNode) Init(config map[string]interface{}) error {
	if err := n.BaseWidget.Init(config); err != nil {
		return err
	}

	// Parse chart-specific configuration
	if chartType, ok := config["chartType"].(string); ok {
		n.chartType = ChartType(chartType)
	}
	if xAxisLabel, ok := config["xAxisLabel"].(string); ok {
		n.xAxisLabel = xAxisLabel
	}
	if yAxisLabel, ok := config["yAxisLabel"].(string); ok {
		n.yAxisLabel = yAxisLabel
	}
	if maxDataSize, ok := config["maxDataSize"].(float64); ok {
		n.maxDataSize = int(maxDataSize)
	}
	if legend, ok := config["legend"].(bool); ok {
		n.legend = legend
	}

	return nil
}

// Execute processes incoming messages and updates the chart
func (n *ChartNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Extract data from payload
	var dataPoint ChartDataPoint

	if payload, ok := msg.Payload["value"]; ok {
		// Single value - use timestamp as X
		dataPoint = ChartDataPoint{
			Y: payload,
		}
	} else if x, xOk := msg.Payload["x"]; xOk {
		// X,Y pair provided
		dataPoint = ChartDataPoint{
			X: x,
			Y: msg.Payload["y"],
		}
		if label, labelOk := msg.Payload["label"].(string); labelOk {
			dataPoint.Label = label
		}
	} else {
		return node.Message{}, fmt.Errorf("invalid chart data format")
	}

	// Add data point
	n.dataPoints = append(n.dataPoints, dataPoint)

	// Limit data points to maxDataSize
	if len(n.dataPoints) > n.maxDataSize {
		n.dataPoints = n.dataPoints[len(n.dataPoints)-n.maxDataSize:]
	}

	// Update dashboard
	if n.manager != nil {
		chartData := map[string]interface{}{
			"type":       n.chartType,
			"dataPoints": n.dataPoints,
			"xAxisLabel": n.xAxisLabel,
			"yAxisLabel": n.yAxisLabel,
			"legend":     n.legend,
		}
		n.manager.UpdateWidget(n.id, chartData)
	}

	return node.Message{}, nil
}

// SetManager sets the dashboard manager
func (n *ChartNode) SetManager(manager *Manager) {
	n.manager = manager
}
