package network

import (
	"context"
	"testing"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestYAMLParser_Parse tests YAML to JSON conversion
func TestYAMLParser_Parse(t *testing.T) {
	executor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    string
		wantKeys []string
		wantErr  bool
	}{
		{
			name: "simple YAML",
			input: `name: Test
value: 123
enabled: true`,
			wantKeys: []string{"name", "value", "enabled"},
			wantErr:  false,
		},
		{
			name: "nested YAML",
			input: `person:
  name: John
  age: 30
  address:
    city: NYC
    zip: 10001`,
			wantKeys: []string{"person"},
			wantErr:  false,
		},
		{
			name: "YAML array",
			input: `items:
  - one
  - two
  - three`,
			wantKeys: []string{"items"},
			wantErr:  false,
		},
		{
			name: "YAML with anchors",
			input: `defaults: &defaults
  timeout: 30
  retries: 3

service1:
  <<: *defaults
  name: Service 1`,
			wantKeys: []string{"defaults", "service1"},
			wantErr:  false,
		},
		{
			name:     "invalid YAML",
			input:    `invalid:\n  - unclosed`,
			wantKeys: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{
				Payload: tt.input,
			}

			result, err := executor.Execute(context.Background(), msg)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			payload, ok := result.Payload.(map[string]interface{})
			require.True(t, ok)

			// Check that expected keys are present
			for _, key := range tt.wantKeys {
				assert.Contains(t, payload, key)
			}
		})
	}
}

// TestYAMLParser_Stringify tests JSON to YAML conversion
func TestYAMLParser_Stringify(t *testing.T) {
	executor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "stringify",
	})
	require.NoError(t, err)

	tests := []struct {
		name     string
		input    interface{}
		contains []string
		wantErr  bool
	}{
		{
			name: "simple object",
			input: map[string]interface{}{
				"name":  "Test",
				"value": 123,
			},
			contains: []string{"name: Test", "value: 123"},
			wantErr:  false,
		},
		{
			name: "nested object",
			input: map[string]interface{}{
				"person": map[string]interface{}{
					"name": "John",
					"age":  30,
				},
			},
			contains: []string{"person:", "name: John", "age: 30"},
			wantErr:  false,
		},
		{
			name: "array",
			input: map[string]interface{}{
				"items": []interface{}{"one", "two", "three"},
			},
			contains: []string{"items:", "- one", "- two", "- three"},
			wantErr:  false,
		},
		{
			name: "boolean values",
			input: map[string]interface{}{
				"enabled":  true,
				"disabled": false,
			},
			contains: []string{"enabled: true", "disabled: false"},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := node.Message{
				Payload: tt.input,
			}

			result, err := executor.Execute(context.Background(), msg)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)

			yamlStr, ok := result.Payload.(string)
			require.True(t, ok)

			// Check that expected strings are in the YAML
			for _, substr := range tt.contains {
				assert.Contains(t, yamlStr, substr)
			}
		})
	}
}

// TestYAMLParser_ParseArray tests parsing YAML arrays
func TestYAMLParser_ParseArray(t *testing.T) {
	executor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `- name: Item 1
  value: 100
- name: Item 2
  value: 200`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.([]interface{})
	require.True(t, ok)
	assert.Len(t, payload, 2)

	item1, ok := payload[0].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Item 1", item1["name"])
	assert.Equal(t, float64(100), item1["value"])
}

// TestYAMLParser_MultilineStrings tests multiline string handling
func TestYAMLParser_MultilineStrings(t *testing.T) {
	executor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `description: |
  This is a
  multiline
  string`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)

	description, ok := payload["description"].(string)
	require.True(t, ok)
	assert.Contains(t, description, "multiline")
}

// TestYAMLParser_InvalidConfig tests invalid configuration
func TestYAMLParser_InvalidConfig(t *testing.T) {
	_, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "invalid",
	})
	assert.Error(t, err)
}

// TestYAMLParser_MissingAction tests missing action
func TestYAMLParser_MissingAction(t *testing.T) {
	_, err := NewYAMLParserExecutor(map[string]interface{}{})
	assert.Error(t, err)
}

// TestYAMLParser_NumberTypes tests different number types
func TestYAMLParser_NumberTypes(t *testing.T) {
	executor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `integer: 123
float: 123.45
scientific: 1.23e+10
hex: 0x1A
octal: 0o17`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, payload, "integer")
	assert.Contains(t, payload, "float")
}

// TestYAMLParser_NullValues tests null value handling
func TestYAMLParser_NullValues(t *testing.T) {
	executor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `value: null
another: ~
empty:`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)

	assert.Contains(t, payload, "value")
	assert.Nil(t, payload["value"])
}

// TestYAMLParser_MixedTypes tests mixed data types
func TestYAMLParser_MixedTypes(t *testing.T) {
	executor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	msg := node.Message{
		Payload: `config:
  string: "text"
  number: 42
  boolean: true
  null_value: null
  array:
    - 1
    - 2
    - 3
  object:
    key: value`,
	}

	result, err := executor.Execute(context.Background(), msg)
	require.NoError(t, err)

	payload, ok := result.Payload.(map[string]interface{})
	require.True(t, ok)

	config, ok := payload["config"].(map[string]interface{})
	require.True(t, ok)

	assert.Equal(t, "text", config["string"])
	assert.Equal(t, float64(42), config["number"])
	assert.Equal(t, true, config["boolean"])
	assert.Nil(t, config["null_value"])
}

// TestYAMLParser_RoundTrip tests parse -> stringify round trip
func TestYAMLParser_RoundTrip(t *testing.T) {
	parseExecutor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	stringifyExecutor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "stringify",
	})
	require.NoError(t, err)

	original := `name: Test
value: 123`

	// Parse
	parseMsg := node.Message{Payload: original}
	parseResult, err := parseExecutor.Execute(context.Background(), parseMsg)
	require.NoError(t, err)

	// Stringify
	stringifyResult, err := stringifyExecutor.Execute(context.Background(), parseResult)
	require.NoError(t, err)

	yamlStr, ok := stringifyResult.Payload.(string)
	require.True(t, ok)
	assert.Contains(t, yamlStr, "name:")
	assert.Contains(t, yamlStr, "value:")
}

// TestYAMLParser_Cleanup tests cleanup
func TestYAMLParser_Cleanup(t *testing.T) {
	executor, err := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})
	require.NoError(t, err)

	err = executor.Cleanup()
	assert.NoError(t, err)
}

// BenchmarkYAMLParser_Parse benchmarks YAML parsing
func BenchmarkYAMLParser_Parse(b *testing.B) {
	executor, _ := NewYAMLParserExecutor(map[string]interface{}{
		"action": "parse",
	})

	yaml := `person:
  name: John
  age: 30
  address:
    city: NYC
    zip: 10001
  hobbies:
    - reading
    - gaming
    - coding`

	msg := node.Message{Payload: yaml}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute(context.Background(), msg)
	}
}

// BenchmarkYAMLParser_Stringify benchmarks YAML stringification
func BenchmarkYAMLParser_Stringify(b *testing.B) {
	executor, _ := NewYAMLParserExecutor(map[string]interface{}{
		"action": "stringify",
	})

	data := map[string]interface{}{
		"person": map[string]interface{}{
			"name": "John",
			"age":  30,
			"address": map[string]interface{}{
				"city": "NYC",
				"zip":  10001,
			},
			"hobbies": []interface{}{"reading", "gaming", "coding"},
		},
	}

	msg := node.Message{Payload: data}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		executor.Execute(context.Background(), msg)
	}
}
