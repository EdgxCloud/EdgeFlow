package websocket

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gofiber/websocket/v2"
)

// MessageType defines the type of WebSocket message
type MessageType string

const (
	MessageTypeFlowStatus   MessageType = "flow_status"
	MessageTypeNodeStatus   MessageType = "node_status"
	MessageTypeModuleStatus MessageType = "module_status"
	MessageTypeExecution    MessageType = "execution"
	MessageTypeLog          MessageType = "log"
	MessageTypeNotification MessageType = "notification"
)

// Message represents a WebSocket message
type Message struct {
	Type      MessageType            `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Data      map[string]interface{} `json:"data"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID   string
	Conn *websocket.Conn
	Send chan Message
	Hub  *Hub
}

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	clients    map[string]*Client
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[string]*Client),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// registerClient adds a new client to the hub
func (h *Hub) registerClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[client.ID] = client
}

// unregisterClient removes a client from the hub
func (h *Hub) unregisterClient(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if _, ok := h.clients[client.ID]; ok {
		delete(h.clients, client.ID)
		close(client.Send)
	}
}

// broadcastMessage sends a message to all connected clients
func (h *Hub) broadcastMessage(message Message) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for _, client := range h.clients {
		select {
		case client.Send <- message:
		default:
			// Client's send channel is full, skip
		}
	}
}

// Broadcast sends a message to all connected clients
func (h *Hub) Broadcast(messageType MessageType, data map[string]interface{}) {
	message := Message{
		Type:      messageType,
		Timestamp: time.Now(),
		Data:      data,
	}
	h.broadcast <- message
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// HandleWebSocket handles new WebSocket connections
func (h *Hub) HandleWebSocket(c *websocket.Conn) {
	client := &Client{
		ID:   generateClientID(),
		Conn: c,
		Send: make(chan Message, 256),
		Hub:  h,
	}

	h.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	client.readPump()
}

// readPump reads messages from the WebSocket connection
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	for {
		messageType, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				// Log error
			}
			break
		}

		if messageType == websocket.TextMessage {
			// Handle incoming messages (optional)
			var msg map[string]interface{}
			if err := json.Unmarshal(message, &msg); err == nil {
				// Process message if needed
			}
		}
	}
}

// writePump writes messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// Channel closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			// Send message
			data, err := json.Marshal(message)
			if err != nil {
				continue
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			// Send ping
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// Helper function to generate a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}
