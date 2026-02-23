package core

import (
	"context"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDelayNode(t *testing.T) {
	n := NewDelayNode()
	require.NotNil(t, n)
	assert.Equal(t, 1*time.Second, n.duration)
	assert.Equal(t, 1*time.Minute, n.timeout)
}

func TestDelayNode_Init(t *testing.T) {
	tests := []struct {
		name        string
		config      map[string]interface{}
		wantErr     bool
		errContains string
		checkResult func(*testing.T, *DelayNode)
	}{
		{
			name: "valid duration string",
			config: map[string]interface{}{
				"duration": "500ms",
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *DelayNode) {
				assert.Equal(t, 500*time.Millisecond, n.duration)
			},
		},
		{
			name: "valid duration in milliseconds",
			config: map[string]interface{}{
				"duration": 200.0,
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *DelayNode) {
				assert.Equal(t, 200*time.Millisecond, n.duration)
			},
		},
		{
			name: "valid timeout string",
			config: map[string]interface{}{
				"duration": "1s",
				"timeout":  "5m",
			},
			wantErr: false,
			checkResult: func(t *testing.T, n *DelayNode) {
				assert.Equal(t, 5*time.Minute, n.timeout)
			},
		},
		{
			name: "invalid duration format",
			config: map[string]interface{}{
				"duration": "invalid",
			},
			wantErr:     true,
			errContains: "invalid duration format",
		},
		{
			name: "invalid timeout format",
			config: map[string]interface{}{
				"duration": "1s",
				"timeout":  "invalid",
			},
			wantErr:     true,
			errContains: "invalid timeout format",
		},
		{
			name: "negative duration",
			config: map[string]interface{}{
				"duration": -100.0,
			},
			wantErr:     true,
			errContains: "cannot be negative",
		},
		{
			name: "duration exceeds timeout",
			config: map[string]interface{}{
				"duration": "10m",
				"timeout":  "1m",
			},
			wantErr:     true,
			errContains: "cannot exceed timeout",
		},
		{
			name:    "empty config uses defaults",
			config:  map[string]interface{}{},
			wantErr: false,
			checkResult: func(t *testing.T, n *DelayNode) {
				assert.Equal(t, 1*time.Second, n.duration)
				assert.Equal(t, 1*time.Minute, n.timeout)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := NewDelayNode()
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

func TestDelayNode_Execute(t *testing.T) {
	n := NewDelayNode()
	err := n.Init(map[string]interface{}{
		"duration": "50ms",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"data": "test",
		},
	}

	start := time.Now()
	result, err := n.Execute(ctx, msg)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, msg.Payload, result.Payload)
	// Verify delay was applied (at least 40ms to account for timing)
	assert.True(t, elapsed >= 40*time.Millisecond)
}

func TestDelayNode_Execute_ZeroDelay(t *testing.T) {
	n := NewDelayNode()
	err := n.Init(map[string]interface{}{
		"duration": "0s",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"value": 123,
		},
	}

	start := time.Now()
	result, err := n.Execute(ctx, msg)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, msg.Payload, result.Payload)
	// Zero delay should complete very quickly
	assert.True(t, elapsed < 50*time.Millisecond)
}

func TestDelayNode_Execute_ContextCancellation(t *testing.T) {
	n := NewDelayNode()
	err := n.Init(map[string]interface{}{
		"duration": "5s", // Long delay
	})
	require.NoError(t, err)

	// Create a context that will be cancelled
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	msg := node.Message{
		Payload: map[string]interface{}{
			"data": "test",
		},
	}

	start := time.Now()
	_, err = n.Execute(ctx, msg)
	elapsed := time.Since(start)

	// Should error due to context cancellation
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cancelled")
	// Should not wait for the full 5s delay
	assert.True(t, elapsed < 1*time.Second)
}

func TestDelayNode_Execute_PreservesMessage(t *testing.T) {
	n := NewDelayNode()
	err := n.Init(map[string]interface{}{
		"duration": "10ms",
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
			"nested": map[string]interface{}{
				"inner": "data",
			},
		},
		Topic: "test/topic",
	}

	result, err := n.Execute(ctx, msg)
	require.NoError(t, err)

	// Verify all message fields are preserved
	assert.Equal(t, msg.Type, result.Type)
	assert.Equal(t, msg.Topic, result.Topic)
	assert.Equal(t, msg.Payload["key1"], result.Payload["key1"])
	assert.Equal(t, msg.Payload["key2"], result.Payload["key2"])

	nested := result.Payload["nested"].(map[string]interface{})
	assert.Equal(t, "data", nested["inner"])
}

func TestDelayNode_Cleanup(t *testing.T) {
	n := NewDelayNode()
	err := n.Init(map[string]interface{}{
		"duration": "1s",
	})
	require.NoError(t, err)

	err = n.Cleanup()
	assert.NoError(t, err)
}

func TestDelayNode_MultipleExecutions(t *testing.T) {
	n := NewDelayNode()
	err := n.Init(map[string]interface{}{
		"duration": "20ms",
	})
	require.NoError(t, err)

	ctx := context.Background()

	// Execute multiple times
	for i := 0; i < 3; i++ {
		msg := node.Message{
			Payload: map[string]interface{}{
				"iteration": i,
			},
		}

		result, err := n.Execute(ctx, msg)
		require.NoError(t, err)
		assert.Equal(t, i, result.Payload["iteration"])
	}
}

func TestDelayNode_ConcurrentExecutions(t *testing.T) {
	n := NewDelayNode()
	err := n.Init(map[string]interface{}{
		"duration": "50ms",
	})
	require.NoError(t, err)

	ctx := context.Background()
	numGoroutines := 5
	results := make(chan int, numGoroutines)

	// Launch concurrent executions
	for i := 0; i < numGoroutines; i++ {
		go func(idx int) {
			msg := node.Message{
				Payload: map[string]interface{}{
					"index": idx,
				},
			}
			result, err := n.Execute(ctx, msg)
			if err == nil {
				results <- result.Payload["index"].(int)
			}
		}(i)
	}

	// Collect results
	collected := make([]int, 0, numGoroutines)
	timeout := time.After(2 * time.Second)
	for i := 0; i < numGoroutines; i++ {
		select {
		case r := <-results:
			collected = append(collected, r)
		case <-timeout:
			t.Fatal("timeout waiting for concurrent results")
		}
	}

	assert.Len(t, collected, numGoroutines)
}
