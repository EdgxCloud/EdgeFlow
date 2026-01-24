package subflow

import (
	"context"
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Message represents a message flowing through the subflow
type Message struct {
	Payload  any            `json:"payload"`
	Topic    string         `json:"topic,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Context  MessageContext `json:"context,omitempty"`
}

// MessageContext holds contextual information for message routing
type MessageContext struct {
	FlowID       string         `json:"flowId,omitempty"`
	SubflowID    string         `json:"subflowId,omitempty"`
	InstanceID   string         `json:"instanceId,omitempty"`
	SourceNodeID string         `json:"sourceNodeId,omitempty"`
	SourcePort   int            `json:"sourcePort,omitempty"`
	Variables    map[string]any `json:"variables,omitempty"`
}

// NodeExecutor defines the interface for executing nodes within a subflow
type NodeExecutor interface {
	Execute(ctx context.Context, nodeID string, config map[string]any, msg *Message) ([]*Message, error)
}

// Executor executes subflow instances
type Executor struct {
	mu            sync.RWMutex
	registry      *Registry
	nodeExecutor  NodeExecutor
	activeFlows   map[string]*FlowExecution
	messageQueues map[string]chan *Message
}

// FlowExecution represents an active subflow execution
type FlowExecution struct {
	InstanceID     string
	SubflowID      string
	Definition     *SubflowDefinition
	Instance       *SubflowInstance
	Context        context.Context
	Cancel         context.CancelFunc
	NodeStates     map[string]any
	PendingOutputs map[int][]*Message // Pending messages for each output port
	mu             sync.Mutex
}

// NewExecutor creates a new subflow executor
func NewExecutor(registry *Registry, nodeExecutor NodeExecutor) *Executor {
	return &Executor{
		registry:      registry,
		nodeExecutor:  nodeExecutor,
		activeFlows:   make(map[string]*FlowExecution),
		messageQueues: make(map[string]chan *Message),
	}
}

// Execute executes a subflow instance with an input message
func (e *Executor) Execute(ctx context.Context, instanceID string, inputPort int, msg *Message) ([]*Message, error) {
	instance, err := e.registry.GetInstance(instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance: %w", err)
	}

	definition, err := e.registry.GetDefinition(instance.SubflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get definition: %w", err)
	}

	// Validate input port
	inputPortDef := definition.GetInputPort(inputPort)
	if inputPortDef == nil {
		return nil, fmt.Errorf("invalid input port: %d", inputPort)
	}

	// Create or get flow execution
	flowExec, err := e.getOrCreateFlowExecution(ctx, instance, definition)
	if err != nil {
		return nil, fmt.Errorf("failed to create flow execution: %w", err)
	}

	// Set message context
	if msg.Context.Variables == nil {
		msg.Context.Variables = make(map[string]any)
	}
	msg.Context.SubflowID = instance.SubflowID
	msg.Context.InstanceID = instanceID

	// Apply instance config to message context
	for key, value := range instance.Config {
		msg.Context.Variables[key] = value
	}

	// Apply environment variables
	for key, value := range instance.Env {
		msg.Context.Variables[key] = value
	}

	// Route message to nodes connected to input port
	outputs, err := e.routeFromInputPort(flowExec, inputPort, msg)
	if err != nil {
		return nil, fmt.Errorf("failed to route from input port: %w", err)
	}

	return outputs, nil
}

// getOrCreateFlowExecution gets or creates a flow execution instance
func (e *Executor) getOrCreateFlowExecution(ctx context.Context, instance *SubflowInstance, definition *SubflowDefinition) (*FlowExecution, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if exec, ok := e.activeFlows[instance.ID]; ok {
		return exec, nil
	}

	flowCtx, cancel := context.WithCancel(ctx)

	exec := &FlowExecution{
		InstanceID:     instance.ID,
		SubflowID:      instance.SubflowID,
		Definition:     definition,
		Instance:       instance,
		Context:        flowCtx,
		Cancel:         cancel,
		NodeStates:     make(map[string]any),
		PendingOutputs: make(map[int][]*Message),
	}

	e.activeFlows[instance.ID] = exec
	return exec, nil
}

// routeFromInputPort routes a message from an input port to connected nodes
func (e *Executor) routeFromInputPort(flowExec *FlowExecution, inputPort int, msg *Message) ([]*Message, error) {
	// Find connections from this input port
	inputPortID := fmt.Sprintf("port-input-%d", inputPort)
	var outputs []*Message

	for _, conn := range flowExec.Definition.Connections {
		if conn.Source == inputPortID {
			targetNode := flowExec.Definition.GetNode(conn.Target)
			if targetNode == nil {
				continue
			}

			// Execute target node
			nodeOutputs, err := e.executeNode(flowExec, targetNode, msg)
			if err != nil {
				return nil, fmt.Errorf("failed to execute node %s: %w", targetNode.ID, err)
			}

			// Route outputs to next nodes
			for i, nodeOutput := range nodeOutputs {
				if nodeOutput != nil {
					routedOutputs, err := e.routeFromNode(flowExec, targetNode.ID, i, nodeOutput)
					if err != nil {
						return nil, err
					}
					outputs = append(outputs, routedOutputs...)
				}
			}
		}
	}

	return outputs, nil
}

// executeNode executes a single node within the subflow
func (e *Executor) executeNode(flowExec *FlowExecution, node *NodeDefinition, msg *Message) ([]*Message, error) {
	if e.nodeExecutor == nil {
		return nil, fmt.Errorf("node executor not configured")
	}

	// Set node context
	msg.Context.SourceNodeID = node.ID

	// Execute node
	outputs, err := e.nodeExecutor.Execute(flowExec.Context, node.ID, node.Config, msg)
	if err != nil {
		return nil, fmt.Errorf("node execution failed: %w", err)
	}

	// Update node state
	flowExec.mu.Lock()
	flowExec.NodeStates[node.ID] = map[string]any{
		"lastExecution": msg.Metadata,
		"outputCount":   len(outputs),
	}
	flowExec.mu.Unlock()

	return outputs, nil
}

// routeFromNode routes messages from a node's output port to connected nodes or subflow outputs
func (e *Executor) routeFromNode(flowExec *FlowExecution, nodeID string, outputPort int, msg *Message) ([]*Message, error) {
	var outputs []*Message

	// Find all connections from this node's output port
	for _, conn := range flowExec.Definition.Connections {
		if conn.Source == nodeID && conn.SourcePort == outputPort {
			// Check if target is a subflow output port
			if isOutputPort(conn.Target) {
				// Extract output port index
				outputPortIndex := extractPortIndex(conn.Target)
				flowExec.mu.Lock()
				flowExec.PendingOutputs[outputPortIndex] = append(flowExec.PendingOutputs[outputPortIndex], msg)
				flowExec.mu.Unlock()

				outputs = append(outputs, msg)
				continue
			}

			// Route to another internal node
			targetNode := flowExec.Definition.GetNode(conn.Target)
			if targetNode == nil {
				continue
			}

			nodeOutputs, err := e.executeNode(flowExec, targetNode, msg)
			if err != nil {
				return nil, err
			}

			// Recursively route outputs
			for i, nodeOutput := range nodeOutputs {
				if nodeOutput != nil {
					routedOutputs, err := e.routeFromNode(flowExec, targetNode.ID, i, nodeOutput)
					if err != nil {
						return nil, err
					}
					outputs = append(outputs, routedOutputs...)
				}
			}
		}
	}

	return outputs, nil
}

// isOutputPort checks if a target ID is an output port
func isOutputPort(target string) bool {
	return len(target) > 12 && target[:12] == "port-output-"
}

// extractPortIndex extracts the port index from a port ID
func extractPortIndex(portID string) int {
	var index int
	fmt.Sscanf(portID, "port-output-%d", &index)
	return index
}

// StopInstance stops a running subflow instance
func (e *Executor) StopInstance(instanceID string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	flowExec, ok := e.activeFlows[instanceID]
	if !ok {
		return fmt.Errorf("instance not running: %s", instanceID)
	}

	flowExec.Cancel()
	delete(e.activeFlows, instanceID)

	// Close message queue if exists
	if queue, ok := e.messageQueues[instanceID]; ok {
		close(queue)
		delete(e.messageQueues, instanceID)
	}

	return nil
}

// GetInstanceState retrieves the current state of a running instance
func (e *Executor) GetInstanceState(instanceID string) (map[string]any, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	flowExec, ok := e.activeFlows[instanceID]
	if !ok {
		return nil, fmt.Errorf("instance not running: %s", instanceID)
	}

	flowExec.mu.Lock()
	defer flowExec.mu.Unlock()

	state := map[string]any{
		"instanceId":  flowExec.InstanceID,
		"subflowId":   flowExec.SubflowID,
		"nodeStates":  flowExec.NodeStates,
		"nodeCount":   len(flowExec.Definition.Nodes),
		"activeNodes": len(flowExec.NodeStates),
	}

	return state, nil
}

// CreateMessage creates a new message with a unique ID
func CreateMessage(payload any, topic string) *Message {
	return &Message{
		Payload: payload,
		Topic:   topic,
		Metadata: map[string]any{
			"_msgid": uuid.New().String(),
		},
		Context: MessageContext{
			Variables: make(map[string]any),
		},
	}
}

// CloneMessage creates a deep copy of a message
func CloneMessage(msg *Message) *Message {
	clone := &Message{
		Payload:  msg.Payload,
		Topic:    msg.Topic,
		Metadata: make(map[string]any),
		Context:  msg.Context,
	}

	for k, v := range msg.Metadata {
		clone.Metadata[k] = v
	}

	clone.Metadata["_msgid"] = uuid.New().String()

	if msg.Context.Variables != nil {
		clone.Context.Variables = make(map[string]any)
		for k, v := range msg.Context.Variables {
			clone.Context.Variables[k] = v
		}
	}

	return clone
}
