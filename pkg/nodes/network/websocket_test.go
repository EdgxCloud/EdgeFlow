package network

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func TestNewWebSocketClientExecutor(t *testing.T) {
	tests := []struct {
		name    string
		config  map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid config",
			config: map[string]interface{}{
				"url": "ws://localhost:8080/ws",
			},
			wantErr: false,
		},
		{
			name:    "missing URL",
			config:  map[string]interface{}{},
			wantErr: true,
		},
		{
			name: "wss URL",
			config: map[string]interface{}{
				"url": "wss://secure.example.com/ws",
			},
			wantErr: false,
		},
		{
			name: "with auto reconnect",
			config: map[string]interface{}{
				"url":           "ws://localhost:8080/ws",
				"autoReconnect": true,
			},
			wantErr: false,
		},
		{
			name: "with ping interval",
			config: map[string]interface{}{
				"url":          "ws://localhost:8080/ws",
				"pingInterval": 30,
			},
			wantErr: false,
		},
		{
			name: "with headers",
			config: map[string]interface{}{
				"url": "ws://localhost:8080/ws",
				"headers": map[string]string{
					"Authorization": "Bearer token",
				},
			},
			wantErr: false,
		},
		{
			name: "with compression",
			config: map[string]interface{}{
				"url":               "ws://localhost:8080/ws",
				"enableCompression": true,
			},
			wantErr: false,
		},
		{
			name: "with subprotocols",
			config: map[string]interface{}{
				"url":          "ws://localhost:8080/ws",
				"subprotocols": []string{"graphql-ws"},
			},
			wantErr: false,
		},
		{
			name: "with max message size",
			config: map[string]interface{}{
				"url":            "ws://localhost:8080/ws",
				"maxMessageSize": 1048576,
			},
			wantErr: false,
		},
		{
			name: "with reconnect attempts",
			config: map[string]interface{}{
				"url":                  "ws://localhost:8080/ws",
				"autoReconnect":        true,
				"maxReconnectAttempts": 5,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor, err := NewWebSocketClientExecutor(tt.config)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, executor)
		})
	}
}

func TestWebSocketClientExecutor_Connect(t *testing.T) {
	// Create mock WebSocket server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Echo messages back
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				return
			}
			conn.WriteMessage(messageType, p)
		}
	}))
	defer server.Close()

	// Convert http URL to ws URL
	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	executor, err := NewWebSocketClientExecutor(map[string]interface{}{
		"url": wsURL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"send": "Hello WebSocket!",
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	// Should have sent successfully
	payload := result.Payload.(map[string]interface{})
	assert.True(t, payload["sent"].(bool))

	// Cleanup
	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestWebSocketClientExecutor_ReceiveMessage(t *testing.T) {
	// Create mock WebSocket server that sends a message
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Send a message to the client
		testMsg := map[string]interface{}{
			"type": "greeting",
			"data": "Hello from server",
		}
		msgBytes, _ := json.Marshal(testMsg)
		conn.WriteMessage(websocket.TextMessage, msgBytes)

		// Keep connection open
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	executor, err := NewWebSocketClientExecutor(map[string]interface{}{
		"url": wsURL,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// First call connects and waits for message
	result, err := executor.Execute(ctx, node.Message{})
	require.NoError(t, err)

	// Should receive the server's message
	payload := result.Payload.(map[string]interface{})
	assert.NotNil(t, payload["payload"])

	executor.Cleanup()
}

func TestWebSocketClientExecutor_BinaryMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()

		// Receive and echo binary
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			return
		}
		conn.WriteMessage(messageType, p)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	executor, err := NewWebSocketClientExecutor(map[string]interface{}{
		"url": wsURL,
	})
	require.NoError(t, err)

	ctx := context.Background()
	msg := node.Message{
		Payload: map[string]interface{}{
			"send": []byte{0x01, 0x02, 0x03, 0x04},
		},
	}

	result, err := executor.Execute(ctx, msg)
	require.NoError(t, err)

	payload := result.Payload.(map[string]interface{})
	assert.True(t, payload["sent"].(bool))

	executor.Cleanup()
}

func TestWebSocketClientExecutor_Cleanup(t *testing.T) {
	executor, err := NewWebSocketClientExecutor(map[string]interface{}{
		"url": "ws://localhost:8080/ws",
	})
	require.NoError(t, err)

	// Cleanup should work even without connection
	err = executor.Cleanup()
	assert.NoError(t, err)
}

func TestWebSocketClientExecutor_Reconnect(t *testing.T) {
	connectionCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		connectionCount++
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		// Close immediately to trigger reconnect
		conn.Close()
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	executor, err := NewWebSocketClientExecutor(map[string]interface{}{
		"url":                  wsURL,
		"autoReconnect":        true,
		"reconnectDelay":       100,
		"maxReconnectAttempts": 2,
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	// This will try to connect and potentially reconnect
	executor.Execute(ctx, node.Message{})

	// Give time for reconnect attempts
	time.Sleep(300 * time.Millisecond)

	executor.Cleanup()
}

func TestWebSocketClientExecutor_ConnectionHeaders(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		defer conn.Close()
		time.Sleep(100 * time.Millisecond)
	}))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")

	executor, err := NewWebSocketClientExecutor(map[string]interface{}{
		"url": wsURL,
		"headers": map[string]string{
			"Authorization": "Bearer test-token",
		},
	})
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	executor.Execute(ctx, node.Message{})

	assert.Equal(t, "Bearer test-token", receivedAuth)

	executor.Cleanup()
}
