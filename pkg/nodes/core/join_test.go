package core

import (
	"context"
	"testing"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinNode_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "auto mode",
			config: map[string]interface{}{
				"mode":  "auto",
				"build": "array",
			},
			wantErr: false,
		},
		{
			name: "manual mode with count",
			config: map[string]interface{}{
				"mode":  "manual",
				"count": 3,
			},
			wantErr: false,
		},
		{
			name: "manual mode without count",
			config: map[string]interface{}{
				"mode": "manual",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewJoinNode()
			err := n.Init(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestJoinNode_AutoModeArray(t *testing.T) {
	config := map[string]interface{}{
		"mode":  "auto",
		"build": "array",
	}

	n := NewJoinNode()
	err := n.Init(config)
	require.NoError(t, err)

	// Create sequence of 3 messages
	sequenceID := "test-seq-123"
	messages := []*node.EnhancedMessage{
		{
			Payload: "first",
			Topic:   "test",
			ID:      "msg1",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 0,
				Count: 3,
				Type:  "array",
			},
			Metadata: make(map[string]interface{}),
		},
		{
			Payload: "second",
			Topic:   "test",
			ID:      "msg2",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 1,
				Count: 3,
				Type:  "array",
			},
			Metadata: make(map[string]interface{}),
		},
		{
			Payload: "third",
			Topic:   "test",
			ID:      "msg3",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 2,
				Count: 3,
				Type:  "array",
			},
			Metadata: make(map[string]interface{}),
		},
	}

	// Send messages in order
	for i, msg := range messages {
		result, err := n.Execute(context.Background(), node.Message{
			Payload: msg.Payload,
			Topic:   msg.Topic,
		})
		require.NoError(t, err)

		if i < 2 {
			// First two messages shouldn't produce output
			assert.Nil(t, result.Payload)
		} else {
			// Last message should trigger join
			if arr, ok := result.Payload.([]interface{}); ok {
				assert.Equal(t, 3, len(arr))
				assert.Equal(t, "first", arr[0])
				assert.Equal(t, "second", arr[1])
				assert.Equal(t, "third", arr[2])
			}
		}
	}
}

func TestJoinNode_ManualMode(t *testing.T) {
	config := map[string]interface{}{
		"mode":  "manual",
		"count": 3,
		"build": "array",
	}

	n := NewJoinNode()
	err := n.Init(config)
	require.NoError(t, err)

	// Send 3 messages
	messages := []string{"msg1", "msg2", "msg3"}

	for i, msgPayload := range messages {
		msg := &node.EnhancedMessage{
			Payload:  msgPayload,
			Topic:    "test",
			ID:       msgPayload,
			Metadata: make(map[string]interface{}),
		}

		result, err := n.Execute(context.Background(), node.Message{
			Payload: msg.Payload,
			Topic:   msg.Topic,
		})
		require.NoError(t, err)

		if i < 2 {
			// First two messages don't produce output
			assert.Nil(t, result.Payload)
		} else {
			// Third message triggers join
			if arr, ok := result.Payload.([]interface{}); ok {
				assert.Equal(t, 3, len(arr))
			}
		}
	}
}

func TestJoinNode_ObjectMode(t *testing.T) {
	config := map[string]interface{}{
		"mode":  "auto",
		"build": "object",
	}

	n := NewJoinNode()
	err := n.Init(config)
	require.NoError(t, err)

	sequenceID := "obj-seq-123"

	// Create sequence of object properties
	messages := []*node.EnhancedMessage{
		{
			Payload: 25.5,
			Topic:   "test",
			ID:      "msg1",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 0,
				Count: 2,
				Type:  "object",
				Key:   "temperature",
			},
			Metadata: make(map[string]interface{}),
		},
		{
			Payload: 60,
			Topic:   "test",
			ID:      "msg2",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 1,
				Count: 2,
				Type:  "object",
				Key:   "humidity",
			},
			Metadata: make(map[string]interface{}),
		},
	}

	// Send both messages
	for i, msg := range messages {
		result, err := n.Execute(context.Background(), node.Message{
			Payload: msg.Payload,
			Topic:   msg.Topic,
		})
		require.NoError(t, err)

		if i < 1 {
			assert.Nil(t, result.Payload)
		} else {
			// Second message triggers join
			if obj, ok := result.Payload.(map[string]interface{}); ok {
				assert.Equal(t, 2, len(obj))
				assert.Equal(t, 25.5, obj["temperature"])
				assert.Equal(t, 60, obj["humidity"])
			}
		}
	}
}

func TestJoinNode_StringMode(t *testing.T) {
	config := map[string]interface{}{
		"mode":   "auto",
		"build":  "string",
		"joiner": "\\n",
	}

	n := NewJoinNode()
	err := n.Init(config)
	require.NoError(t, err)

	sequenceID := "str-seq-123"

	messages := []*node.EnhancedMessage{
		{
			Payload: "line1",
			Topic:   "test",
			ID:      "msg1",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 0,
				Count: 3,
				Type:  "string",
			},
			Metadata: make(map[string]interface{}),
		},
		{
			Payload: "line2",
			Topic:   "test",
			ID:      "msg2",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 1,
				Count: 3,
				Type:  "string",
			},
			Metadata: make(map[string]interface{}),
		},
		{
			Payload: "line3",
			Topic:   "test",
			ID:      "msg3",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 2,
				Count: 3,
				Type:  "string",
			},
			Metadata: make(map[string]interface{}),
		},
	}

	// Send all messages
	for i, msg := range messages {
		result, err := n.Execute(context.Background(), node.Message{
			Payload: msg.Payload,
			Topic:   msg.Topic,
		})
		require.NoError(t, err)

		if i < 2 {
			assert.Nil(t, result.Payload)
		} else {
			// Last message triggers join
			expected := "line1\nline2\nline3"
			assert.Equal(t, expected, result.Payload)
		}
	}
}

func TestJoinNode_MergeMode(t *testing.T) {
	config := map[string]interface{}{
		"mode":  "merge",
		"count": 2,
	}

	n := NewJoinNode()
	err := n.Init(config)
	require.NoError(t, err)

	// Send two objects to merge
	msg1 := &node.EnhancedMessage{
		Payload: map[string]interface{}{
			"temp": 25.5,
			"time": "12:00",
		},
		Topic:    "test",
		ID:       "msg1",
		Metadata: make(map[string]interface{}),
	}

	msg2 := &node.EnhancedMessage{
		Payload: map[string]interface{}{
			"humidity": 60,
			"pressure": 1013,
		},
		Topic:    "test",
		ID:       "msg2",
		Metadata: make(map[string]interface{}),
	}

	// First message
	result, err := n.Execute(context.Background(), node.Message{
		Payload: msg1.Payload,
		Topic:   msg1.Topic,
	})
	require.NoError(t, err)
	assert.Nil(t, result.Payload)

	// Second message triggers merge
	result, err = n.Execute(context.Background(), node.Message{
		Payload: msg2.Payload,
		Topic:   msg2.Topic,
	})
	require.NoError(t, err)

	if merged, ok := result.Payload.(map[string]interface{}); ok {
		assert.Equal(t, 4, len(merged))
		assert.Equal(t, 25.5, merged["temp"])
		assert.Equal(t, "12:00", merged["time"])
		assert.Equal(t, 60, merged["humidity"])
		assert.Equal(t, 1013, merged["pressure"])
	}
}

func TestJoinNode_OutOfOrderMessages(t *testing.T) {
	config := map[string]interface{}{
		"mode":  "auto",
		"build": "array",
	}

	n := NewJoinNode()
	err := n.Init(config)
	require.NoError(t, err)

	sequenceID := "out-of-order-123"

	// Create messages but send out of order
	messages := []*node.EnhancedMessage{
		{
			Payload: "second",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 1,
				Count: 3,
				Type:  "array",
			},
			Metadata: make(map[string]interface{}),
		},
		{
			Payload: "first",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 0,
				Count: 3,
				Type:  "array",
			},
			Metadata: make(map[string]interface{}),
		},
		{
			Payload: "third",
			Parts: &node.MessageParts{
				ID:    sequenceID,
				Index: 2,
				Count: 3,
				Type:  "array",
			},
			Metadata: make(map[string]interface{}),
		},
	}

	// Send in out-of-order
	for i, msg := range messages {
		result, err := n.Execute(context.Background(), node.Message{
			Payload: msg.Payload,
		})
		require.NoError(t, err)

		if i < 2 {
			assert.Nil(t, result.Payload)
		} else {
			// Should still assemble in correct order
			if arr, ok := result.Payload.([]interface{}); ok {
				assert.Equal(t, 3, len(arr))
				assert.Equal(t, "first", arr[0])
				assert.Equal(t, "second", arr[1])
				assert.Equal(t, "third", arr[2])
			}
		}
	}
}

func TestJoinNode_Cleanup(t *testing.T) {
	n := NewJoinNode()

	// Add some sequences
	n.sequences["test"] = &messageSequence{
		messages: make([]*node.EnhancedMessage, 0),
	}

	err := n.Cleanup()
	require.NoError(t, err)

	// Sequences should be cleared
	assert.Equal(t, 0, len(n.sequences))
}
