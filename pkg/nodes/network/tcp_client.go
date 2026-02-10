package network

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// TCPClientConfig TCP Client node
type TCPClientConfig struct {
	Host           string `json:"host"`           // Host address
	Port           int    `json:"port"`           // Port number
	AutoReconnect  bool   `json:"autoReconnect"`  // Auto reconnect
	ReconnectDelay int    `json:"reconnectDelay"` // Reconnect delay (ms)
	Timeout        int    `json:"timeout"`        // Connection timeout (seconds)
}

// TCPClientExecutor TCP Client node executor
type TCPClientExecutor struct {
	config     TCPClientConfig
	conn       net.Conn
	outputChan chan node.Message
	connected  bool
	mu         sync.RWMutex
	stopChan   chan struct{}
}

// NewTCPClientExecutor create TCPClientExecutor
func NewTCPClientExecutor() node.Executor {
	return &TCPClientExecutor{
		outputChan: make(chan node.Message, 100),
		stopChan:   make(chan struct{}),
	}
}

// Init initializes the executor with configuration
func (e *TCPClientExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var tcpConfig TCPClientConfig
	if err := json.Unmarshal(configJSON, &tcpConfig); err != nil {
		return fmt.Errorf("invalid tcp config: %w", err)
	}

	// Validate
	if tcpConfig.Host == "" {
		return fmt.Errorf("host is required")
	}
	if tcpConfig.Port == 0 {
		return fmt.Errorf("port is required")
	}

	// Default values
	if tcpConfig.ReconnectDelay == 0 {
		tcpConfig.ReconnectDelay = 5000
	}
	if tcpConfig.Timeout == 0 {
		tcpConfig.Timeout = 10
	}

	e.config = tcpConfig
	return nil
}

// Execute execute node
func (e *TCPClientExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Connect if not connected
	if !e.isConnected() {
		if err := e.connect(); err != nil {
			return node.Message{}, fmt.Errorf("failed to connect: %w", err)
		}

		// Start read loop
		go e.readLoop()
	}

	// Check if we need to send data
	if data, ok := msg.Payload["send"]; ok {
		if err := e.send(data); err != nil {
			return node.Message{}, fmt.Errorf("failed to send: %w", err)
		}

		return node.Message{
			Payload: map[string]interface{}{
				"sent": true,
			},
		}, nil
	}

	// Wait for incoming data or context cancellation
	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case tcpMsg := <-e.outputChan:
		return tcpMsg, nil
	}
}

// connect connect to TCP server
func (e *TCPClientExecutor) connect() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.connected {
		return nil
	}

	address := fmt.Sprintf("%s:%d", e.config.Host, e.config.Port)
	timeout := time.Duration(e.config.Timeout) * time.Second

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return fmt.Errorf("dial failed: %w", err)
	}

	e.conn = conn
	e.connected = true

	return nil
}

// readLoop data read loop
func (e *TCPClientExecutor) readLoop() {
	reader := bufio.NewReader(e.conn)

	for {
		select {
		case <-e.stopChan:
			return
		default:
		}

		if !e.isConnected() {
			if e.config.AutoReconnect {
				time.Sleep(time.Duration(e.config.ReconnectDelay) * time.Millisecond)
				if err := e.connect(); err != nil {
					continue
				}
				reader = bufio.NewReader(e.conn)
			} else {
				return
			}
		}

		// Read line (terminated by \n)
		data, err := reader.ReadString('\n')
		if err != nil {
			e.mu.Lock()
			e.connected = false
			if e.conn != nil {
				e.conn.Close()
			}
			e.mu.Unlock()
			continue
		}

		// Create message
		msg := node.Message{
			Payload: map[string]interface{}{
				"data": data,
			},
		}

		// Send to output channel (non-blocking)
		select {
		case e.outputChan <- msg:
		default:
		}
	}
}

// send send data
func (e *TCPClientExecutor) send(data interface{}) error {
	if !e.isConnected() {
		return fmt.Errorf("not connected")
	}

	e.mu.RLock()
	conn := e.conn
	e.mu.RUnlock()

	var bytes []byte
	switch v := data.(type) {
	case string:
		bytes = []byte(v)
	case []byte:
		bytes = v
	default:
		var err error
		bytes, err = json.Marshal(v)
		if err != nil {
			return fmt.Errorf("failed to marshal data: %w", err)
		}
	}

	// Add newline if not present
	if len(bytes) == 0 || bytes[len(bytes)-1] != '\n' {
		bytes = append(bytes, '\n')
	}

	_, err := conn.Write(bytes)
	return err
}

// isConnected check connection status
func (e *TCPClientExecutor) isConnected() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.connected && e.conn != nil
}

// Cleanup cleanup resources
func (e *TCPClientExecutor) Cleanup() error {
	close(e.stopChan)

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.conn != nil {
		e.conn.Close()
		e.conn = nil
		e.connected = false
	}

	close(e.outputChan)
	return nil
}
