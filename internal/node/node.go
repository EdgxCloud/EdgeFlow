package node

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// MessageType defines the type of message being passed between nodes
type MessageType string

const (
	MessageTypeData  MessageType = "data"
	MessageTypeError MessageType = "error"
	MessageTypeEvent MessageType = "event"
)

// Message represents data flowing between nodes
type Message struct {
	Type    MessageType            `json:"type"`
	Payload map[string]interface{} `json:"payload"`
	Topic   string                 `json:"topic,omitempty"`
	Error   error                  `json:"error,omitempty"`
}

// NodeType defines the category of a node
type NodeType string

const (
	NodeTypeInput      NodeType = "input"
	NodeTypeOutput     NodeType = "output"
	NodeTypeProcessing NodeType = "processing"
	NodeTypeFunction   NodeType = "function"
)

// NodeStatus represents the current state of a node
type NodeStatus string

const (
	NodeStatusIdle    NodeStatus = "idle"
	NodeStatusRunning NodeStatus = "running"
	NodeStatusError   NodeStatus = "error"
)

// Node represents a single processing unit in a flow
type Node struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Category    NodeType               `json:"category"`
	Config      map[string]interface{} `json:"config"`
	Inputs      []string               `json:"inputs"`
	Outputs     []string               `json:"outputs"`
	Status      NodeStatus             `json:"status"`
	mu          sync.RWMutex
	executor    Executor
	inputChan   chan Message
	outputChans []chan Message
	ctx         context.Context
	cancel      context.CancelFunc
}

// Executor defines the interface for node execution logic
type Executor interface {
	Execute(ctx context.Context, msg Message) (Message, error)
	Init(config map[string]interface{}) error
	Cleanup() error
}

// NewNode creates a new node instance
func NewNode(nodeType, name string, category NodeType, executor Executor) *Node {
	return &Node{
		ID:          uuid.New().String(),
		Type:        nodeType,
		Name:        name,
		Category:    category,
		Config:      make(map[string]interface{}),
		Inputs:      []string{},
		Outputs:     []string{},
		Status:      NodeStatusIdle,
		executor:    executor,
		inputChan:   make(chan Message, 100),
		outputChans: []chan Message{},
	}
}

// Start begins processing messages
func (n *Node) Start(ctx context.Context) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Status == NodeStatusRunning {
		return fmt.Errorf("node %s is already running", n.ID)
	}

	n.ctx, n.cancel = context.WithCancel(ctx)
	n.Status = NodeStatusRunning

	// Initialize executor
	if err := n.executor.Init(n.Config); err != nil {
		n.Status = NodeStatusError
		return fmt.Errorf("failed to initialize node: %w", err)
	}

	// Start message processing goroutine
	go n.process()

	return nil
}

// Stop halts message processing
func (n *Node) Stop() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.Status != NodeStatusRunning {
		return nil
	}

	if n.cancel != nil {
		n.cancel()
	}

	n.Status = NodeStatusIdle

	// Cleanup executor
	return n.executor.Cleanup()
}

// Send sends a message to this node
func (n *Node) Send(msg Message) error {
	select {
	case n.inputChan <- msg:
		return nil
	case <-n.ctx.Done():
		return fmt.Errorf("node %s is stopped", n.ID)
	default:
		return fmt.Errorf("node %s input buffer is full", n.ID)
	}
}

// Connect connects this node's output to another node's input
func (n *Node) Connect(targetNode *Node) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.outputChans = append(n.outputChans, targetNode.inputChan)
	n.Outputs = append(n.Outputs, targetNode.ID)
	targetNode.Inputs = append(targetNode.Inputs, n.ID)
}

// process handles incoming messages
func (n *Node) process() {
	for {
		select {
		case <-n.ctx.Done():
			return
		case msg := <-n.inputChan:
			n.handleMessage(msg)
		}
	}
}

// handleMessage processes a single message
func (n *Node) handleMessage(msg Message) {
	// Execute node logic
	result, err := n.executor.Execute(n.ctx, msg)

	if err != nil {
		n.mu.Lock()
		n.Status = NodeStatusError
		n.mu.Unlock()

		// Send error message to outputs
		errorMsg := Message{
			Type:  MessageTypeError,
			Error: err,
			Payload: map[string]interface{}{
				"node_id": n.ID,
				"error":   err.Error(),
			},
		}
		n.sendToOutputs(errorMsg)
		return
	}

	// Send result to connected nodes
	n.sendToOutputs(result)
}

// sendToOutputs broadcasts a message to all connected output nodes
func (n *Node) sendToOutputs(msg Message) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	for _, outChan := range n.outputChans {
		select {
		case outChan <- msg:
		case <-n.ctx.Done():
			return
		default:
			// Output buffer full, skip
		}
	}
}

// GetStatus returns the current node status
func (n *Node) GetStatus() NodeStatus {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.Status
}

// UpdateConfig updates the node configuration
func (n *Node) UpdateConfig(config map[string]interface{}) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Config = config

	// Re-initialize if running
	if n.Status == NodeStatusRunning {
		return n.executor.Init(config)
	}

	return nil
}
