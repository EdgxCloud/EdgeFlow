package engine

import (
	"sync"
	"time"
)

// ExecutionContext holds the runtime context for a flow execution
type ExecutionContext struct {
	FlowID       string                 `json:"flow_id"`
	ExecutionID  string                 `json:"execution_id"`
	StartTime    time.Time              `json:"start_time"`
	Variables    map[string]interface{} `json:"variables"`
	GlobalVars   map[string]interface{} `json:"global_vars"`
	NodeStates   map[string]NodeState   `json:"node_states"`
	mu           sync.RWMutex
}

// NodeState tracks the execution state of a node
type NodeState struct {
	NodeID       string                 `json:"node_id"`
	Status       string                 `json:"status"` // pending, running, completed, error
	StartTime    time.Time              `json:"start_time"`
	EndTime      time.Time              `json:"end_time"`
	Duration     time.Duration          `json:"duration"`
	Error        string                 `json:"error,omitempty"`
	Input        map[string]interface{} `json:"input"`
	Output       map[string]interface{} `json:"output"`
	ExecutionCount int                  `json:"execution_count"`
}

// NewExecutionContext creates a new execution context
func NewExecutionContext(flowID, executionID string) *ExecutionContext {
	return &ExecutionContext{
		FlowID:      flowID,
		ExecutionID: executionID,
		StartTime:   time.Now(),
		Variables:   make(map[string]interface{}),
		GlobalVars:  make(map[string]interface{}),
		NodeStates:  make(map[string]NodeState),
	}
}

// SetVariable sets a variable in the execution context
func (ec *ExecutionContext) SetVariable(key string, value interface{}) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.Variables[key] = value
}

// GetVariable gets a variable from the execution context
func (ec *ExecutionContext) GetVariable(key string) (interface{}, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	val, exists := ec.Variables[key]
	return val, exists
}

// SetGlobalVariable sets a global variable
func (ec *ExecutionContext) SetGlobalVariable(key string, value interface{}) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.GlobalVars[key] = value
}

// GetGlobalVariable gets a global variable
func (ec *ExecutionContext) GetGlobalVariable(key string) (interface{}, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	val, exists := ec.GlobalVars[key]
	return val, exists
}

// SetNodeState sets the state of a node
func (ec *ExecutionContext) SetNodeState(nodeID string, state NodeState) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	ec.NodeStates[nodeID] = state
}

// GetNodeState gets the state of a node
func (ec *ExecutionContext) GetNodeState(nodeID string) (NodeState, bool) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	state, exists := ec.NodeStates[nodeID]
	return state, exists
}

// UpdateNodeStatus updates the status of a node
func (ec *ExecutionContext) UpdateNodeStatus(nodeID, status string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	state, exists := ec.NodeStates[nodeID]
	if !exists {
		state = NodeState{
			NodeID:    nodeID,
			StartTime: time.Now(),
		}
	}

	state.Status = status

	if status == "running" && state.StartTime.IsZero() {
		state.StartTime = time.Now()
	} else if status == "completed" || status == "error" {
		state.EndTime = time.Now()
		state.Duration = state.EndTime.Sub(state.StartTime)
	}

	ec.NodeStates[nodeID] = state
}

// IncrementNodeExecution increments the execution count for a node
func (ec *ExecutionContext) IncrementNodeExecution(nodeID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()

	state, exists := ec.NodeStates[nodeID]
	if !exists {
		state = NodeState{
			NodeID: nodeID,
		}
	}

	state.ExecutionCount++
	ec.NodeStates[nodeID] = state
}

// GetDuration returns the total execution duration
func (ec *ExecutionContext) GetDuration() time.Duration {
	return time.Since(ec.StartTime)
}

// Clone creates a copy of the execution context
func (ec *ExecutionContext) Clone() *ExecutionContext {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	clone := &ExecutionContext{
		FlowID:      ec.FlowID,
		ExecutionID: ec.ExecutionID,
		StartTime:   ec.StartTime,
		Variables:   make(map[string]interface{}),
		GlobalVars:  make(map[string]interface{}),
		NodeStates:  make(map[string]NodeState),
	}

	// Copy variables
	for k, v := range ec.Variables {
		clone.Variables[k] = v
	}

	// Copy global variables
	for k, v := range ec.GlobalVars {
		clone.GlobalVars[k] = v
	}

	// Copy node states
	for k, v := range ec.NodeStates {
		clone.NodeStates[k] = v
	}

	return clone
}

// GetStats returns execution statistics
func (ec *ExecutionContext) GetStats() map[string]interface{} {
	ec.mu.RLock()
	defer ec.mu.RUnlock()

	completedNodes := 0
	errorNodes := 0
	runningNodes := 0

	for _, state := range ec.NodeStates {
		switch state.Status {
		case "completed":
			completedNodes++
		case "error":
			errorNodes++
		case "running":
			runningNodes++
		}
	}

	return map[string]interface{}{
		"total_nodes":     len(ec.NodeStates),
		"completed_nodes": completedNodes,
		"error_nodes":     errorNodes,
		"running_nodes":   runningNodes,
		"duration":        ec.GetDuration().String(),
		"variables_count": len(ec.Variables),
	}
}
