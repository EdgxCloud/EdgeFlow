package network

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMQTTInExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"broker": "tcp://localhost:1883",
				"topic":  "test/topic",
			},
			wantErr: false,
		},
		{
			name: "missing broker",
			config: map[string]interface{}{
				"topic": "test/topic",
			},
			wantErr: true,
		},
		{
			name: "missing topic",
			config: map[string]interface{}{
				"broker": "tcp://localhost:1883",
			},
			wantErr: true,
		},
		{
			name: "with QoS",
			config: map[string]interface{}{
				"broker": "tcp://localhost:1883",
				"topic":  "test/topic",
				"qos":    1,
			},
			wantErr: false,
		},
		{
			name: "with credentials",
			config: map[string]interface{}{
				"broker":   "tcp://localhost:1883",
				"topic":    "test/topic",
				"username": "user",
				"password": "pass",
			},
			wantErr: false,
		},
		{
			name: "with LWT",
			config: map[string]interface{}{
				"broker":      "tcp://localhost:1883",
				"topic":       "test/topic",
				"willTopic":   "status/offline",
				"willPayload": "offline",
				"willQos":     1,
				"willRetain":  true,
			},
			wantErr: false,
		},
		{
			name: "with keep alive",
			config: map[string]interface{}{
				"broker":         "tcp://localhost:1883",
				"topic":          "test/topic",
				"keepAlive":      30,
				"connectTimeout": 10,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewMQTTInExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestNewMQTTOutExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"broker": "tcp://localhost:1883",
				"topic":  "test/output",
			},
			wantErr: false,
		},
		{
			name: "missing broker",
			config: map[string]interface{}{
				"topic": "test/output",
			},
			wantErr: true,
		},
		{
			name: "with retain flag",
			config: map[string]interface{}{
				"broker": "tcp://localhost:1883",
				"topic":  "test/output",
				"retain": true,
			},
			wantErr: false,
		},
		{
			name: "with QoS 2",
			config: map[string]interface{}{
				"broker": "tcp://localhost:1883",
				"topic":  "test/output",
				"qos":    2,
			},
			wantErr: false,
		},
		{
			name: "with message expiry",
			config: map[string]interface{}{
				"broker":        "tcp://localhost:1883",
				"topic":         "test/output",
				"messageExpiry": 3600,
			},
			wantErr: false,
		},
		{
			name: "with LWT config",
			config: map[string]interface{}{
				"broker":      "tcp://localhost:1883",
				"topic":       "test/output",
				"willTopic":   "status/device",
				"willPayload": "disconnected",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewMQTTOutExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestMQTTInConfig_Validation(t *testing.T) {
	// Test that QoS values are validated
	tests := []struct {
		name    string
		qos     int
		wantErr bool
	}{
		{"QoS 0", 0, false},
		{"QoS 1", 1, false},
		{"QoS 2", 2, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewMQTTInExecutor(map[string]interface{}{
				"broker": "tcp://localhost:1883",
				"topic":  "test/topic",
				"qos":    tt.qos,
			})

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMQTTTopicWildcards(t *testing.T) {
	// Test wildcard topics
	tests := []struct {
		name  string
		topic string
		valid bool
	}{
		{"single level wildcard", "sensors/+/temperature", true},
		{"multi level wildcard", "sensors/#", true},
		{"mixed wildcards", "home/+/room/#", true},
		{"simple topic", "sensors/temp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewMQTTInExecutor(map[string]interface{}{
				"broker": "tcp://localhost:1883",
				"topic":  tt.topic,
			})

			if tt.valid {
				assert.NoError(t, err)
				assert.NotNil(t, executor)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestMQTTOutExecutor_Cleanup(t *testing.T) {
	executor, err := NewMQTTOutExecutor(map[string]interface{}{
		"broker": "tcp://localhost:1883",
		"topic":  "test/topic",
	})
	require.NoError(t, err)

	// Cleanup should not error even if not connected
	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestMQTTInExecutor_Cleanup(t *testing.T) {
	executor, err := NewMQTTInExecutor(map[string]interface{}{
		"broker": "tcp://localhost:1883",
		"topic":  "test/topic",
	})
	require.NoError(t, err)

	// Cleanup should not error even if not connected
	err = executor.Cleanup()
	assert.NoError(t, err)
}
