package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/google/uuid"
)

// FlowStatus represents the current state of a flow
type FlowStatus string

const (
	FlowStatusIdle    FlowStatus = "idle"
	FlowStatusRunning FlowStatus = "running"
	FlowStatusStopped FlowStatus = "stopped"
	FlowStatusError   FlowStatus = "error"
)

// Flow represents a complete workflow containing multiple nodes
type Flow struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Status      FlowStatus             `json:"status"`
	Nodes       map[string]*node.Node  `json:"nodes"`
	Connections []Connection           `json:"connections"`
	Config      map[string]interface{} `json:"config"`
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	onExecution node.ExecutionCallback
}

// Connection represents a link between two nodes
type Connection struct {
	ID       string `json:"id"`
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
}

// NewFlow creates a new flow instance
func NewFlow(name, description string) *Flow {
	return &Flow{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		Status:      FlowStatusIdle,
		Nodes:       make(map[string]*node.Node),
		Connections: []Connection{},
		Config:      make(map[string]interface{}),
	}
}

// AddNode adds a node to the flow
func (f *Flow) AddNode(n *node.Node) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if _, exists := f.Nodes[n.ID]; exists {
		return fmt.Errorf("node with ID %s already exists", n.ID)
	}

	f.Nodes[n.ID] = n
	return nil
}

// RemoveNode removes a node from the flow
func (f *Flow) RemoveNode(nodeID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	node, exists := f.Nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	// Stop the node if running
	if err := node.Stop(); err != nil {
		return fmt.Errorf("failed to stop node: %w", err)
	}

	// Remove connections involving this node
	newConnections := []Connection{}
	for _, conn := range f.Connections {
		if conn.SourceID != nodeID && conn.TargetID != nodeID {
			newConnections = append(newConnections, conn)
		}
	}
	f.Connections = newConnections

	delete(f.Nodes, nodeID)
	return nil
}

// Connect creates a connection between two nodes
func (f *Flow) Connect(sourceID, targetID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	sourceNode, sourceExists := f.Nodes[sourceID]
	targetNode, targetExists := f.Nodes[targetID]

	if !sourceExists {
		return fmt.Errorf("source node %s not found", sourceID)
	}
	if !targetExists {
		return fmt.Errorf("target node %s not found", targetID)
	}

	// Create connection
	sourceNode.Connect(targetNode)

	// Record connection
	conn := Connection{
		ID:       uuid.New().String(),
		SourceID: sourceID,
		TargetID: targetID,
	}
	f.Connections = append(f.Connections, conn)

	return nil
}

// Disconnect removes a connection between two nodes
func (f *Flow) Disconnect(connectionID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	for i, conn := range f.Connections {
		if conn.ID == connectionID {
			// Remove connection from list
			f.Connections = append(f.Connections[:i], f.Connections[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("connection %s not found", connectionID)
}

// SetExecutionCallback sets a callback for node execution events
func (f *Flow) SetExecutionCallback(cb node.ExecutionCallback) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.onExecution = cb
}

// Start begins executing the flow
func (f *Flow) Start(ctx context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Status == FlowStatusRunning {
		return fmt.Errorf("flow %s is already running", f.ID)
	}

	f.ctx, f.cancel = context.WithCancel(ctx)
	f.Status = FlowStatusRunning

	// Start all nodes and set execution callback
	for _, n := range f.Nodes {
		if f.onExecution != nil {
			n.SetExecutionCallback(f.onExecution)
		}
		if err := n.Start(f.ctx); err != nil {
			f.Status = FlowStatusError
			return fmt.Errorf("failed to start node %s: %w", n.ID, err)
		}
	}

	return nil
}

// Stop halts the flow execution
func (f *Flow) Stop() error {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.Status != FlowStatusRunning {
		return nil
	}

	if f.cancel != nil {
		f.cancel()
	}

	// Clear execution callbacks before stopping to prevent stale broadcasts
	for _, node := range f.Nodes {
		node.SetExecutionCallback(nil)
	}

	// Stop all nodes
	var stopErrors []error
	for _, node := range f.Nodes {
		if err := node.Stop(); err != nil {
			stopErrors = append(stopErrors, err)
		}
	}

	f.Status = FlowStatusStopped

	if len(stopErrors) > 0 {
		return fmt.Errorf("errors stopping nodes: %v", stopErrors)
	}

	return nil
}

// GetNode retrieves a node by ID
func (f *Flow) GetNode(nodeID string) (*node.Node, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	node, exists := f.Nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	return node, nil
}

// GetStatus returns the current flow status
func (f *Flow) GetStatus() FlowStatus {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.Status
}

// Validate checks if the flow is valid
func (f *Flow) Validate() error {
	f.mu.RLock()
	defer f.mu.RUnlock()

	if len(f.Nodes) == 0 {
		return fmt.Errorf("flow has no nodes")
	}

	// Check for circular dependencies (simplified check)
	// TODO: Implement proper cycle detection

	return nil
}
