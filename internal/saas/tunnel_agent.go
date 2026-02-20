package saas

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// TunnelAgent manages the WebSocket connection to SaaS platform
type TunnelAgent struct {
	config         *Config
	conn           *websocket.Conn
	logger         *zap.Logger
	connected      bool
	mu             sync.RWMutex
	stopCh         chan struct{}
	reconnectTimer *time.Timer
	reconnectCount int

	// Command handling
	commandHandler CommandHandler
	pendingCmds    map[string]chan *TunnelMessage
	cmdMu          sync.Mutex

	// Callbacks
	onConnected    func()
	onDisconnected func()
}

// CommandHandler processes commands from SaaS
type CommandHandler interface {
	HandleCommand(cmd *TunnelMessage) (*TunnelMessage, error)
}

// NewTunnelAgent creates a new tunnel agent
func NewTunnelAgent(config *Config, logger *zap.Logger) *TunnelAgent {
	return &TunnelAgent{
		config:      config,
		logger:      logger,
		stopCh:      make(chan struct{}),
		pendingCmds: make(map[string]chan *TunnelMessage),
	}
}

// SetCommandHandler sets the handler for incoming commands
func (t *TunnelAgent) SetCommandHandler(handler CommandHandler) {
	t.commandHandler = handler
}

// SetCallbacks sets connection lifecycle callbacks
func (t *TunnelAgent) SetCallbacks(onConnected, onDisconnected func()) {
	t.onConnected = onConnected
	t.onDisconnected = onDisconnected
}

// Start initiates the connection to SaaS
func (t *TunnelAgent) Start() error {
	if !t.config.Enabled {
		t.logger.Info("SaaS connection disabled")
		return nil
	}

	if err := t.config.Validate(); err != nil {
		return err
	}

	if !t.config.IsProvisioned() {
		return ErrNotProvisioned("device must be provisioned before connecting")
	}

	t.logger.Info("Starting SaaS tunnel agent",
		zap.String("server", t.config.ServerURL),
		zap.String("device_id", t.config.DeviceID))

	return t.connect()
}

// Stop gracefully closes the tunnel connection
func (t *TunnelAgent) Stop() error {
	t.logger.Info("Stopping SaaS tunnel agent")

	close(t.stopCh)

	t.mu.Lock()
	defer t.mu.Unlock()

	if t.reconnectTimer != nil {
		t.reconnectTimer.Stop()
	}

	if t.conn != nil {
		// Send graceful close
		err := t.conn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			t.logger.Warn("Failed to send close message", zap.Error(err))
		}
		t.conn.Close()
		t.conn = nil
	}

	t.connected = false
	return nil
}

// IsConnected returns current connection status
func (t *TunnelAgent) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// connect establishes WebSocket connection and authenticates
func (t *TunnelAgent) connect() error {
	tunnelURL := t.config.TunnelURL()
	t.logger.Info("Connecting to SaaS tunnel", zap.String("url", tunnelURL))

	dialer := websocket.DefaultDialer
	dialer.HandshakeTimeout = 10 * time.Second

	conn, _, err := dialer.Dial(tunnelURL, nil)
	if err != nil {
		return ErrConnectionFailed(err)
	}

	t.mu.Lock()
	t.conn = conn
	t.mu.Unlock()

	// Send connect message with authentication
	connectMsg := &TunnelMessage{
		Type:     "connect",
		DeviceID: t.config.DeviceID,
		APIKey:   t.config.APIKey,
		Version:  "1.0.0", // EdgeFlow agent version
	}

	if err := t.sendMessage(connectMsg); err != nil {
		conn.Close()
		return ErrAuthenticationFailed("failed to send connect message")
	}

	// Wait for connected confirmation (with timeout)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	_, msgBytes, err := conn.ReadMessage()
	if err != nil {
		conn.Close()
		return ErrAuthenticationFailed("no response from server")
	}
	conn.SetReadDeadline(time.Time{}) // Clear deadline

	var response TunnelMessage
	if err := json.Unmarshal(msgBytes, &response); err != nil {
		conn.Close()
		return ErrAuthenticationFailed("invalid server response")
	}

	if response.Type != "connected" {
		conn.Close()
		return ErrAuthenticationFailed("authentication rejected: " + response.Error)
	}

	t.mu.Lock()
	t.connected = true
	t.reconnectCount = 0
	t.mu.Unlock()

	t.logger.Info("Successfully connected to SaaS",
		zap.String("device_id", t.config.DeviceID))

	if t.onConnected != nil {
		t.onConnected()
	}

	// Start message handlers
	go t.readLoop()
	go t.heartbeatLoop()

	return nil
}

// readLoop processes incoming messages
func (t *TunnelAgent) readLoop() {
	defer func() {
		t.handleDisconnect()
	}()

	for {
		select {
		case <-t.stopCh:
			return
		default:
		}

		_, msgBytes, err := t.conn.ReadMessage()
		if err != nil {
			t.logger.Error("Read error", zap.Error(err))
			return
		}

		var msg TunnelMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			t.logger.Error("Failed to parse message", zap.Error(err))
			continue
		}

		t.handleMessage(&msg)
	}
}

// handleMessage routes incoming messages
func (t *TunnelAgent) handleMessage(msg *TunnelMessage) {
	switch msg.Type {
	case "pong":
		// Heartbeat response, nothing to do

	case "command":
		go t.handleCommand(msg)

	case "response":
		// Response to a command we sent
		t.cmdMu.Lock()
		if ch, ok := t.pendingCmds[msg.ID]; ok {
			delete(t.pendingCmds, msg.ID)
			t.cmdMu.Unlock()
			select {
			case ch <- msg:
			default:
			}
		} else {
			t.cmdMu.Unlock()
		}

	default:
		t.logger.Warn("Unknown message type", zap.String("type", msg.Type))
	}
}

// handleCommand processes a command from SaaS
func (t *TunnelAgent) handleCommand(msg *TunnelMessage) {
	if t.commandHandler == nil {
		t.logger.Warn("No command handler registered, ignoring command",
			zap.String("action", msg.Action))
		return
	}

	t.logger.Info("Received command",
		zap.String("id", msg.ID),
		zap.String("action", msg.Action))

	response, err := t.commandHandler.HandleCommand(msg)
	if err != nil {
		t.logger.Error("Command execution failed",
			zap.String("id", msg.ID),
			zap.String("action", msg.Action),
			zap.Error(err))

		response = &TunnelMessage{
			Type:   "response",
			ID:     msg.ID,
			Status: "error",
			Error:  err.Error(),
		}
	} else {
		if response == nil {
			response = &TunnelMessage{
				Type:   "response",
				ID:     msg.ID,
				Status: "success",
			}
		} else {
			response.Type = "response"
			response.ID = msg.ID
			if response.Status == "" {
				response.Status = "success"
			}
		}
	}

	if err := t.sendMessage(response); err != nil {
		t.logger.Error("Failed to send command response",
			zap.String("id", msg.ID),
			zap.Error(err))
	}
}

// heartbeatLoop sends periodic ping messages
func (t *TunnelAgent) heartbeatLoop() {
	ticker := time.NewTicker(t.config.HeartbeatInterval)
	defer ticker.Stop()

	for {
		select {
		case <-t.stopCh:
			return
		case <-ticker.C:
			if !t.IsConnected() {
				return
			}

			pingMsg := &TunnelMessage{Type: "ping"}
			if err := t.sendMessage(pingMsg); err != nil {
				t.logger.Error("Failed to send heartbeat", zap.Error(err))
				return
			}
		}
	}
}

// handleDisconnect handles connection loss and reconnection
func (t *TunnelAgent) handleDisconnect() {
	t.mu.Lock()
	wasConnected := t.connected
	t.connected = false
	if t.conn != nil {
		t.conn.Close()
		t.conn = nil
	}
	t.mu.Unlock()

	if wasConnected {
		t.logger.Warn("Disconnected from SaaS")
		if t.onDisconnected != nil {
			t.onDisconnected()
		}
	}

	// Attempt reconnection
	select {
	case <-t.stopCh:
		return
	default:
	}

	t.reconnect()
}

// reconnect attempts to re-establish connection
func (t *TunnelAgent) reconnect() {
	t.mu.Lock()
	t.reconnectCount++
	count := t.reconnectCount
	t.mu.Unlock()

	if count > t.config.MaxReconnectAttempts {
		t.logger.Error("Max reconnect attempts reached, giving up",
			zap.Int("attempts", count))
		return
	}

	// Exponential backoff
	delay := time.Duration(count) * 5 * time.Second
	if delay > 60*time.Second {
		delay = 60 * time.Second
	}

	t.logger.Info("Attempting reconnection",
		zap.Int("attempt", count),
		zap.Duration("delay", delay))

	t.reconnectTimer = time.AfterFunc(delay, func() {
		if err := t.connect(); err != nil {
			t.logger.Error("Reconnection failed", zap.Error(err))
			t.handleDisconnect()
		}
	})
}

// sendMessage sends a message over the WebSocket
func (t *TunnelAgent) sendMessage(msg *TunnelMessage) error {
	t.mu.RLock()
	conn := t.conn
	t.mu.RUnlock()

	if conn == nil {
		return fmt.Errorf("not connected")
	}

	msg.Timestamp = time.Now()
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return conn.WriteMessage(websocket.TextMessage, msgBytes)
}

// SendCommand sends a command to SaaS and waits for response
func (t *TunnelAgent) SendCommand(action string, payload map[string]interface{}, timeout time.Duration) (*TunnelMessage, error) {
	if !t.IsConnected() {
		return nil, fmt.Errorf("not connected")
	}

	cmdID := fmt.Sprintf("cmd-%d", time.Now().UnixNano())
	msg := &TunnelMessage{
		Type:    "command",
		ID:      cmdID,
		Action:  action,
		Payload: payload,
	}

	// Register pending command
	respCh := make(chan *TunnelMessage, 1)
	t.cmdMu.Lock()
	t.pendingCmds[cmdID] = respCh
	t.cmdMu.Unlock()

	// Send command
	if err := t.sendMessage(msg); err != nil {
		t.cmdMu.Lock()
		delete(t.pendingCmds, cmdID)
		t.cmdMu.Unlock()
		return nil, err
	}

	// Wait for response
	select {
	case resp := <-respCh:
		return resp, nil
	case <-time.After(timeout):
		t.cmdMu.Lock()
		delete(t.pendingCmds, cmdID)
		t.cmdMu.Unlock()
		return nil, ErrCommandTimeout(cmdID)
	}
}
