package core

import (
	"context"
	"testing"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewIfNode(t *testing.T) {
	n := NewIfNode()
	require.NotNil(t, n)
	assert.Equal(t, "eq", n.operator)
}

func TestIfNode_Init(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		wantErr     bool
		checkResult func(*testing.T, *IfNode)
	}{
		{
			name: "full config",
			config: map[string]interface{}{
				"field":    "temperature",
				"operator": "gt",
				"value":    30.0,
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *IfNode) {
				assert.Equal(t, "temperature", n.field)
				assert.Equal(t, "gt", n.operator)
				assert.Equal(t, 30.0, n.value)
			},
		},
		{
			name: "with condition expression",
			config: map[string]interface{}{
				"condition": "msg.payload.value > 100",
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *IfNode) {
				assert.Equal(t, "msg.payload.value > 100", n.condition)
			},
		},
		{
			name:    "empty config uses defaults",
			config:  map[string]interface{}{},
			wantErr: false,
			checkResult: func(t *testing.T, n *IfNode) {
				assert.Equal(t, "eq", n.operator)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewIfNode()
			err := n.Init(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			if tt.checkResult != nil {
				tt.checkResult(t, n)
			}
		})
	}
}

func TestIfNode_Execute_Equal(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "status",
		"operator": "eq",
		"value":    "active",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test true case
	msg := node.Message{
		Payload: map[string]interface{}{
			"status": "active",
		},
	}
	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test false case
	msg = node.Message{
		Payload: map[string]interface{}{
			"status": "inactive",
		},
	}
	result, err = n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.False(t, result.Payload["_condition_result"].(bool))
}

func TestIfNode_Execute_NotEqual(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "type",
		"operator": "ne",
		"value":    "error",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test true case (value is not "error")
	msg := node.Message{
		Payload: map[string]interface{}{
			"type": "success",
		},
	}
	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test false case (value is "error")
	msg = node.Message{
		Payload: map[string]interface{}{
			"type": "error",
		},
	}
	result, err = n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.False(t, result.Payload["_condition_result"].(bool))
}

func TestIfNode_Execute_GreaterThan(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "temperature",
		"operator": "gt",
		"value":    25.0,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test true case (30 > 25)
	msg := node.Message{
		Payload: map[string]interface{}{
			"temperature": 30.0,
		},
	}
	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test false case (20 > 25)
	msg = node.Message{
		Payload: map[string]interface{}{
			"temperature": 20.0,
		},
	}
	result, err = n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.False(t, result.Payload["_condition_result"].(bool))
}

func TestIfNode_Execute_LessThan(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "count",
		"operator": "lt",
		"value":    10,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test true case (5 < 10)
	msg := node.Message{
		Payload: map[string]interface{}{
			"count": 5,
		},
	}
	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test false case (15 < 10)
	msg = node.Message{
		Payload: map[string]interface{}{
			"count": 15,
		},
	}
	result, err = n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.False(t, result.Payload["_condition_result"].(bool))
}

func TestIfNode_Execute_GreaterThanOrEqual(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "value",
		"operator": "gte",
		"value":    50,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test true case (exactly 50)
	msg := node.Message{
		Payload: map[string]interface{}{
			"value": 50,
		},
	}
	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test true case (greater than 50)
	msg = node.Message{
		Payload: map[string]interface{}{
			"value": 100,
		},
	}
	result, err = n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test false case (less than 50)
	msg = node.Message{
		Payload: map[string]interface{}{
			"value": 49,
		},
	}
	result, err = n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.False(t, result.Payload["_condition_result"].(bool))
}

func TestIfNode_Execute_LessThanOrEqual(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "humidity",
		"operator": "lte",
		"value":    70,
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test true case (exactly 70)
	msg := node.Message{
		Payload: map[string]interface{}{
			"humidity": 70,
		},
	}
	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test false case (greater than 70)
	msg = node.Message{
		Payload: map[string]interface{}{
			"humidity": 80,
		},
	}
	result, err = n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.False(t, result.Payload["_condition_result"].(bool))
}

func TestIfNode_Execute_Contains(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "message",
		"operator": "contains",
		"value":    "error",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test true case
	msg := node.Message{
		Payload: map[string]interface{}{
			"message": "An error occurred",
		},
	}
	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test false case
	msg = node.Message{
		Payload: map[string]interface{}{
			"message": "Success!",
		},
	}
	result, err = n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.False(t, result.Payload["_condition_result"].(bool))
}

func TestIfNode_Execute_Exists(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "optional_field",
		"operator": "exists",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Test true case (field exists)
	msg := node.Message{
		Payload: map[string]interface{}{
			"optional_field": "some value",
		},
	}
	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, result.Payload["_condition_result"].(bool))

	// Test false case (field doesn't exist)
	msg = node.Message{
		Payload: map[string]interface{}{
			"other_field": "value",
		},
	}
	_, err = n.Execute(ctx, msg)
	assert.Error(t, err) // Field not found
}

func TestIfNode_Execute_FieldNotFound(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "missing_field",
		"operator": "eq",
		"value":    "test",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"some_field": "value",
		},
	}

	_, err = n.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIfNode_Execute_UnknownOperator(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "value",
		"operator": "unknown",
		"value":    10,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"value": 10,
		},
	}

	_, err = n.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown operator")
}

func TestIfNode_Execute_NonNumericComparison(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "name",
		"operator": "gt",
		"value":    10,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"name": "test string",
		},
	}

	_, err = n.Execute(ctx, msg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "non-numeric")
}

func TestIfNode_Execute_BranchMetadata(t *testing.T) {
	n := NewIfNode()
	err := n.Init(map[string]interface{}{
		"field":    "active",
		"operator": "eq",
		"value":    true,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"active": true,
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	branch := result.Payload["_branch"].(map[string]bool)
	assert.True(t, branch["true"])
	assert.False(t, branch["false"])
}

func TestIfNode_Cleanup(t *testing.T) {
	n := NewIfNode()
	err := n.Cleanup()
	assert.NoError(t, err)
}

func TestToFloat64(t *testing.T) {
	tests := []struct {
		name   string
		input  interface{}
		want   float64
		wantOk bool
	}{
		{"float64", 10.5, 10.5, true},
		{"float32", float32(5.5), 5.5, true},
		{"int", 100, 100.0, true},
		{"int64", int64(200), 200.0, true},
		{"int32", int32(50), 50.0, true},
		{"string", "test", 0, false},
		{"nil", nil, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := toFloat64(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCompareEqual(t *testing.T) {
	assert.True(t, compareEqual("test", "test"))
	assert.True(t, compareEqual(100, 100))
	assert.True(t, compareEqual(10.5, 10.5))
	assert.False(t, compareEqual("test", "other"))
	assert.False(t, compareEqual(100, 200))
}

func TestContains(t *testing.T) {
	assert.True(t, contains("hello world", "world"))
	assert.True(t, contains("test", "test"))
	assert.False(t, contains("hello", "world"))
	assert.False(t, contains("", "test"))
}
