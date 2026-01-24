package core

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/edgeflow/edgeflow/internal/node"
)

// StatisticsNode computes statistical measures
type StatisticsNode struct {
	windowSize int      // Size of sliding window (0 = process entire array)
	operations []string // Statistics to calculate
	property   string   // Property containing data

	// Sliding window state
	window []float64
	mu     sync.Mutex
}

// Init initializes the statistics node
func (n *StatisticsNode) Init(config map[string]interface{}) error {
	// Window size (0 = no window, operate on entire array)
	if ws, ok := config["windowSize"].(float64); ok {
		n.windowSize = int(ws)
	} else if ws, ok := config["windowSize"].(int); ok {
		n.windowSize = ws
	} else {
		n.windowSize = 0
	}

	// Operations to perform
	if ops, ok := config["operations"].([]interface{}); ok {
		n.operations = make([]string, len(ops))
		for i, op := range ops {
			n.operations[i] = fmt.Sprintf("%v", op)
		}
	} else if ops, ok := config["operations"].([]string); ok {
		n.operations = ops
	} else {
		n.operations = []string{"mean", "min", "max", "count"}
	}

	// Property containing data
	if prop, ok := config["property"].(string); ok {
		n.property = prop
	} else {
		n.property = "value"
	}

	// Initialize window
	if n.windowSize > 0 {
		n.window = make([]float64, 0, n.windowSize)
	}

	return nil
}

// Execute computes statistics
func (n *StatisticsNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get input data
	var data []float64

	if n.windowSize > 0 {
		// Sliding window mode - add new value to window
		n.mu.Lock()
		defer n.mu.Unlock()

		var newValue float64
		if val, ok := msg.Payload[n.property]; ok {
			newValue = n.toFloat64(val)
		} else if val, ok := msg.Payload["value"]; ok {
			newValue = n.toFloat64(val)
		} else {
			return msg, fmt.Errorf("property %s not found in payload", n.property)
		}

		// Add to window
		n.window = append(n.window, newValue)

		// Trim window if needed
		if len(n.window) > n.windowSize {
			n.window = n.window[len(n.window)-n.windowSize:]
		}

		data = make([]float64, len(n.window))
		copy(data, n.window)
	} else {
		// Array mode - operate on entire array
		if val, ok := msg.Payload[n.property]; ok {
			data = n.toFloat64Array(val)
		} else if val, ok := msg.Payload["value"]; ok {
			data = n.toFloat64Array(val)
		} else {
			return msg, fmt.Errorf("property %s not found in payload", n.property)
		}
	}

	if len(data) == 0 {
		return msg, fmt.Errorf("no data to process")
	}

	// Compute statistics
	stats := make(map[string]interface{})

	for _, op := range n.operations {
		switch op {
		case "count":
			stats["count"] = len(data)
		case "sum":
			stats["sum"] = n.sum(data)
		case "mean", "average", "avg":
			stats["mean"] = n.mean(data)
		case "median":
			stats["median"] = n.median(data)
		case "mode":
			stats["mode"] = n.mode(data)
		case "min":
			stats["min"] = n.min(data)
		case "max":
			stats["max"] = n.max(data)
		case "range":
			stats["range"] = n.max(data) - n.min(data)
		case "variance", "var":
			stats["variance"] = n.variance(data)
		case "stddev", "std":
			stats["stddev"] = n.stddev(data)
		case "p25", "q1":
			stats["p25"] = n.percentile(data, 25)
		case "p50":
			stats["p50"] = n.percentile(data, 50)
		case "p75", "q3":
			stats["p75"] = n.percentile(data, 75)
		case "p90":
			stats["p90"] = n.percentile(data, 90)
		case "p95":
			stats["p95"] = n.percentile(data, 95)
		case "p99":
			stats["p99"] = n.percentile(data, 99)
		case "iqr":
			stats["iqr"] = n.percentile(data, 75) - n.percentile(data, 25)
		case "skewness":
			stats["skewness"] = n.skewness(data)
		case "kurtosis":
			stats["kurtosis"] = n.kurtosis(data)
		}
	}

	msg.Payload["stats"] = stats
	msg.Payload["_statistics"] = map[string]interface{}{
		"dataPoints":  len(data),
		"windowSize":  n.windowSize,
		"windowMode":  n.windowSize > 0,
		"operations":  n.operations,
	}

	return msg, nil
}

// Statistical functions

func (n *StatisticsNode) sum(data []float64) float64 {
	var sum float64
	for _, v := range data {
		sum += v
	}
	return sum
}

func (n *StatisticsNode) mean(data []float64) float64 {
	return n.sum(data) / float64(len(data))
}

func (n *StatisticsNode) median(data []float64) float64 {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

func (n *StatisticsNode) mode(data []float64) float64 {
	counts := make(map[float64]int)
	for _, v := range data {
		counts[v]++
	}

	var mode float64
	maxCount := 0
	for v, count := range counts {
		if count > maxCount {
			maxCount = count
			mode = v
		}
	}
	return mode
}

func (n *StatisticsNode) min(data []float64) float64 {
	min := data[0]
	for _, v := range data[1:] {
		if v < min {
			min = v
		}
	}
	return min
}

func (n *StatisticsNode) max(data []float64) float64 {
	max := data[0]
	for _, v := range data[1:] {
		if v > max {
			max = v
		}
	}
	return max
}

func (n *StatisticsNode) variance(data []float64) float64 {
	mean := n.mean(data)
	var sumSquares float64
	for _, v := range data {
		diff := v - mean
		sumSquares += diff * diff
	}
	return sumSquares / float64(len(data))
}

func (n *StatisticsNode) stddev(data []float64) float64 {
	return math.Sqrt(n.variance(data))
}

func (n *StatisticsNode) percentile(data []float64, p float64) float64 {
	sorted := make([]float64, len(data))
	copy(sorted, data)
	sort.Float64s(sorted)

	rank := (p / 100) * float64(len(sorted)-1)
	lower := int(math.Floor(rank))
	upper := int(math.Ceil(rank))

	if lower == upper {
		return sorted[lower]
	}

	// Linear interpolation
	frac := rank - float64(lower)
	return sorted[lower]*(1-frac) + sorted[upper]*frac
}

func (n *StatisticsNode) skewness(data []float64) float64 {
	mean := n.mean(data)
	std := n.stddev(data)
	if std == 0 {
		return 0
	}

	var sum float64
	for _, v := range data {
		z := (v - mean) / std
		sum += z * z * z
	}
	return sum / float64(len(data))
}

func (n *StatisticsNode) kurtosis(data []float64) float64 {
	mean := n.mean(data)
	std := n.stddev(data)
	if std == 0 {
		return 0
	}

	var sum float64
	for _, v := range data {
		z := (v - mean) / std
		sum += z * z * z * z
	}
	return sum/float64(len(data)) - 3 // Excess kurtosis
}

// Conversion helpers

func (n *StatisticsNode) toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	default:
		return 0
	}
}

func (n *StatisticsNode) toFloat64Array(val interface{}) []float64 {
	switch v := val.(type) {
	case []float64:
		return v
	case []interface{}:
		result := make([]float64, len(v))
		for i, item := range v {
			result[i] = n.toFloat64(item)
		}
		return result
	case []int:
		result := make([]float64, len(v))
		for i, item := range v {
			result[i] = float64(item)
		}
		return result
	default:
		return []float64{}
	}
}

// Cleanup releases resources
func (n *StatisticsNode) Cleanup() error {
	return nil
}

// ResetWindow resets the sliding window
func (n *StatisticsNode) ResetWindow() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.window = make([]float64, 0, n.windowSize)
}

// NewStatisticsExecutor creates a new statistics node executor
func NewStatisticsExecutor() node.Executor {
	return &StatisticsNode{}
}

// init registers the statistics node
func init() {
	registry := node.GetGlobalRegistry()
	registry.Register(&node.NodeInfo{
		Type:        "statistics",
		Name:        "Statistics",
		Category:    node.NodeTypeFunction,
		Description: "Compute statistical measures (mean, median, stddev, percentiles, etc.)",
		Icon:        "chart-bar",
		Color:       "#4169E1",
		Properties: []node.PropertySchema{
			{
				Name:        "windowSize",
				Label:       "Window Size",
				Type:        "number",
				Default:     0,
				Required:    false,
				Description: "Sliding window size (0 = process entire array at once)",
			},
			{
				Name:        "operations",
				Label:       "Operations",
				Type:        "multiselect",
				Default:     []string{"mean", "min", "max", "count"},
				Required:    true,
				Description: "Statistics to calculate",
				Options:     []string{"count", "sum", "mean", "median", "mode", "min", "max", "range", "variance", "stddev", "p25", "p50", "p75", "p90", "p95", "p99", "iqr", "skewness", "kurtosis"},
			},
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "value",
				Required:    false,
				Description: "Property containing the numeric data (array or single value for sliding window)",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Numeric data (array or single value)"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Statistics in msg.payload.stats"},
		},
		Factory: NewStatisticsExecutor,
	})
}
