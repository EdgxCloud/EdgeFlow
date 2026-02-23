package core

import (
	"context"
	"math"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRangeNode_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid scale config",
			config: map[string]interface{}{
				"action":   "scale",
				"property": "payload",
				"minin":    0.0,
				"maxin":    100.0,
				"minout":   0.0,
				"maxout":   1.0,
				"round":    false,
			},
			wantErr: false,
		},
		{
			name: "valid clamp config",
			config: map[string]interface{}{
				"action":   "clamp",
				"property": "payload.temperature",
				"minin":    -10.0,
				"maxin":    50.0,
				"minout":   0.0,
				"maxout":   100.0,
			},
			wantErr: false,
		},
		{
			name: "invalid range (minin == maxin)",
			config: map[string]interface{}{
				"action":   "scale",
				"minin":    10.0,
				"maxin":    10.0,
			},
			wantErr: true,
		},
		{
			name: "invalid action",
			config: map[string]interface{}{
				"action":   "invalid",
				"minin":    0.0,
				"maxin":    100.0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewRangeNode()
			err := n.Init(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRangeNode_Scale(t *testing.T) {
	tests := []struct {
		name     string
		config   map[string]interface{}
		input    float64
		expected float64
	}{
		{
			name: "0-100 to 0-1",
			config: map[string]interface{}{
				"action":   "scale",
				"property": "payload",
				"minin":    0.0,
				"maxin":    100.0,
				"minout":   0.0,
				"maxout":   1.0,
			},
			input:    50.0,
			expected: 0.5,
		},
		{
			name: "0-10 to 0-100",
			config: map[string]interface{}{
				"action":   "scale",
				"property": "payload",
				"minin":    0.0,
				"maxin":    10.0,
				"minout":   0.0,
				"maxout":   100.0,
			},
			input:    5.0,
			expected: 50.0,
		},
		{
			name: "-10 to 10 mapped to 0-100",
			config: map[string]interface{}{
				"action":   "scale",
				"property": "payload",
				"minin":    -10.0,
				"maxin":    10.0,
				"minout":   0.0,
				"maxout":   100.0,
			},
			input:    0.0,
			expected: 50.0,
		},
		{
			name: "inverted range (100-0 to 0-100)",
			config: map[string]interface{}{
				"action":   "scale",
				"property": "payload",
				"minin":    100.0,
				"maxin":    0.0,
				"minout":   0.0,
				"maxout":   100.0,
			},
			input:    50.0,
			expected: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewRangeNode()
			err := n.Init(tt.config)
			require.NoError(t, err)

			inputMsg := node.Message{
				Payload: tt.input,
			}

			resultMsg, err := n.Execute(context.Background(), inputMsg)
			require.NoError(t, err)

			output := resultMsg.Payload.(float64)
			assert.InDelta(t, tt.expected, output, 0.0001)
		})
	}
}

func TestRangeNode_Clamp(t *testing.T) {
	config := map[string]interface{}{
		"action":   "clamp",
		"property": "payload",
		"minin":    0.0,
		"maxin":    100.0,
		"minout":   0.0,
		"maxout":   10.0,
	}

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "within range",
			input:    50.0,
			expected: 5.0,
		},
		{
			name:     "below minimum",
			input:    -50.0,
			expected: 0.0, // Clamped to minout
		},
		{
			name:     "above maximum",
			input:    150.0,
			expected: 10.0, // Clamped to maxout
		},
		{
			name:     "at minimum",
			input:    0.0,
			expected: 0.0,
		},
		{
			name:     "at maximum",
			input:    100.0,
			expected: 10.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewRangeNode()
			err := n.Init(config)
			require.NoError(t, err)

			inputMsg := node.Message{
				Payload: tt.input,
			}

			resultMsg, err := n.Execute(context.Background(), inputMsg)
			require.NoError(t, err)

			output := resultMsg.Payload.(float64)
			assert.InDelta(t, tt.expected, output, 0.0001)
		})
	}
}

func TestRangeNode_Wrap(t *testing.T) {
	config := map[string]interface{}{
		"action":   "wrap",
		"property": "payload",
		"minin":    0.0,
		"maxin":    360.0,
		"minout":   0.0,
		"maxout":   10.0,
	}

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "within range",
			input:    180.0,
			expected: 5.0,
		},
		{
			name:     "wraps around once",
			input:    450.0, // 360 + 90
			expected: 2.5,
		},
		{
			name:     "wraps around multiple times",
			input:    900.0, // 360 * 2 + 180
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewRangeNode()
			err := n.Init(config)
			require.NoError(t, err)

			inputMsg := node.Message{
				Payload: tt.input,
			}

			resultMsg, err := n.Execute(context.Background(), inputMsg)
			require.NoError(t, err)

			output := resultMsg.Payload.(float64)
			assert.InDelta(t, tt.expected, output, 0.0001)
		})
	}
}

func TestRangeNode_Round(t *testing.T) {
	config := map[string]interface{}{
		"action":   "scale",
		"property": "payload",
		"minin":    0.0,
		"maxin":    100.0,
		"minout":   0.0,
		"maxout":   10.0,
		"round":    true,
	}

	tests := []struct {
		name     string
		input    float64
		expected float64
	}{
		{
			name:     "rounds down",
			input:    44.0, // Maps to 4.4 -> rounds to 4
			expected: 4.0,
		},
		{
			name:     "rounds up",
			input:    56.0, // Maps to 5.6 -> rounds to 6
			expected: 6.0,
		},
		{
			name:     "exact value",
			input:    50.0, // Maps to 5.0 -> stays 5
			expected: 5.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewRangeNode()
			err := n.Init(config)
			require.NoError(t, err)

			inputMsg := node.Message{
				Payload: tt.input,
			}

			resultMsg, err := n.Execute(context.Background(), inputMsg)
			require.NoError(t, err)

			output := resultMsg.Payload.(float64)
			assert.Equal(t, tt.expected, output)
		})
	}
}

func TestRangeNode_NestedProperty(t *testing.T) {
	config := map[string]interface{}{
		"action":   "scale",
		"property": "temperature",
		"minin":    0.0,
		"maxin":    100.0,
		"minout":   32.0,
		"maxout":   212.0, // Celsius to Fahrenheit
	}

	n := NewRangeNode()
	err := n.Init(config)
	require.NoError(t, err)

	inputMsg := node.Message{
		Payload: map[string]interface{}{
			"temperature": 0.0,
			"humidity":    60.0,
		},
	}

	resultMsg, err := n.Execute(context.Background(), inputMsg)
	require.NoError(t, err)

	payloadMap := resultMsg.Payload.(map[string]interface{})
	temp := payloadMap["temperature"].(float64)
	assert.InDelta(t, 32.0, temp, 0.0001)

	// Verify other properties are preserved
	assert.Equal(t, 60.0, payloadMap["humidity"])
}

func TestRangeNode_InvalidInput(t *testing.T) {
	config := map[string]interface{}{
		"action":   "scale",
		"property": "payload",
		"minin":    0.0,
		"maxin":    100.0,
		"minout":   0.0,
		"maxout":   10.0,
	}

	n := NewRangeNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		inputMsg node.Message
	}{
		{
			name: "string payload",
			inputMsg: node.Message{
				Payload: "not a number",
			},
		},
		{
			name: "boolean payload",
			inputMsg: node.Message{
				Payload: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := n.Execute(context.Background(), tt.inputMsg)
			assert.Error(t, err)
		})
	}
}

func TestRangeNode_Cleanup(t *testing.T) {
	n := NewRangeNode()
	err := n.Cleanup()
	assert.NoError(t, err)
}

func TestRangeToFloat64(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected float64
		ok       bool
	}{
		{"float64", 3.14, 3.14, true},
		{"float32", float32(2.71), 2.71, true},
		{"int", 42, 42.0, true},
		{"int32", int32(100), 100.0, true},
		{"int64", int64(200), 200.0, true},
		{"uint", uint(50), 50.0, true},
		{"string", "not a number", 0.0, false},
		{"bool", true, 0.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := toFloat64(tt.input)
			assert.Equal(t, tt.ok, ok)
			if ok {
				assert.InDelta(t, tt.expected, result, 0.0001)
			}
		})
	}
}
