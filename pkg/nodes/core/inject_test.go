package core

import (
	"context"
	"testing"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewInjectNode(t *testing.T) {
	n := NewInjectNode()
	require.NotNil(t, n)
	assert.Equal(t, 1*time.Second, n.interval)
	assert.NotNil(t, n.payload)
	assert.NotNil(t, n.stopChan)
}

func TestInjectNode_Init(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		wantErr     bool
		checkResult func(*testing.T, *InjectNode)
	}{
		{
			name: "valid interval string",
			config: map[string]interface{}{
				"interval": "5s",
				"topic":    "test/topic",
				"payload": map[string]interface{}{
					"key": "value",
				},
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *InjectNode) {
				assert.Equal(t, 5*time.Second, n.interval)
				assert.Equal(t, "test/topic", n.topic)
				assert.Equal(t, "value", n.payload["key"])
			},
		},
		{
			name: "valid interval with milliseconds",
			config: map[string]interface{}{
				"interval": "500ms",
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *InjectNode) {
				assert.Equal(t, 500*time.Millisecond, n.interval)
			},
		},
		{
			name: "invalid interval format",
			config: map[string]interface{}{
				"interval": "invalid",
			},
			wantErr: true,
		},
		{
			name:    "empty config uses defaults",
			config:  map[string]interface{}{},
			wantErr: false,
			checkResult: func(t *testing.T, n *InjectNode) {
				assert.Equal(t, 1*time.Second, n.interval)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewInjectNode()
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

func TestInjectNode_Execute(t *testing.T) {
	n := NewInjectNode()
	err := n.Init(map[string]interface{}{
		"payload": map[string]interface{}{
			"temperature": 25.5,
			"unit":        "celsius",
		},
		"topic": "sensors/temp",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Type: node.MessageTypeEvent,
		Payload: map[string]interface{}{
			"trigger": true,
		},
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.Equal(t, node.MessageTypeData, result.Type)
	assert.Equal(t, "sensors/temp", result.Topic)
	assert.Equal(t, 25.5, result.Payload["temperature"])
	assert.Equal(t, "celsius", result.Payload["unit"])
}

func TestInjectNode_ExecuteWithoutTrigger(t *testing.T) {
	n := NewInjectNode()
	err := n.Init(map[string]interface{}{
		"payload": map[string]interface{}{
			"message": "hello",
		},
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)
	assert.Equal(t, "hello", result.Payload["message"])
}

func TestInjectNode_Cleanup(t *testing.T) {
	n := NewInjectNode()
	err := n.Init(map[string]interface{}{})
	require.NoError(t, err)

	// Cleanup should not error
	err = n.Cleanup()
	assert.NoError(t, err)
}

func TestInjectNode_CreateMessage(t *testing.T) {
	n := NewInjectNode()
	n.payload = map[string]interface{}{
		"data": "test",
	}
	n.topic = "test/topic"

	msg := n.createMessage()
	assert.Equal(t, node.MessageTypeData, msg.Type)
	assert.Equal(t, "test/topic", msg.Topic)
	assert.Equal(t, "test", msg.Payload["data"])
}
