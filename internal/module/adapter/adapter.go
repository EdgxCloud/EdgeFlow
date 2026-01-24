// Package adapter provides adapters for running imported modules
// Supports Node-RED JavaScript nodes and n8n TypeScript nodes
package adapter

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgeflow/edgeflow/internal/module/parser"
	"github.com/edgeflow/edgeflow/internal/node"
)

// Adapter interface for module runtime adapters
type Adapter interface {
	// Format returns the module format this adapter handles
	Format() parser.ModuleFormat

	// CanExecute checks if this adapter can execute the given node
	CanExecute(nodeInfo *parser.NodeInfo) bool

	// CreateExecutor creates a node executor from node info
	CreateExecutor(nodeInfo *parser.NodeInfo, sourceCode string) (node.Executor, error)

	// Cleanup releases adapter resources
	Cleanup() error
}

// AdapterRegistry manages module adapters
type AdapterRegistry struct {
	mu       sync.RWMutex
	adapters map[parser.ModuleFormat]Adapter
}

var (
	globalAdapterRegistry *AdapterRegistry
	adapterOnce           sync.Once
)

// GetAdapterRegistry returns the global adapter registry
func GetAdapterRegistry() *AdapterRegistry {
	adapterOnce.Do(func() {
		globalAdapterRegistry = &AdapterRegistry{
			adapters: make(map[parser.ModuleFormat]Adapter),
		}
	})
	return globalAdapterRegistry
}

// Register registers an adapter for a module format
func (r *AdapterRegistry) Register(adapter Adapter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.adapters[adapter.Format()] = adapter
}

// Get returns the adapter for a module format
func (r *AdapterRegistry) Get(format parser.ModuleFormat) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapter, ok := r.adapters[format]
	return adapter, ok
}

// GetAll returns all registered adapters
func (r *AdapterRegistry) GetAll() []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	adapters := make([]Adapter, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		adapters = append(adapters, adapter)
	}
	return adapters
}

// BaseExecutor provides common functionality for adapted node executors
type BaseExecutor struct {
	NodeInfo   *parser.NodeInfo
	SourceCode string
	Config     map[string]interface{}
}

// Init initializes the executor with configuration
func (e *BaseExecutor) Init(config map[string]interface{}) error {
	e.Config = config
	return nil
}

// Cleanup releases executor resources
func (e *BaseExecutor) Cleanup() error {
	return nil
}

// Message represents a flow message compatible with Node-RED/n8n
type Message struct {
	Payload   interface{}            `json:"payload"`
	Topic     string                 `json:"topic,omitempty"`
	MsgID     string                 `json:"_msgid,omitempty"`
	Parts     *MessageParts          `json:"parts,omitempty"`
	Complete  bool                   `json:"complete,omitempty"`
	Error     *MessageError          `json:"error,omitempty"`
	Extra     map[string]interface{} `json:"-"`
}

// MessageParts contains split/join message parts info
type MessageParts struct {
	ID    string `json:"id"`
	Index int    `json:"index"`
	Count int    `json:"count"`
	Type  string `json:"type"`
}

// MessageError contains error information
type MessageError struct {
	Message    string `json:"message"`
	Source     string `json:"source"`
	SourceNode string `json:"source_node"`
}

// ToNodeMessage converts adapter Message to internal node.Message
func (m *Message) ToNodeMessage() node.Message {
	payload := make(map[string]interface{})

	// Set payload
	if m.Payload != nil {
		if mp, ok := m.Payload.(map[string]interface{}); ok {
			payload = mp
		} else {
			payload["value"] = m.Payload
		}
	}

	// Copy extra properties
	for k, v := range m.Extra {
		payload[k] = v
	}

	return node.Message{
		Payload: payload,
		Topic:   m.Topic,
	}
}

// FromNodeMessage converts internal node.Message to adapter Message
func FromNodeMessage(msg node.Message) *Message {
	m := &Message{
		Topic: msg.Topic,
		Extra: make(map[string]interface{}),
	}

	// Extract payload
	if val, ok := msg.Payload["value"]; ok {
		m.Payload = val
	} else {
		m.Payload = msg.Payload
	}

	// Copy other properties
	for k, v := range msg.Payload {
		if k != "value" {
			m.Extra[k] = v
		}
	}

	return m
}

// ExecutionContext provides context for node execution
type ExecutionContext struct {
	context.Context
	NodeID   string
	FlowID   string
	NodeName string
	Send     func(msg *Message, output int) error
	Done     func(err error)
	Status   func(status NodeStatus)
	Log      func(level, message string)
}

// NodeStatus represents node status for UI
type NodeStatus struct {
	Fill  string `json:"fill"`  // "red", "green", "yellow", "blue", "grey"
	Shape string `json:"shape"` // "ring", "dot"
	Text  string `json:"text"`
}

// CreateExecutionContext creates a new execution context
func CreateExecutionContext(ctx context.Context, nodeID, flowID, nodeName string) *ExecutionContext {
	return &ExecutionContext{
		Context:  ctx,
		NodeID:   nodeID,
		FlowID:   flowID,
		NodeName: nodeName,
		Send: func(msg *Message, output int) error {
			return nil // Default no-op
		},
		Done: func(err error) {
			// Default no-op
		},
		Status: func(status NodeStatus) {
			// Default no-op
		},
		Log: func(level, message string) {
			// Default no-op
		},
	}
}

// NodeRedContext provides Node-RED compatible context
type NodeRedContext struct {
	*ExecutionContext
	node *NodeRedNodeWrapper
}

// NodeRedNodeWrapper wraps node info for Node-RED compatibility
type NodeRedNodeWrapper struct {
	ID     string                 `json:"id"`
	Type   string                 `json:"type"`
	Name   string                 `json:"name"`
	Config map[string]interface{} `json:"config"`
}

// NewNodeRedContext creates a Node-RED compatible context
func NewNodeRedContext(execCtx *ExecutionContext, nodeInfo *parser.NodeInfo) *NodeRedContext {
	return &NodeRedContext{
		ExecutionContext: execCtx,
		node: &NodeRedNodeWrapper{
			ID:     execCtx.NodeID,
			Type:   nodeInfo.Type,
			Name:   nodeInfo.Name,
			Config: nodeInfo.Config,
		},
	}
}

// Send sends a message to output
func (c *NodeRedContext) Send(msg interface{}) error {
	if m, ok := msg.(*Message); ok {
		return c.ExecutionContext.Send(m, 0)
	}
	return fmt.Errorf("invalid message type")
}

// SendToOutput sends a message to specific output
func (c *NodeRedContext) SendToOutput(msg interface{}, output int) error {
	if m, ok := msg.(*Message); ok {
		return c.ExecutionContext.Send(m, output)
	}
	return fmt.Errorf("invalid message type")
}

// Error reports an error
func (c *NodeRedContext) Error(err error) {
	c.ExecutionContext.Done(err)
}

// Warn logs a warning
func (c *NodeRedContext) Warn(message string) {
	c.ExecutionContext.Log("warn", message)
}

// Debug logs a debug message
func (c *NodeRedContext) Debug(message string) {
	c.ExecutionContext.Log("debug", message)
}

// Status sets node status
func (c *NodeRedContext) Status(status NodeStatus) {
	c.ExecutionContext.Status(status)
}

// N8NContext provides n8n compatible context
type N8NContext struct {
	*ExecutionContext
	nodeInfo *parser.NodeInfo
}

// NewN8NContext creates an n8n compatible context
func NewN8NContext(execCtx *ExecutionContext, nodeInfo *parser.NodeInfo) *N8NContext {
	return &N8NContext{
		ExecutionContext: execCtx,
		nodeInfo:         nodeInfo,
	}
}

// GetInputData returns input data for n8n node
func (c *N8NContext) GetInputData() []map[string]interface{} {
	return []map[string]interface{}{}
}

// GetNodeParameter returns a node parameter value
func (c *N8NContext) GetNodeParameter(name string, defaultValue interface{}) interface{} {
	if c.nodeInfo.Config != nil {
		if val, ok := c.nodeInfo.Config[name]; ok {
			return val
		}
	}
	return defaultValue
}
