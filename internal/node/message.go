package node

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// EnhancedMessage represents an enhanced message with Node-RED-like features
type EnhancedMessage struct {
	// Core fields
	Payload interface{}            `json:"payload"`           // Main message data
	Topic   string                 `json:"topic,omitempty"`   // Message identifier/subject
	ID      string                 `json:"_msgid"`            // Unique message ID

	// Sequence tracking
	Parts   *MessageParts          `json:"parts,omitempty"`   // For split/join operations

	// Error handling
	Error   *MessageError          `json:"error,omitempty"`   // Error information

	// HTTP context (for HTTP nodes)
	Req     *http.Request          `json:"-"`                 // HTTP request
	Res     http.ResponseWriter    `json:"-"`                 // HTTP response writer

	// Additional metadata
	Metadata map[string]interface{} `json:"_metadata,omitempty"` // Custom metadata

	// Internal fields
	timestamp time.Time
	mu        sync.RWMutex
}

// MessageParts tracks message sequences for split/join operations
type MessageParts struct {
	ID    string      `json:"id"`              // Sequence identifier
	Index int         `json:"index"`           // Position in sequence (0-based)
	Count int         `json:"count"`           // Total messages in sequence
	Type  string      `json:"type"`            // "array", "object", "string"
	Ch    interface{} `json:"ch,omitempty"`    // Character for string split
	Key   string      `json:"key,omitempty"`   // Key for object split
	Len   int         `json:"len,omitempty"`   // Length for fixed-length splits
}

// MessageError contains detailed error information
type MessageError struct {
	Message string                 `json:"message"`         // Error message
	Source  *ErrorSource           `json:"source"`          // Where error occurred
	Stack   string                 `json:"stack,omitempty"` // Stack trace
	Code    string                 `json:"code,omitempty"`  // Error code
	Level   string                 `json:"level,omitempty"` // "warn", "error", "fatal"
}

// Error implements the error interface
func (e *MessageError) Error() string {
	return e.Message
}

// ErrorSource identifies where an error originated
type ErrorSource struct {
	ID   string `json:"id"`   // Node ID
	Type string `json:"type"` // Node type
	Name string `json:"name"` // Node name
	Count int   `json:"count"` // Number of errors from this node
}

// NewEnhancedMessage creates a new enhanced message
func NewEnhancedMessage(payload interface{}) *EnhancedMessage {
	return &EnhancedMessage{
		Payload:   payload,
		ID:        uuid.New().String(),
		Metadata:  make(map[string]interface{}),
		timestamp: time.Now(),
	}
}

// NewMessageFromLegacy converts a legacy Message to EnhancedMessage
func NewMessageFromLegacy(msg Message) *EnhancedMessage {
	em := &EnhancedMessage{
		Payload:   msg.Payload,
		Topic:     msg.Topic,
		ID:        uuid.New().String(),
		Metadata:  make(map[string]interface{}),
		timestamp: time.Now(),
	}

	// Convert error if present
	if msg.Error != nil {
		em.Error = &MessageError{
			Message: msg.Error.Error(),
		}
	}

	return em
}

// Clone creates a deep copy of the message
func (m *EnhancedMessage) Clone() *EnhancedMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()

	clone := &EnhancedMessage{
		Payload:   cloneInterface(m.Payload),
		Topic:     m.Topic,
		ID:        m.ID, // Keep same ID for clones (Node-RED behavior)
		Req:       m.Req,
		Res:       m.Res,
		Metadata:  make(map[string]interface{}),
		timestamp: m.timestamp,
	}

	// Clone Parts
	if m.Parts != nil {
		clone.Parts = &MessageParts{
			ID:    m.Parts.ID,
			Index: m.Parts.Index,
			Count: m.Parts.Count,
			Type:  m.Parts.Type,
			Ch:    m.Parts.Ch,
			Key:   m.Parts.Key,
			Len:   m.Parts.Len,
		}
	}

	// Clone Error
	if m.Error != nil {
		clone.Error = &MessageError{
			Message: m.Error.Message,
			Stack:   m.Error.Stack,
			Code:    m.Error.Code,
			Level:   m.Error.Level,
		}
		if m.Error.Source != nil {
			clone.Error.Source = &ErrorSource{
				ID:    m.Error.Source.ID,
				Type:  m.Error.Source.Type,
				Name:  m.Error.Source.Name,
				Count: m.Error.Source.Count,
			}
		}
	}

	// Clone Metadata
	for k, v := range m.Metadata {
		clone.Metadata[k] = cloneInterface(v)
	}

	return clone
}

// cloneInterface performs a deep copy of an interface value
func cloneInterface(v interface{}) interface{} {
	if v == nil {
		return nil
	}

	// Use JSON marshal/unmarshal for deep copy
	// This handles maps, slices, and nested structures
	data, err := json.Marshal(v)
	if err != nil {
		return v // Return original on error
	}

	var clone interface{}
	if err := json.Unmarshal(data, &clone); err != nil {
		return v // Return original on error
	}

	return clone
}

// SetPayload safely sets the payload
func (m *EnhancedMessage) SetPayload(payload interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Payload = payload
}

// GetPayload safely retrieves the payload
func (m *EnhancedMessage) GetPayload() interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Payload
}

// SetProperty sets a property on the message payload (if it's a map)
func (m *EnhancedMessage) SetProperty(key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	payloadMap, ok := m.Payload.(map[string]interface{})
	if !ok {
		// Convert payload to map
		payloadMap = make(map[string]interface{})
		m.Payload = payloadMap
	}

	payloadMap[key] = value
	return nil
}

// GetProperty gets a property from the message payload (if it's a map)
func (m *EnhancedMessage) GetProperty(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	payloadMap, ok := m.Payload.(map[string]interface{})
	if !ok {
		return nil, false
	}

	val, exists := payloadMap[key]
	return val, exists
}

// SetError sets an error on the message
func (m *EnhancedMessage) SetError(err error, nodeID, nodeType, nodeName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Error == nil {
		m.Error = &MessageError{
			Message: err.Error(),
			Source: &ErrorSource{
				ID:    nodeID,
				Type:  nodeType,
				Name:  nodeName,
				Count: 1,
			},
			Level: "error",
		}
	} else {
		// Update existing error
		m.Error.Message = err.Error()
		if m.Error.Source != nil {
			m.Error.Source.Count++
		}
	}
}

// HasError returns true if the message has an error
func (m *EnhancedMessage) HasError() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Error != nil
}

// ClearError removes the error from the message
func (m *EnhancedMessage) ClearError() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Error = nil
}

// IsPartOfSequence returns true if this message is part of a sequence
func (m *EnhancedMessage) IsPartOfSequence() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Parts != nil
}

// IsLastInSequence returns true if this is the last message in a sequence
func (m *EnhancedMessage) IsLastInSequence() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Parts != nil && m.Parts.Index == m.Parts.Count-1
}

// IsFirstInSequence returns true if this is the first message in a sequence
func (m *EnhancedMessage) IsFirstInSequence() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Parts != nil && m.Parts.Index == 0
}

// GetSequenceID returns the sequence ID if this message is part of a sequence
func (m *EnhancedMessage) GetSequenceID() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.Parts != nil {
		return m.Parts.ID
	}
	return ""
}

// ToLegacyMessage converts EnhancedMessage back to legacy Message
func (m *EnhancedMessage) ToLegacyMessage() Message {
	m.mu.RLock()
	defer m.mu.RUnlock()

	legacyMsg := Message{
		Topic: m.Topic,
	}

	// Convert payload to map[string]interface{} if possible
	if payloadMap, ok := m.Payload.(map[string]interface{}); ok {
		legacyMsg.Payload = payloadMap
	} else {
		// Wrap non-map payloads
		legacyMsg.Payload = map[string]interface{}{
			"value": m.Payload,
		}
	}

	// Set type based on error presence
	if m.Error != nil {
		legacyMsg.Type = MessageTypeError
		// Convert MessageError to error
		legacyMsg.Error = &MessageError{
			Message: m.Error.Message,
		}
	} else {
		legacyMsg.Type = MessageTypeData
	}

	return legacyMsg
}

// MarshalJSON implements json.Marshaler for custom serialization
func (m *EnhancedMessage) MarshalJSON() ([]byte, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	type Alias EnhancedMessage
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"_timestamp"`
	}{
		Alias:     (*Alias)(m),
		Timestamp: m.timestamp.Format(time.RFC3339),
	})
}

// UnmarshalJSON implements json.Unmarshaler for custom deserialization
func (m *EnhancedMessage) UnmarshalJSON(data []byte) error {
	type Alias EnhancedMessage
	aux := &struct {
		*Alias
		Timestamp string `json:"_timestamp"`
	}{
		Alias: (*Alias)(m),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if aux.Timestamp != "" {
		if t, err := time.Parse(time.RFC3339, aux.Timestamp); err == nil {
			m.timestamp = t
		}
	}

	if m.Metadata == nil {
		m.Metadata = make(map[string]interface{})
	}

	return nil
}

// GetTimestamp returns the message creation timestamp
func (m *EnhancedMessage) GetTimestamp() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.timestamp
}

// Age returns how long ago the message was created
func (m *EnhancedMessage) Age() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return time.Since(m.timestamp)
}
