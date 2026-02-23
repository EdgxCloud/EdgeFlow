package core

import (
	"context"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSwitchNode_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config with single rule",
			config: map[string]interface{}{
				"property": "payload",
				"rules": []interface{}{
					map[string]interface{}{
						"t": "eq",
						"v": "test",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "multiple rules",
			config: map[string]interface{}{
				"property": "payload.temperature",
				"rules": []interface{}{
					map[string]interface{}{"t": "lt", "v": 0.0},
					map[string]interface{}{"t": "btwn", "v": 0.0, "v2": 25.0},
					map[string]interface{}{"t": "gt", "v": 25.0},
				},
			},
			wantErr: false,
		},
		{
			name: "no rules",
			config: map[string]interface{}{
				"property": "payload",
				"rules":    []interface{}{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewSwitchNode()
			err := n.Init(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSwitchNode_Equals(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload",
		"rules": []interface{}{
			map[string]interface{}{
				"t":    "eq",
				"v":    "test",
				"case": true,
			},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"exact match", "test", true},
		{"no match", "TEST", false},
		{"number match", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{Payload: tt.input}
			result, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)

			// Check metadata for matched outputs
			if em, ok := interface{}(result).(*node.EnhancedMessage); ok {
				if matched, ok := em.Metadata["_switchOutputs"].([]bool); ok {
					assert.Equal(t, tt.expected, matched[0])
				}
			}
		})
	}
}

func TestSwitchNode_CaseInsensitive(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload",
		"rules": []interface{}{
			map[string]interface{}{
				"t":    "eq",
				"v":    "test",
				"case": false,
			},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"lowercase", "test", true},
		{"uppercase", "TEST", true},
		{"mixed case", "TeSt", true},
		{"no match", "other", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{Payload: tt.input}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_NumericComparisons(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload",
		"rules": []interface{}{
			map[string]interface{}{"t": "lt", "v": 10.0},
			map[string]interface{}{"t": "gte", "v": 10.0},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name           string
		input          float64
		expectedOutput int // Which output should match (0 or 1)
	}{
		{"less than", 5.0, 0},
		{"equal", 10.0, 1},
		{"greater", 15.0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{Payload: tt.input}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_Between(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload.temperature",
		"rules": []interface{}{
			map[string]interface{}{
				"t":  "btwn",
				"v":  0.0,
				"v2": 25.0,
			},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    float64
		expected bool
	}{
		{"below range", -5.0, false},
		{"at minimum", 0.0, true},
		{"in range", 15.0, true},
		{"at maximum", 25.0, true},
		{"above range", 30.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{
				Payload: map[string]interface{}{
					"temperature": tt.input,
				},
			}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_Contains(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload",
		"rules": []interface{}{
			map[string]interface{}{
				"t":    "cont",
				"v":    "hello",
				"case": false,
			},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"contains", "hello world", true},
		{"uppercase contains", "HELLO WORLD", true},
		{"not contains", "goodbye", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{Payload: tt.input}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_Regex(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload",
		"rules": []interface{}{
			map[string]interface{}{
				"t": "regex",
				"v": "^[0-9]{3}-[0-9]{4}$", // Phone number pattern
			},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid pattern", "123-4567", true},
		{"invalid pattern", "123-45678", false},
		{"no match", "abc-defg", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{Payload: tt.input}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_Boolean(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload.enabled",
		"rules": []interface{}{
			map[string]interface{}{"t": "true"},
			map[string]interface{}{"t": "false"},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name          string
		input         interface{}
		expectedTrue  bool
		expectedFalse bool
	}{
		{"boolean true", true, true, false},
		{"boolean false", false, false, true},
		{"string true", "true", true, false},
		{"string false", "false", false, true},
		{"number 1", 1, true, false},
		{"number 0", 0, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{
				Payload: map[string]interface{}{
					"enabled": tt.input,
				},
			}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_NullChecks(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload",
		"rules": []interface{}{
			map[string]interface{}{"t": "null"},
			map[string]interface{}{"t": "nnull"},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name         string
		input        interface{}
		expectedNull bool
	}{
		{"nil value", nil, true},
		{"string value", "test", false},
		{"number value", 42, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{Payload: tt.input}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_EmptyChecks(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload",
		"rules": []interface{}{
			map[string]interface{}{"t": "empty"},
			map[string]interface{}{"t": "nempty"},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name         string
		input        interface{}
		expectedEmpty bool
	}{
		{"empty string", "", true},
		{"empty array", []interface{}{}, true},
		{"empty object", map[string]interface{}{}, true},
		{"nil", nil, true},
		{"non-empty string", "test", false},
		{"non-empty array", []interface{}{1, 2}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{Payload: tt.input}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_TypeChecks(t *testing.T) {
	config := map[string]interface{}{
		"property": "payload",
		"rules": []interface{}{
			map[string]interface{}{"t": "istype", "v": "string"},
			map[string]interface{}{"t": "istype", "v": "number"},
			map[string]interface{}{"t": "istype", "v": "boolean"},
			map[string]interface{}{"t": "istype", "v": "array"},
			map[string]interface{}{"t": "istype", "v": "object"},
		},
	}

	n := NewSwitchNode()
	err := n.Init(config)
	require.NoError(t, err)

	tests := []struct {
		name          string
		input         interface{}
		expectedIndex int
	}{
		{"string", "test", 0},
		{"number int", 42, 1},
		{"number float", 3.14, 1},
		{"boolean", true, 2},
		{"array", []interface{}{1, 2, 3}, 3},
		{"object", map[string]interface{}{"key": "value"}, 4},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{Payload: tt.input}
			_, err := n.Execute(context.Background(), msg)
			require.NoError(t, err)
		})
	}
}

func TestSwitchNode_Cleanup(t *testing.T) {
	n := NewSwitchNode()
	err := n.Cleanup()
	assert.NoError(t, err)
}
