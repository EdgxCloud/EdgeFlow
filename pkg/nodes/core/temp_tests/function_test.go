package core

import (
	"context"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFunctionNode(t *testing.T) {
	n := NewFunctionNode()
	require.NotNil(t, n)
	assert.Equal(t, "result", n.outputKey)
}

func TestFunctionNode_Init(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		wantErr     bool
		errContains string
		checkResult func(*testing.T, *FunctionNode)
	}{
		{
			name: "valid code",
			config: map[string]interface{}{
				"code": "return msg.payload.value * 2",
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *FunctionNode) {
				assert.Equal(t, "return msg.payload.value * 2", n.code)
			},
		},
		{
			name: "custom output key",
			config: map[string]interface{}{
				"code":       "return msg.payload",
				"output_key": "transformed",
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *FunctionNode) {
				assert.Equal(t, "transformed", n.outputKey)
			},
		},
		{
			name: "empty code",
			config: map[string]interface{}{
				"code": "",
			},
			wantErr:     true,
			errContains: "cannot be empty",
		},
		{
			name:        "missing code",
			config:      map[string]interface{}{},
			wantErr:     true,
			errContains: "cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewFunctionNode()
			err := n.Init(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, n)
			}
		})
	}
}

func TestFunctionNode_Execute(t *testing.T) {
	n := NewFunctionNode()
	err := n.Init(map[string]interface{}{
		"code": "return msg.payload",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"temperature": 25.5,
			"humidity":    60,
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Check that result key exists
	assert.NotNil(t, result.Payload["result"])

	// The simplified implementation returns metadata about execution
	resultData := result.Payload["result"].(map[string]interface{})
	assert.True(t, resultData["code_executed"].(bool))
	assert.NotNil(t, resultData["original_payload"])
}

func TestFunctionNode_Execute_CustomOutputKey(t *testing.T) {
	n := NewFunctionNode()
	err := n.Init(map[string]interface{}{
		"code":       "return msg.payload.value + 10",
		"output_key": "calculated",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"value": 100,
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Check that custom output key is used
	assert.NotNil(t, result.Payload["calculated"])
	assert.Nil(t, result.Payload["result"]) // Default key should not exist
}

func TestFunctionNode_Execute_NilPayload(t *testing.T) {
	n := NewFunctionNode()
	err := n.Init(map[string]interface{}{
		"code": "return true",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: nil,
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Payload should be created with result
	assert.NotNil(t, result.Payload)
	assert.NotNil(t, result.Payload["result"])
}

func TestFunctionNode_Execute_PreservesExistingPayload(t *testing.T) {
	n := NewFunctionNode()
	err := n.Init(map[string]interface{}{
		"code": "return 42",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"existing_key": "existing_value",
			"number":       123,
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Existing payload keys should be preserved
	assert.Equal(t, "existing_value", result.Payload["existing_key"])
	assert.Equal(t, 123, result.Payload["number"])
	// And result should be added
	assert.NotNil(t, result.Payload["result"])
}

func TestFunctionNode_Cleanup(t *testing.T) {
	n := NewFunctionNode()
	err := n.Init(map[string]interface{}{
		"code": "return 1",
	})
	require.NoError(t, err)

	err = n.Cleanup()
	assert.NoError(t, err)
}

func TestFunctionNode_ExecuteSimpleFunction(t *testing.T) {
	n := NewFunctionNode()
	n.code = "return msg.payload"

	msg := node.Message{
		Payload: map[string]interface{}{
			"data": "test",
		},
	}

	result, err := n.executeSimpleFunction(msg)
	require.NoError(t, err)

	resultMap := result.(map[string]interface{})
	assert.True(t, resultMap["code_executed"].(bool))
	assert.NotNil(t, resultMap["original_payload"])
}

func TestGetPayloadValue(t *testing.T) {
	payload := map[string]interface{}{
		"key1": "value1",
		"key2": 123,
		"key3": true,
	}

	// Test existing keys
	val, err := getPayloadValue(payload, "key1")
	require.NoError(t, err)
	assert.Equal(t, "value1", val)

	val, err = getPayloadValue(payload, "key2")
	require.NoError(t, err)
	assert.Equal(t, 123, val)

	val, err = getPayloadValue(payload, "key3")
	require.NoError(t, err)
	assert.Equal(t, true, val)

	// Test missing key
	_, err = getPayloadValue(payload, "missing")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestToJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "simple string",
			input:   "test",
			want:    `"test"`,
			wantErr: false,
		},
		{
			name:    "number",
			input:   42,
			want:    "42",
			wantErr: false,
		},
		{
			name:    "boolean",
			input:   true,
			want:    "true",
			wantErr: false,
		},
		{
			name: "map",
			input: map[string]interface{}{
				"key": "value",
			},
			want:    `{"key":"value"}`,
			wantErr: false,
		},
		{
			name:    "array",
			input:   []int{1, 2, 3},
			want:    "[1,2,3]",
			wantErr: false,
		},
		{
			name:    "nil",
			input:   nil,
			want:    "null",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := toJSON(tt.input)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestFunctionNode_MultipleExecutions(t *testing.T) {
	n := NewFunctionNode()
	err := n.Init(map[string]interface{}{
		"code":       "return msg.payload.index * 2",
		"output_key": "doubled",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Execute multiple times
	for i := 0; i < 5; i++ {
		msg := node.Message{
			Payload: map[string]interface{}{
				"index": i,
			},
		}

		result, err := n.Execute(ctx, msg)
		require.NoError(t, err)
		assert.NotNil(t, result.Payload["doubled"])
		// Original payload should still have index
		assert.Equal(t, i, result.Payload["index"])
	}
}
