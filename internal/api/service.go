package api

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/engine"
	"github.com/edgeflow/edgeflow/internal/hal"
	"github.com/edgeflow/edgeflow/internal/logger"
	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/edgeflow/edgeflow/internal/resources"
	"github.com/edgeflow/edgeflow/internal/storage"
	"github.com/edgeflow/edgeflow/internal/websocket"
	"go.uber.org/zap"
)

// ExecutionRecord represents a tracked flow execution
type ExecutionRecord struct {
	ID             string                `json:"id"`
	FlowID         string                `json:"flow_id"`
	FlowName       string                `json:"flow_name"`
	Status         string                `json:"status"` // running, completed, failed
	StartTime      time.Time             `json:"start_time"`
	EndTime        *time.Time            `json:"end_time,omitempty"`
	Duration       *int64                `json:"duration,omitempty"` // milliseconds
	NodeCount      int                   `json:"node_count"`
	CompletedNodes int                   `json:"completed_nodes"`
	ErrorNodes     int                   `json:"error_nodes"`
	Error          string                `json:"error,omitempty"`
	NodeEvents     []NodeExecutionEvent  `json:"node_events,omitempty"`
	mu             sync.Mutex
}

// NodeExecutionEvent is a single node execution within a flow run
type NodeExecutionEvent struct {
	NodeID        string                 `json:"node_id"`
	NodeName      string                 `json:"node_name"`
	NodeType      string                 `json:"node_type"`
	Status        string                 `json:"status"`
	ExecutionTime int64                  `json:"execution_time"`
	Timestamp     int64                  `json:"timestamp"`
	Error         string                 `json:"error,omitempty"`
}

// Service handles business logic for the API
type Service struct {
	storage         storage.Storage
	registry        *node.Registry
	// pluginManager   *plugin.Manager
	resourceMonitor *resources.Monitor
	gpioMonitor     *hal.GPIOMonitor
	flows           map[string]*engine.Flow // Active flows in memory
	wsHub           *websocket.Hub
	executions      []*ExecutionRecord // In-memory execution history
	execMu          sync.RWMutex
}

// NewService creates a new API service
func NewService(storage storage.Storage, registry *node.Registry, wsHub *websocket.Hub) *Service {
	// Initialize resource monitor with default limits
	limits := resources.ResourceLimits{
		MemoryLimit:           1024 * 1024 * 1024 * 4, // 4GB
		MemoryHardLimit:       1024 * 1024 * 1024 * 8, // 8GB
		DiskLimit:             1024 * 1024 * 1024 * 50, // 50GB
		LowMemoryThreshold:    100 * 1024 * 1024, // 100MB
		AutoDisableOnLowMemory: false,
	}
	resourceMonitor := resources.NewMonitor(limits)

	// Start resource monitoring in background
	go resourceMonitor.Start(context.Background(), 5 * 1000000000) // 5 seconds

	// Initialize plugin manager
	// pluginManager := plugin.NewManager(registry, resourceMonitor)

	// Initialize GPIO monitor for real-time pin state broadcasting
	gpioMonitor := hal.NewGPIOMonitor(200, func(state hal.GPIOMonitorState) {
		wsHub.Broadcast(websocket.MessageTypeGPIOState, map[string]interface{}{
			"pins":       state.Pins,
			"board_name": state.BoardName,
			"gpio_chip":  state.GPIOChip,
			"available":  state.Available,
			"timestamp":  state.Timestamp,
		})
	})
	go gpioMonitor.Start()
	hal.SetGlobalGPIOMonitor(gpioMonitor)

	return &Service{
		storage:         storage,
		registry:        registry,
		// pluginManager:   pluginManager,
		resourceMonitor: resourceMonitor,
		gpioMonitor:     gpioMonitor,
		flows:           make(map[string]*engine.Flow),
		wsHub:           wsHub,
		executions:      make([]*ExecutionRecord, 0),
	}
}

// logActivity logs an activity using the structured logger (which also broadcasts to WebSocket)
func (s *Service) logActivity(level, message, source string) {
	l := logger.Get().With(zap.String("source", source))
	switch level {
	case "error":
		l.Error(message)
	case "warn":
		l.Warn(message)
	case "debug":
		l.Debug(message)
	default:
		l.Info(message)
	}
}

// GetGPIOState returns the current GPIO pin states
func (s *Service) GetGPIOState() hal.GPIOMonitorState {
	if s.gpioMonitor != nil {
		return s.gpioMonitor.GetState()
	}
	return hal.GPIOMonitorState{
		Pins:      make(map[int]*hal.PinState),
		Available: false,
		Timestamp: time.Now(),
	}
}

// CreateFlow creates a new flow
func (s *Service) CreateFlow(name, description string) (*engine.Flow, error) {
	flow := engine.NewFlow(name, description)

	// Save to storage (convert to storage type)
	storageFlow := engineFlowToStorage(flow)
	if err := s.storage.SaveFlow(storageFlow); err != nil {
		return nil, fmt.Errorf("failed to save flow: %w", err)
	}

	// Add to active flows
	s.flows[flow.ID] = flow

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeFlowStatus, map[string]interface{}{
		"flow_id": flow.ID,
		"action":  "created",
		"name":    flow.Name,
	})
	s.logActivity("info", fmt.Sprintf("Flow created: %s", name), "flow")

	return flow, nil
}

// GetFlow retrieves a flow by ID
func (s *Service) GetFlow(id string) (*engine.Flow, error) {
	// Check if flow is in memory
	if flow, ok := s.flows[id]; ok {
		return flow, nil
	}

	// Load from storage
	storageFlow, err := s.storage.GetFlow(id)
	if err != nil {
		return nil, err
	}

	// Convert to engine flow
	flow := storageFlowToEngine(storageFlow)
	return flow, nil
}

// ListFlows returns all flows
func (s *Service) ListFlows() ([]*engine.Flow, error) {
	storageFlows, err := s.storage.ListFlows()
	if err != nil {
		return nil, err
	}
	return storageFlowsToEngine(storageFlows), nil
}

// UpdateFlow updates an existing flow
func (s *Service) UpdateFlow(flow *engine.Flow) error {
	storageFlow := engineFlowToStorage(flow)
	if err := s.storage.UpdateFlow(storageFlow); err != nil {
		return fmt.Errorf("failed to update flow: %w", err)
	}

	// Update in memory
	s.flows[flow.ID] = flow

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeFlowStatus, map[string]interface{}{
		"flow_id": flow.ID,
		"action":  "updated",
		"name":    flow.Name,
	})
	s.logActivity("info", fmt.Sprintf("Flow updated: %s", flow.Name), "flow")

	return nil
}

// DeleteFlow deletes a flow
func (s *Service) DeleteFlow(id string) error {
	// Stop flow if it's in memory (best effort - don't fail if this errors)
	if flow, ok := s.flows[id]; ok {
		// Try to stop, but don't return error if it fails
		_ = flow.Stop()
		delete(s.flows, id)
	}

	// Delete from storage - best effort, don't fail if flow doesn't exist in storage
	if err := s.storage.DeleteFlow(id); err != nil {
		// Only return error if it's not a "not found" error
		if err.Error() != fmt.Sprintf("flow not found: %s", id) {
			return fmt.Errorf("failed to delete flow: %w", err)
		}
		// Flow wasn't in storage, that's OK - it might have only been in memory
	}

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeFlowStatus, map[string]interface{}{
		"flow_id": id,
		"action":  "deleted",
	})
	s.logActivity("warn", fmt.Sprintf("Flow deleted: %s", id), "flow")

	return nil
}

// StartFlow starts a flow execution
func (s *Service) StartFlow(id string) error {
	flowLogger := logger.WithFlow(id, "")
	flowLogger.Info("Starting flow")
	flow, err := s.GetFlow(id)
	if err != nil {
		flowLogger.Error("Failed to get flow", zap.Error(err))
		return err
	}

	flowLogger = logger.WithFlow(id, flow.Name)
	flowLogger.Info("Flow loaded", zap.Int("nodes", len(flow.Nodes)), zap.Int("connections", len(flow.Connections)))
	for nid, n := range flow.Nodes {
		flowLogger.Debug("Node", zap.String("node_id", nid), zap.String("type", n.Type), zap.String("name", n.Name))
	}
	for _, conn := range flow.Connections {
		flowLogger.Debug("Connection", zap.String("source", conn.SourceID), zap.String("target", conn.TargetID))
	}

	// Create execution record
	execID := fmt.Sprintf("exec-%d", time.Now().UnixNano())
	record := &ExecutionRecord{
		ID:        execID,
		FlowID:    flow.ID,
		FlowName:  flow.Name,
		Status:    "running",
		StartTime: time.Now(),
		NodeCount: len(flow.Nodes),
	}
	s.execMu.Lock()
	// Keep max 100 execution records
	if len(s.executions) >= 100 {
		s.executions = s.executions[len(s.executions)-99:]
	}
	s.executions = append(s.executions, record)
	s.execMu.Unlock()

	// Set execution callback to broadcast node execution data via WebSocket
	flowID := flow.ID
	flow.SetExecutionCallback(func(event node.ExecutionEvent) {
		// Broadcast via WebSocket
		s.wsHub.Broadcast(websocket.MessageTypeExecution, map[string]interface{}{
			"flow_id":        flowID,
			"node_id":        event.NodeID,
			"node_name":      event.NodeName,
			"node_type":      event.NodeType,
			"input":          event.Input,
			"output":         event.Output,
			"status":         event.Status,
			"error":          event.Error,
			"execution_time": event.ExecutionTime,
			"timestamp":      event.Timestamp,
		})

		// Track in execution record
		record.mu.Lock()
		record.NodeEvents = append(record.NodeEvents, NodeExecutionEvent{
			NodeID:        event.NodeID,
			NodeName:      event.NodeName,
			NodeType:      event.NodeType,
			Status:        event.Status,
			ExecutionTime: event.ExecutionTime,
			Timestamp:     event.Timestamp,
			Error:         event.Error,
		})
		if event.Status == "success" {
			record.CompletedNodes++
		} else if event.Status == "error" {
			record.ErrorNodes++
		}
		record.mu.Unlock()
	})

	// Start the flow
	ctx := context.Background()
	if err := flow.Start(ctx); err != nil {
		record.mu.Lock()
		record.Status = "failed"
		now := time.Now()
		record.EndTime = &now
		dur := now.Sub(record.StartTime).Milliseconds()
		record.Duration = &dur
		record.Error = err.Error()
		record.mu.Unlock()
		return fmt.Errorf("failed to start flow: %w", err)
	}

	// Store in active flows
	s.flows[id] = flow

	// Persist "running" status to storage
	s.persistFlowStatus(id, "running")

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeFlowStatus, map[string]interface{}{
		"flow_id": flow.ID,
		"action":  "started",
		"status":  flow.Status,
	})
	s.logActivity("success", fmt.Sprintf("Flow started: %s", flow.Name), "runtime")

	return nil
}

// StopFlow stops a flow execution
func (s *Service) StopFlow(id string) error {
	flow, ok := s.flows[id]
	if !ok {
		// Flow not in memory — just update storage status
		s.persistFlowStatus(id, "stopped")
		return nil
	}

	if err := flow.Stop(); err != nil {
		return fmt.Errorf("failed to stop flow: %w", err)
	}

	// Remove from active flows
	delete(s.flows, id)

	// Finalize execution record
	s.finalizeExecution(id, "completed", "")

	// Persist "stopped" status to storage
	s.persistFlowStatus(id, "stopped")

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeFlowStatus, map[string]interface{}{
		"flow_id": flow.ID,
		"action":  "stopped",
		"status":  flow.Status,
	})
	s.logActivity("info", fmt.Sprintf("Flow stopped: %s", flow.Name), "runtime")

	return nil
}

// persistFlowStatus updates the flow's status in storage
func (s *Service) persistFlowStatus(id string, status string) {
	storageFlow, err := s.storage.GetFlow(id)
	if err != nil {
		logger.Warn("Failed to get flow from storage for status update", zap.String("flow_id", id), zap.Error(err))
		return
	}
	storageFlow.Status = status
	storageFlow.UpdatedAt = time.Now()
	if err := s.storage.SaveFlow(storageFlow); err != nil {
		logger.Warn("Failed to persist flow status", zap.String("flow_id", id), zap.String("status", status), zap.Error(err))
	}
}

// AddNodeToFlow adds a node to a flow
func (s *Service) AddNodeToFlow(flowID, nodeType, nodeName string, config map[string]interface{}) (*node.Node, error) {
	flow, err := s.GetFlow(flowID)
	if err != nil {
		return nil, err
	}

	// Create node from registry
	newNode, err := s.registry.CreateNode(nodeType, nodeName)
	if err != nil {
		return nil, fmt.Errorf("failed to create node: %w", err)
	}

	// Configure node
	if config != nil {
		if err := newNode.UpdateConfig(config); err != nil {
			return nil, fmt.Errorf("failed to configure node: %w", err)
		}
	}

	// Add to flow
	if err := flow.AddNode(newNode); err != nil {
		return nil, fmt.Errorf("failed to add node to flow: %w", err)
	}

	// Save flow
	if err := s.storage.UpdateFlow(engineFlowToStorage(flow)); err != nil {
		return nil, fmt.Errorf("failed to save flow: %w", err)
	}

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeNodeStatus, map[string]interface{}{
		"flow_id": flowID,
		"node_id": newNode.ID,
		"action":  "added",
		"type":    nodeType,
	})
	s.logActivity("info", fmt.Sprintf("Node added: %s (%s)", nodeName, nodeType), "node")

	return newNode, nil
}

// RemoveNodeFromFlow removes a node from a flow
func (s *Service) RemoveNodeFromFlow(flowID, nodeID string) error {
	flow, err := s.GetFlow(flowID)
	if err != nil {
		return err
	}

	if err := flow.RemoveNode(nodeID); err != nil {
		return fmt.Errorf("failed to remove node: %w", err)
	}

	// Save flow
	if err := s.storage.UpdateFlow(engineFlowToStorage(flow)); err != nil {
		return fmt.Errorf("failed to save flow: %w", err)
	}

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeNodeStatus, map[string]interface{}{
		"flow_id": flowID,
		"node_id": nodeID,
		"action":  "removed",
	})
	s.logActivity("warn", fmt.Sprintf("Node removed: %s from flow %s", nodeID, flowID), "node")

	return nil
}

// ConnectNodes creates a connection between two nodes in a flow
func (s *Service) ConnectNodes(flowID, sourceID, targetID string) error {
	flow, err := s.GetFlow(flowID)
	if err != nil {
		return err
	}

	if err := flow.Connect(sourceID, targetID); err != nil {
		return fmt.Errorf("failed to connect nodes: %w", err)
	}

	// Save flow
	if err := s.storage.UpdateFlow(engineFlowToStorage(flow)); err != nil {
		return fmt.Errorf("failed to save flow: %w", err)
	}

	return nil
}

// GetNodeTypes returns all available node types
func (s *Service) GetNodeTypes() []*node.NodeInfo {
	return s.registry.List()
}

// GetNodeTypeInfo returns info for a specific node type
func (s *Service) GetNodeTypeInfo(nodeType string) (*node.NodeInfo, error) {
	return s.registry.Get(nodeType)
}

// Close closes the service and cleans up resources
func (s *Service) Close() error {
	// Stop all active flows
	for _, flow := range s.flows {
		flow.Stop()
	}

	// Shutdown plugin manager
	// if s.pluginManager != nil {
	// 	if err := s.pluginManager.Shutdown(); err != nil {
	// 		return fmt.Errorf("failed to shutdown plugin manager: %w", err)
	// 	}
	// }

	// Close storage
	if err := s.storage.Close(); err != nil {
		return err
	}

	return nil
}

// GetResourceStats returns resource statistics
func (s *Service) GetResourceStats() resources.ResourceStats {
	return s.resourceMonitor.GetStats()
}

// GetStorageFlow retrieves a flow from storage in raw format
func (s *Service) GetStorageFlow(id string) (*storage.Flow, error) {
	return s.storage.GetFlow(id)
}

// UpdateStorageFlow updates a flow directly in storage
func (s *Service) UpdateStorageFlow(flow *storage.Flow) error {
	return s.storage.UpdateFlow(flow)
}

// ListStorageFlows retrieves all flows from storage in raw format
func (s *Service) ListStorageFlows() ([]*storage.Flow, error) {
	return s.storage.ListFlows()
}

// IsFlowRunning checks if a flow is actively running in memory
func (s *Service) IsFlowRunning(id string) bool {
	flow, ok := s.flows[id]
	if !ok {
		return false
	}
	return flow.GetStatus() == engine.FlowStatusRunning
}

// finalizeExecution marks an execution record as completed/failed
func (s *Service) finalizeExecution(flowID, status, errMsg string) {
	s.execMu.RLock()
	defer s.execMu.RUnlock()

	// Find the latest running execution for this flow
	for i := len(s.executions) - 1; i >= 0; i-- {
		rec := s.executions[i]
		if rec.FlowID == flowID && rec.Status == "running" {
			rec.mu.Lock()
			rec.Status = status
			now := time.Now()
			rec.EndTime = &now
			dur := now.Sub(rec.StartTime).Milliseconds()
			rec.Duration = &dur
			if errMsg != "" {
				rec.Error = errMsg
			}
			if rec.ErrorNodes > 0 && status != "failed" {
				rec.Status = "completed"
			}
			rec.mu.Unlock()
			return
		}
	}
}

// ListExecutions returns all execution records
func (s *Service) ListExecutions() []*ExecutionRecord {
	s.execMu.RLock()
	defer s.execMu.RUnlock()
	// Return in reverse order (newest first)
	result := make([]*ExecutionRecord, len(s.executions))
	for i, rec := range s.executions {
		result[len(s.executions)-1-i] = rec
	}
	return result
}

// GetExecution returns a single execution record by ID
func (s *Service) GetExecution(id string) (*ExecutionRecord, error) {
	s.execMu.RLock()
	defer s.execMu.RUnlock()
	for _, rec := range s.executions {
		if rec.ID == id {
			return rec, nil
		}
	}
	return nil, fmt.Errorf("execution not found: %s", id)
}
