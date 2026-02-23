package core

import (
	"context"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTrigger_SendMode(t *testing.T) {
	// Create Trigger node in send mode
	trigger := NewTriggerNode()
	err := trigger.Init(map[string]interface{}{
		"op":             "send",
		"initialPayload": "triggered",
	})
	require.NoError(t, err)

	err = trigger.Start(context.Background())
	require.NoError(t, err)

	// Send message
	inputMsg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"input": "data"},
		Topic:   "test/trigger",
	}

	outputMsg, err := trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)

	// Should immediately return with initialPayload wrapped in map
	assert.Equal(t, map[string]interface{}{"value": "triggered"}, outputMsg.Payload)
	assert.Equal(t, inputMsg.Topic, outputMsg.Topic)

	// Cleanup
	trigger.Cleanup()
}

func TestTrigger_SendThenSend(t *testing.T) {
	// Create Trigger node in send-then-send mode
	trigger := NewTriggerNode()
	err := trigger.Init(map[string]interface{}{
		"op":             "send-then-send",
		"initialPayload": "first",
		"secondPayload":  "second",
		"delay":          "100ms",
	})
	require.NoError(t, err)

	ctx := context.Background()
	err = trigger.Start(ctx)
	require.NoError(t, err)

	// Send message
	inputMsg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"input": "data"},
	}

	// Execute returns empty because messages go to output channel
	outputMsg, err := trigger.Execute(ctx, inputMsg)
	require.NoError(t, err)
	assert.Empty(t, outputMsg.Payload)

	// Get output channel
	outputChan := trigger.GetOutputChannel()

	// Should immediately receive first message
	select {
	case msg1 := <-outputChan:
		assert.Equal(t, map[string]interface{}{"value": "first"}, msg1.Payload)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Did not receive first message")
	}

	// Should receive second message after delay
	select {
	case msg2 := <-outputChan:
		assert.Equal(t, map[string]interface{}{"value": "second"}, msg2.Payload)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive second message")
	}

	// Cleanup
	trigger.Cleanup()
}

func TestTrigger_SendThenSendPassThrough(t *testing.T) {
	// Create Trigger with nil payloads (pass through mode)
	trigger := NewTriggerNode()
	err := trigger.Init(map[string]interface{}{
		"op":             "send-then-send",
		"initialPayload": nil, // Pass through original
		"secondPayload":  map[string]interface{}{"done": true},
		"delay":          "100ms",
	})
	require.NoError(t, err)

	err = trigger.Start(context.Background())
	require.NoError(t, err)

	// Send message
	inputMsg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"original": "data"},
	}

	_, err = trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)

	outputChan := trigger.GetOutputChannel()

	// First message should pass through original payload
	select {
	case msg1 := <-outputChan:
		assert.Equal(t, inputMsg.Payload, msg1.Payload)
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Did not receive first message")
	}

	// Second message should have specified payload
	select {
	case msg2 := <-outputChan:
		expected := map[string]interface{}{"done": true}
		assert.Equal(t, expected, msg2.Payload)
	case <-time.After(200 * time.Millisecond):
		t.Fatal("Did not receive second message")
	}

	// Cleanup
	trigger.Cleanup()
}

func TestTrigger_ExtendDelay(t *testing.T) {
	// Create Trigger with extend mode
	trigger := NewTriggerNode()
	err := trigger.Init(map[string]interface{}{
		"op":             "send-then-send",
		"initialPayload": "first",
		"secondPayload":  "second",
		"delay":          "150ms",
		"extend":         true, // Extend delay on new messages
	})
	require.NoError(t, err)

	err = trigger.Start(context.Background())
	require.NoError(t, err)

	inputMsg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"data": 1},
	}

	outputChan := trigger.GetOutputChannel()

	// Send first message
	_, err = trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)

	// Receive first immediate message
	<-outputChan

	// Wait 100ms and send another message (should extend the delay)
	time.Sleep(100 * time.Millisecond)
	inputMsg.Payload = map[string]interface{}{"data": 2}
	_, err = trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)

	// Receive second immediate message
	<-outputChan

	// The second delayed message should come after another 150ms from now
	// (not 50ms as it would be without extend)
	start := time.Now()

	select {
	case msg := <-outputChan:
		elapsed := time.Since(start)
		assert.Equal(t, map[string]interface{}{"value": "second"}, msg.Payload)
		// Should take approximately 150ms (with extend)
		assert.True(t, elapsed >= 140*time.Millisecond, "elapsed: %v", elapsed)
		assert.True(t, elapsed <= 200*time.Millisecond, "elapsed: %v", elapsed)
	case <-time.After(300 * time.Millisecond):
		t.Fatal("Did not receive second message")
	}

	// Cleanup
	trigger.Cleanup()
}

func TestTrigger_SendThenNothing(t *testing.T) {
	// Create Trigger in send-then-nothing mode (debouncing)
	trigger := NewTriggerNode()
	err := trigger.Init(map[string]interface{}{
		"op":             "send-then-nothing",
		"initialPayload": "allowed",
		"duration":       "200ms", // Block for 200ms after sending
	})
	require.NoError(t, err)

	err = trigger.Start(context.Background())
	require.NoError(t, err)

	inputMsg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"attempt": 1},
	}

	// First message should go through
	msg1, err := trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"value": "allowed"}, msg1.Payload)

	// Second message immediately after should be blocked
	inputMsg.Payload = map[string]interface{}{"attempt": 2}
	msg2, err := trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)
	assert.Empty(t, msg2.Payload) // Blocked

	// Third message immediately after should also be blocked
	inputMsg.Payload = map[string]interface{}{"attempt": 3}
	msg3, err := trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)
	assert.Empty(t, msg3.Payload) // Blocked

	// Wait for duration to expire
	time.Sleep(250 * time.Millisecond)

	// Fourth message should go through again
	inputMsg.Payload = map[string]interface{}{"attempt": 4}
	msg4, err := trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)
	assert.Equal(t, map[string]interface{}{"value": "allowed"}, msg4.Payload)

	// Cleanup
	trigger.Cleanup()
}

func TestTrigger_DelayConfig(t *testing.T) {
	// Test different delay configuration formats

	tests := []struct {
		name     string
		config   interface{}
		expected time.Duration
	}{
		{
			name:     "string duration",
			config:   "500ms",
			expected: 500 * time.Millisecond,
		},
		{
			name:     "numeric milliseconds",
			config:   float64(1000),
			expected: 1000 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := NewTriggerNode()
			err := trigger.Init(map[string]interface{}{
				"delay": tt.config,
			})
			require.NoError(t, err)
			assert.Equal(t, tt.expected, trigger.delay)
		})
	}
}

func TestTrigger_Cleanup(t *testing.T) {
	// Test that cleanup properly stops timers
	trigger := NewTriggerNode()
	err := trigger.Init(map[string]interface{}{
		"op":             "send-then-send",
		"initialPayload": "first",
		"secondPayload":  "second",
		"delay":          "500ms", // Long delay
	})
	require.NoError(t, err)

	err = trigger.Start(context.Background())
	require.NoError(t, err)

	// Send message
	inputMsg := node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"test": true},
	}

	_, err = trigger.Execute(context.Background(), inputMsg)
	require.NoError(t, err)

	// Receive first message
	outputChan := trigger.GetOutputChannel()
	<-outputChan

	// Cleanup before second message
	err = trigger.Cleanup()
	require.NoError(t, err)

	// Channel should be closed
	select {
	case _, ok := <-outputChan:
		assert.False(t, ok, "Channel should be closed")
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Channel was not closed")
	}
}
