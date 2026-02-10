package network

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"

	"github.com/edgeflow/edgeflow/internal/node"
)

// UDPConfig UDP node
type UDPConfig struct {
	Mode       string `json:"mode"`       // listen or send
	Host       string `json:"host"`       // Host address
	Port       int    `json:"port"`       // Port number
	BufferSize int    `json:"bufferSize"` // Buffer size for receiving
}

// UDPExecutor UDP node executor
type UDPExecutor struct {
	config     UDPConfig
	conn       *net.UDPConn
	outputChan chan node.Message
	mu         sync.RWMutex
	stopChan   chan struct{}
}

// NewUDPExecutor create UDPExecutor
func NewUDPExecutor() node.Executor {
	return &UDPExecutor{
		outputChan: make(chan node.Message, 100),
		stopChan:   make(chan struct{}),
	}
}

// Init initializes the executor with configuration
func (e *UDPExecutor) Init(config map[string]interface{}) error {
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	var udpConfig UDPConfig
	if err := json.Unmarshal(configJSON, &udpConfig); err != nil {
		return fmt.Errorf("invalid udp config: %w", err)
	}

	// Validate
	if udpConfig.Mode == "" {
		udpConfig.Mode = "listen"
	}
	if udpConfig.Mode != "listen" && udpConfig.Mode != "send" {
		return fmt.Errorf("mode must be 'listen' or 'send'")
	}
	if udpConfig.Port == 0 {
		return fmt.Errorf("port is required")
	}

	// Default values
	if udpConfig.BufferSize == 0 {
		udpConfig.BufferSize = 4096
	}

	e.config = udpConfig
	return nil
}

// Execute execute node
func (e *UDPExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Setup connection if not setup
	if e.conn == nil {
		if err := e.setup(); err != nil {
			return node.Message{}, fmt.Errorf("failed to setup: %w", err)
		}

		// Start read loop for listen mode
		if e.config.Mode == "listen" {
			go e.readLoop()
		}
	}

	// Send mode
	if e.config.Mode == "send" {
		if data, ok := msg.Payload["send"]; ok {
			if err := e.send(data, msg.Payload); err != nil {
				return node.Message{}, fmt.Errorf("failed to send: %w", err)
			}

			return node.Message{
				Payload: map[string]interface{}{
					"sent": true,
				},
			}, nil
		}
		return node.Message{}, fmt.Errorf("no data to send")
	}

	// Listen mode - wait for incoming data
	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case udpMsg := <-e.outputChan:
		return udpMsg, nil
	}
}

// setup setup UDP connection
func (e *UDPExecutor) setup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.conn != nil {
		return nil
	}

	var err error
	var addr *net.UDPAddr

	if e.config.Mode == "listen" {
		// Listen mode
		addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", e.config.Port))
		if err != nil {
			return fmt.Errorf("failed to resolve address: %w", err)
		}

		e.conn, err = net.ListenUDP("udp", addr)
		if err != nil {
			return fmt.Errorf("failed to listen: %w", err)
		}
	} else {
		// Send mode
		if e.config.Host == "" {
			return fmt.Errorf("host is required for send mode")
		}

		addr, err = net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", e.config.Host, e.config.Port))
		if err != nil {
			return fmt.Errorf("failed to resolve address: %w", err)
		}

		e.conn, err = net.DialUDP("udp", nil, addr)
		if err != nil {
			return fmt.Errorf("failed to dial: %w", err)
		}
	}

	return nil
}

// readLoop data read loop (listen mode only)
func (e *UDPExecutor) readLoop() {
	buffer := make([]byte, e.config.BufferSize)

	for {
		select {
		case <-e.stopChan:
			return
		default:
		}

		n, addr, err := e.conn.ReadFromUDP(buffer)
		if err != nil {
			continue
		}

		data := make([]byte, n)
		copy(data, buffer[:n])

		// Create message
		msg := node.Message{
			Payload: map[string]interface{}{
				"data": string(data),
				"from": addr.String(),
				"size": n,
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
func (e *UDPExecutor) send(data interface{}, msgPayload map[string]interface{}) error {
	e.mu.RLock()
	conn := e.conn
	e.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection not established")
	}

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

	// Check if we need to send to a specific address
	if host, ok := msgPayload["host"].(string); ok {
		if port, ok := msgPayload["port"].(float64); ok {
			addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, int(port)))
			if err != nil {
				return fmt.Errorf("failed to resolve address: %w", err)
			}
			_, err = conn.WriteToUDP(bytes, addr)
			return err
		}
	}

	// Send to default address (dial mode)
	_, err := conn.Write(bytes)
	return err
}

// Cleanup cleanup resources
func (e *UDPExecutor) Cleanup() error {
	close(e.stopChan)

	e.mu.Lock()
	defer e.mu.Unlock()

	if e.conn != nil {
		e.conn.Close()
		e.conn = nil
	}

	close(e.outputChan)
	return nil
}
