package core

import (
	"context"
	"strings"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDebugNode(t *testing.T) {
	n := NewDebugNode()
	require.NotNil(t, n)
	assert.Equal(t, "console", n.outputTo)
	assert.False(t, n.complete)
	assert.NotNil(t, n.outputFunc)
}

func TestDebugNode_Init(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		wantErr     bool
		checkResult func(*testing.T, *DebugNode)
	}{
		{
			name: "set output to log",
			config: map[string]interface{}{
				"output_to": "log",
				"complete":  true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *DebugNode) {
				assert.Equal(t, "log", n.outputTo)
				assert.True(t, n.complete)
			},
		},
		{
			name:    "empty config uses defaults",
			config:  map[string]interface{}{},
			wantErr: false,
			checkResult: func(t *testing.T, n *DebugNode) {
				assert.Equal(t, "console", n.outputTo)
				assert.False(t, n.complete)
			},
		},
		{
			name: "only complete flag",
			config: map[string]interface{}{
				"complete": true,
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *DebugNode) {
				assert.True(t, n.complete)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewDebugNode()
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

func TestDebugNode_Execute_PayloadOnly(t *testing.T) {
	var capturedOutput string
	n := NewDebugNode()
	n.SetOutputFunc(func(s string) {
		capturedOutput = s
	})

	err := n.Init(map[string]interface{}{
		"complete": false,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"temperature": 25.5,
			"humidity":    60,
		},
		Topic: "sensors/data",
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Message should pass through unchanged
	assert.Equal(t, msg.Type, result.Type)
	assert.Equal(t, msg.Topic, result.Topic)
	assert.Equal(t, msg.Payload["temperature"], result.Payload["temperature"])

	// Output should contain payload data
	assert.True(t, strings.Contains(capturedOutput, "[DEBUG]"))
	assert.True(t, strings.Contains(capturedOutput, "temperature"))
	assert.True(t, strings.Contains(capturedOutput, "25.5"))
}

func TestDebugNode_Execute_CompleteMessage(t *testing.T) {
	var capturedOutput string
	n := NewDebugNode()
	n.SetOutputFunc(func(s string) {
		capturedOutput = s
	})

	err := n.Init(map[string]interface{}{
		"complete": true,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"value": 100,
		},
		Topic: "test/topic",
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Message should pass through unchanged
	assert.Equal(t, msg.Type, result.Type)

	// Output should contain the full message structure
	assert.True(t, strings.Contains(capturedOutput, "[DEBUG]"))
	assert.True(t, strings.Contains(capturedOutput, "Topic"))
	assert.True(t, strings.Contains(capturedOutput, "test/topic"))
}

func TestDebugNode_Execute_NilPayload(t *testing.T) {
	var capturedOutput string
	n := NewDebugNode()
	n.SetOutputFunc(func(s string) {
		capturedOutput = s
	})

	err := n.Init(map[string]interface{}{})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Type:    node.MessageTypeData,
		Payload: nil,
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.Nil(t, result.Payload)
	assert.True(t, strings.Contains(capturedOutput, "[DEBUG]"))
}

func TestDebugNode_Cleanup(t *testing.T) {
	n := NewDebugNode()
	err := n.Cleanup()
	assert.NoError(t, err)
}

func TestDebugNode_SetOutputFunc(t *testing.T) {
	var called bool
	n := NewDebugNode()
	n.SetOutputFunc(func(s string) {
		called = true
	})

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{},
	}

	_, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.True(t, called)
}
