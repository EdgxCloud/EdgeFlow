package core

import (
	"context"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleNode_Init(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config with standard cron",
			config: map[string]interface{}{
				"cron":    "0 * * * * *", // Every minute at 0 seconds
				"payload": map[string]interface{}{"test": "data"},
				"topic":   "scheduled/topic",
			},
			wantErr: false,
		},
		{
			name: "valid config with timezone",
			config: map[string]interface{}{
				"cron":     "0 0 12 * * *", // Every day at noon
				"timezone": "UTC",
			},
			wantErr: false,
		},
		{
			name: "missing cron expression",
			config: map[string]interface{}{
				"payload": map[string]interface{}{"test": "data"},
			},
			wantErr: true,
		},
		{
			name: "invalid cron expression",
			config: map[string]interface{}{
				"cron": "invalid",
			},
			wantErr: true,
		},
		{
			name: "invalid timezone",
			config: map[string]interface{}{
				"cron":     "0 * * * * *",
				"timezone": "Invalid/Timezone",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewScheduleNode()
			err := n.Init(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			n.Cleanup()
		})
	}
}

func TestScheduleNode_Execute(t *testing.T) {
	n := NewScheduleNode()
	err := n.Init(map[string]interface{}{
		"cron":    "* * * * * *", // Every second
		"payload": map[string]interface{}{"test": "data"},
		"topic":   "test/topic",
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()

	// Test manual trigger
	triggerMsg := node.Message{
		Type: node.MessageTypeEvent,
		Payload: map[string]interface{}{
			"trigger": true,
		},
	}

	result, err := n.Execute(ctx, triggerMsg)
	require.NoError(t, err)
	assert.Equal(t, node.MessageTypeData, result.Type)
	assert.Equal(t, "test/topic", result.Topic)
	assert.Equal(t, "data", result.Payload["test"])
	assert.True(t, result.Payload["scheduled"].(bool))
	assert.NotNil(t, result.Payload["timestamp"])
}

func TestScheduleNode_CronScheduling(t *testing.T) {
	n := NewScheduleNode()

	// Configure to trigger every 2 seconds
	err := n.Init(map[string]interface{}{
		"cron":    "*/2 * * * * *", // Every 2 seconds
		"payload": map[string]interface{}{"data": "test"},
		"topic":   "cron/test",
	})
	require.NoError(t, err)
	defer n.Cleanup()

	ctx := context.Background()
	err = n.Start(ctx)
	require.NoError(t, err)

	// Wait for scheduled messages
	receivedCount := 0
	timeout := time.After(5 * time.Second)

	for receivedCount < 2 {
		select {
		case msg := <-n.GetOutputChannel():
			assert.Equal(t, node.MessageTypeData, msg.Type)
			assert.Equal(t, "cron/test", msg.Topic)
			assert.Equal(t, "test", msg.Payload["data"])
			assert.True(t, msg.Payload["scheduled"].(bool), "Should have scheduled flag")
			assert.NotNil(t, msg.Payload["timestamp"], "Should have timestamp")
			receivedCount++
		case <-timeout:
			t.Fatalf("Timeout waiting for scheduled messages, received %d", receivedCount)
		}
	}

	assert.GreaterOrEqual(t, receivedCount, 2, "Should receive at least 2 scheduled messages")
}

func TestScheduleNode_Cleanup(t *testing.T) {
	n := NewScheduleNode()
	err := n.Init(map[string]interface{}{
		"cron":    "* * * * * *",
		"payload": map[string]interface{}{"test": "data"},
	})
	require.NoError(t, err)

	ctx := context.Background()
	err = n.Start(ctx)
	require.NoError(t, err)

	// Cleanup should stop the cron
	err = n.Cleanup()
	assert.NoError(t, err)

	// Output channel should be closed
	time.Sleep(100 * time.Millisecond)
	select {
	case _, ok := <-n.GetOutputChannel():
		assert.False(t, ok, "Output channel should be closed")
	default:
		t.Fatal("Expected channel to be closed")
	}
}

func TestScheduleNode_Timezone(t *testing.T) {
	tests := []struct {
		name     string
		timezone string
		wantErr  bool
	}{
		{
			name:     "UTC timezone",
			timezone: "UTC",
			wantErr:  false,
		},
		{
			name:     "America/New_York timezone",
			timezone: "America/New_York",
			wantErr:  false,
		},
		{
			name:     "Local timezone",
			timezone: "Local",
			wantErr:  false,
		},
		{
			name:     "Empty timezone defaults to Local",
			timezone: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewScheduleNode()
			err := n.Init(map[string]interface{}{
				"cron":     "0 0 * * * *",
				"timezone": tt.timezone,
			})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			n.Cleanup()
		})
	}
}
