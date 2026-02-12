package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/gorilla/websocket"
)

// wsServerConnections manages WebSocket server connections per path
var (
	wsServerConns = make(map[string]map[*websocket.Conn]bool)
	wsServerMu    sync.RWMutex
)

func addWSServerConn(path string, conn *websocket.Conn) {
	wsServerMu.Lock()
	defer wsServerMu.Unlock()
	if wsServerConns[path] == nil {
		wsServerConns[path] = make(map[*websocket.Conn]bool)
	}
	wsServerConns[path][conn] = true
}

func removeWSServerConn(path string, conn *websocket.Conn) {
	wsServerMu.Lock()
	defer wsServerMu.Unlock()
	if conns, ok := wsServerConns[path]; ok {
		delete(conns, conn)
		if len(conns) == 0 {
			delete(wsServerConns, path)
		}
	}
}

func broadcastWSServer(path string, data []byte) {
	wsServerMu.RLock()
	conns := make([]*websocket.Conn, 0)
	if cs, ok := wsServerConns[path]; ok {
		for c := range cs {
			conns = append(conns, c)
		}
	}
	wsServerMu.RUnlock()

	for _, c := range conns {
		if err := c.WriteMessage(websocket.TextMessage, data); err != nil {
			removeWSServerConn(path, c)
			c.Close()
		}
	}
}

// ============================================================
// WebSocket In (Server) - Accepts connections & receives messages
// ============================================================

// WebSocketInExecutor creates a WebSocket server endpoint that receives messages
type WebSocketInExecutor struct {
	path       string
	outputChan chan node.Message
	registered bool
	mu         sync.RWMutex
}

// wsInRegistry global registry for websocket-in endpoints
var (
	wsInRegistry = make(map[string]*WebSocketInExecutor)
	wsInMu       sync.RWMutex
)

// NewWebSocketInExecutor creates a new WebSocketInExecutor
func NewWebSocketInExecutor() node.Executor {
	return &WebSocketInExecutor{
		outputChan: make(chan node.Message, 100),
	}
}

func (e *WebSocketInExecutor) Init(config map[string]interface{}) error {
	if p, ok := config["path"].(string); ok && p != "" {
		e.path = p
	} else {
		e.path = "/ws/nodes"
	}
	return nil
}

func (e *WebSocketInExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if !e.registered {
		wsInMu.Lock()
		if _, exists := wsInRegistry[e.path]; exists {
			wsInMu.Unlock()
			return node.Message{}, fmt.Errorf("websocket-in path %s already registered", e.path)
		}
		wsInRegistry[e.path] = e
		wsInMu.Unlock()
		e.registered = true
	}

	select {
	case <-ctx.Done():
		return node.Message{}, ctx.Err()
	case incoming := <-e.outputChan:
		return incoming, nil
	}
}

// HandleWSConnection handles an incoming WebSocket connection for this endpoint
func (e *WebSocketInExecutor) HandleWSConnection(conn *websocket.Conn) {
	addWSServerConn(e.path, conn)
	defer func() {
		removeWSServerConn(e.path, conn)
		conn.Close()
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var payload interface{}
		if messageType == websocket.TextMessage {
			if err := json.Unmarshal(message, &payload); err != nil {
				payload = string(message)
			}
		} else {
			payload = message
		}

		msg := node.Message{
			Type: node.MessageTypeData,
			Payload: map[string]interface{}{
				"type":    messageType,
				"payload": payload,
				"path":    e.path,
			},
		}

		select {
		case e.outputChan <- msg:
		default:
		}
	}
}

func (e *WebSocketInExecutor) Cleanup() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.registered {
		wsInMu.Lock()
		delete(wsInRegistry, e.path)
		wsInMu.Unlock()
		e.registered = false
	}
	close(e.outputChan)
	return nil
}

// GetWSInHandler retrieves the websocket-in handler for a given path
func GetWSInHandler(path string) (*WebSocketInExecutor, bool) {
	wsInMu.RLock()
	defer wsInMu.RUnlock()
	ex, ok := wsInRegistry[path]
	return ex, ok
}

// ============================================================
// WebSocket Out (Server) - Sends messages to connected clients
// ============================================================

// WebSocketOutExecutor sends messages to all connected WebSocket clients on a path
type WebSocketOutExecutor struct {
	path string
}

// NewWebSocketOutExecutor creates a new WebSocketOutExecutor
func NewWebSocketOutExecutor() node.Executor {
	return &WebSocketOutExecutor{}
}

func (e *WebSocketOutExecutor) Init(config map[string]interface{}) error {
	if p, ok := config["path"].(string); ok && p != "" {
		e.path = p
	} else {
		e.path = "/ws/nodes"
	}
	return nil
}

func (e *WebSocketOutExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	var data []byte
	var err error

	if s, ok := msg.Payload["payload"].(string); ok {
		data = []byte(s)
	} else {
		data, err = json.Marshal(msg.Payload)
		if err != nil {
			return node.Message{}, fmt.Errorf("failed to marshal message: %w", err)
		}
	}

	broadcastWSServer(e.path, data)

	connCount := 0
	wsServerMu.RLock()
	if conns, ok := wsServerConns[e.path]; ok {
		connCount = len(conns)
	}
	wsServerMu.RUnlock()

	return node.Message{
		Type: node.MessageTypeData,
		Payload: map[string]interface{}{
			"sent":        true,
			"path":        e.path,
			"connections": connCount,
		},
	}, nil
}

func (e *WebSocketOutExecutor) Cleanup() error {
	return nil
}
