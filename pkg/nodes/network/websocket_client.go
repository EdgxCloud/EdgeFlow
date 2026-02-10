package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/gorilla/websocket"
)

// WebSocketClientConfig WebSocket Client node
type WebSocketClientConfig struct {
	URL              string            `json:"url"`              // WebSocket URL (ws:// or wss://)
	Headers          map[string]string `json:"headers"`          // Connection headers
	AutoReconnect    bool              `json:"autoReconnect"`    // Auto reconnect on disconnect
	ReconnectDelay   int               `json:"reconnectDelay"`   // Delay between reconnects (ms)
	PingInterval     int               `json:"pingInterval"`     // Ping interval (seconds)
	HandshakeTimeout int               `json:"handshakeTimeout"` // Handshake timeout (seconds)

	// Compression support
	EnableCompression bool `json:"enableCompression"` // Enable permessage-deflate compression

	// Subprotocols
	Subprotocols []string `json:"subprotocols"` // Requested subprotocols (e.g., ["graphql-ws", "subscriptions-transport-ws"])

	// Heartbeat configuration
	HeartbeatInterval int    `json:"heartbeatInterval"` // Heartbeat interval (seconds)
	HeartbeatMessage  string `json:"heartbeatMessage"`  // Custom heartbeat message

	// Message handling
	MaxMessageSize int64 `json:"maxMessageSize"` // Maximum message size in bytes

	// Reconnection settings
	MaxReconnectAttempts int `json:"maxReconnectAttempts"` // Maximum reconnect attempts (0 = unlimited)
}

// WebSocketClientExecutor WebSocket Client node executor
type WebSocketClientExecutor struct {
	config           WebSocketClientConfig
	conn             *websocket.Conn
	outputChan       chan node.Message
	connected        bool
	mu               sync.RWMutex
	stopChan         chan struct{}
	reconnectCount   int
	negotiatedProto  string // Negotiated subprotocol
}

// NewWebSocketClientExecutor create WebSocketClientExecutor
func NewWebSocketClientExecutor() node.Executor {
	return &WebSocketClientExecutor{
		outputChan: make(chan node.Message, 100),
		stopChan:   make(chan struct{}),
	}
}

// Init initializes the executor with configuration
func (e *WebSocketClientExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var wsConfig WebSocketClientConfig
	if err := json.Unmarshal(configJSON, &wsConfig); err != nil {
		return fmt.Errorf("invalid websocket config: %w", err)
	}

	// Validate URL
	if wsConfig.URL == "" {
		return fmt.Errorf("URL is required")
	}

	// Default values
	if wsConfig.ReconnectDelay == 0 {
		wsConfig.ReconnectDelay = 5000
	}
	if wsConfig.PingInterval == 0 {
		wsConfig.PingInterval = 30
	}
	if wsConfig.HandshakeTimeout == 0 {
		wsConfig.HandshakeTimeout = 10
	}
	if wsConfig.MaxMessageSize == 0 {
		wsConfig.MaxMessageSize = 32 * 1024 * 1024 // 32 MB default
	}

	e.config = wsConfig
	return nil
}

// Execute execute node
func (e *WebSocketClientExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Connect if not connected
	if !e.isConnected() {
		if err := e.connect(); err != nil {
			return node.Message{}, fmt.Errorf("failed to connect: %w", err)
		}

		// Start read loop
		go e.readLoop()

		// Start ping loop if enabled
		if e.config.PingInterval > 0 {
			go e.pingLoop()
		}
	}

	// Check if we need to send a message
	if data, ok := msg.Payload["send"]; ok {
		if err := e.send(data); err != nil {
			return node.Message{}, fmt.Errorf("failed to send message: %w", err)
		}

		return node.Message{
			Payload: map[string]interface{}{
				"sent": true,
			},
		}, nil
	}

	// Wait for incoming message or context cancellation
	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case wsMsg := <-e.outputChan:
		return wsMsg, nil
	}
}

// connect connect to WebSocket server
func (e *WebSocketClientExecutor) connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.connected {
		return nil
	}

	// Create dialer with timeout and compression
	dialer := websocket.Dialer{
		HandshakeTimeout:  time.Duration(e.config.HandshakeTimeout) * time.Second,
		EnableCompression: e.config.EnableCompression,
		Subprotocols:      e.config.Subprotocols,
	}

	// Create headers
	headers := make(map[string][]string)
	for key, value := range e.config.Headers {
		headers[key] = []string{value}
	}

	// Connect
	conn, resp, err := dialer.Dial(e.config.URL, headers)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	// Store negotiated subprotocol
	if resp != nil {
		e.negotiatedProto = resp.Header.Get("Sec-WebSocket-Protocol")
	}

	// Set max message size
	conn.SetReadLimit(e.config.MaxMessageSize)

	e.conn = conn
	e.connected = true
	e.reconnectCount = 0 // Reset reconnect count on successful connection

	return nil
}

// readLoop message read loop
func (e *WebSocketClientExecutor) readLoop() {
	for {
		select {
		case <-e.stopChan:
			return
		default:
		}

		if !e.isConnected() {
			if e.config.AutoReconnect {
				// Check max reconnect attempts
				if e.config.MaxReconnectAttempts > 0 && e.reconnectCount >= e.config.MaxReconnectAttempts {
					return
				}

				e.mu.Lock()
				e.reconnectCount++
				e.mu.Unlock()

				time.Sleep(time.Duration(e.config.ReconnectDelay) * time.Millisecond)
				if err := e.connect(); err != nil {
					continue
				}
			} else {
				return
			}
		}

		messageType, message, err := e.conn.ReadMessage()
		if err != nil {
			e.mu.Lock()
			e.connected = false
			if e.conn != nil {
				e.conn.Close()
			}
			e.mu.Unlock()
			continue
		}

		// Parse message
		var payload interface{}
		if messageType == websocket.TextMessage {
			if err := json.Unmarshal(message, &payload); err != nil {
				payload = string(message)
			}
		} else {
			payload = message
		}

		// Create node message
		msg := node.Message{
			Payload: map[string]interface{}{
				"type":        messageType,
				"payload":     payload,
				"subprotocol": e.negotiatedProto,
			},
		}

		// Send to output channel (non-blocking)
		select {
		case e.outputChan <- msg:
		default:
		}
	}
}

// pingLoop ping send loop
func (e *WebSocketClientExecutor) pingLoop() {
	ticker := time.NewTicker(time.Duration(e.config.PingInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopChan:
			return
		case <-ticker.C:
			if e.isConnected() {
				e.mu.RLock()
				conn := e.conn
				e.mu.RUnlock()

				if conn != nil {
					if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
						e.mu.Lock()
						e.connected = false
						e.mu.Unlock()
					}
				}
			}
		}
	}
}

// send send message
func (e *WebSocketClientExecutor) send(data interface{}) error {
	if !e.isConnected() {
		return fmt.Errorf("not connected")
	}

	e.mu.RLock()
	conn := e.conn
	e.mu.RUnlock()

	var messageType int
	var message []byte

	switch v := data.(type) {
	case string:
		messageType = websocket.TextMessage
		message = []byte(v)
	case []byte:
		messageType = websocket.BinaryMessage
		message = v
	default:
		messageType = websocket.TextMessage
		var err error
		message, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	return conn.WriteMessage(messageType, message)
}

// isConnected check connection status
func (e *WebSocketClientExecutor) isConnected() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.connected && e.conn != nil
}

// Cleanup cleanup resources
func (e *WebSocketClientExecutor) Cleanup() error {
	close(e.stopChan)

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.conn != nil {
		e.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		e.conn.Close()
		e.conn = nil
		e.connected = false
	}

	close(e.outputChan)
	return nil
}
