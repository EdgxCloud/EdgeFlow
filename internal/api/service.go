package api

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/engine"
	"github.com/edgeflow/edgeflow/internal/node"
	// "github.com/edgeflow/edgeflow/internal/plugin"
	"github.com/edgeflow/edgeflow/internal/resources"
	"github.com/edgeflow/edgeflow/internal/storage"
	"github.com/edgeflow/edgeflow/internal/websocket"
	// "github.com/google/uuid"
)

// Service handles business logic for the API
type Service struct {
	storage         storage.Storage
	registry        *node.Registry
	// pluginManager   *plugin.Manager
	resourceMonitor *resources.Monitor
	flows           map[string]*engine.Flow // Active flows in memory
	wsHub           *websocket.Hub
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

	return &Service{
		storage:         storage,
		registry:        registry,
		// pluginManager:   pluginManager,
		resourceMonitor: resourceMonitor,
		flows:           make(map[string]*engine.Flow),
		wsHub:           wsHub,
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

	return nil
}

// StartFlow starts a flow execution
func (s *Service) StartFlow(id string) error {
	flow, err := s.GetFlow(id)
	if err != nil {
		return err
	}

	// Start the flow
	ctx := context.Background()
	if err := flow.Start(ctx); err != nil {
		return fmt.Errorf("failed to start flow: %w", err)
	}

	// Store in active flows
	s.flows[id] = flow

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeFlowStatus, map[string]interface{}{
		"flow_id": flow.ID,
		"action":  "started",
		"status":  flow.Status,
	})

	return nil
}

// StopFlow stops a flow execution
func (s *Service) StopFlow(id string) error {
	flow, ok := s.flows[id]
	if !ok {
		return fmt.Errorf("flow not found or not running")
	}

	if err := flow.Stop(); err != nil {
		return fmt.Errorf("failed to stop flow: %w", err)
	}

	// Notify via WebSocket
	s.wsHub.Broadcast(websocket.MessageTypeFlowStatus, map[string]interface{}{
		"flow_id": flow.ID,
		"action":  "stopped",
		"status":  flow.Status,
	})

	return nil
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

// Module Management Methods (temporarily disabled)

// ListModules returns all available modules
func (s *Service) ListModules() []map[string]interface{} {
	// return // s.pluginManager.ListPlugins()
	return []map[string]interface{}{}
}

// GetModuleInfo returns info for a specific module
func (s *Service) GetModuleInfo(name string) (map[string]interface{}, error) {
	// return // s.pluginManager.GetPluginInfo(name)
	return map[string]interface{}{}, fmt.Errorf("plugin manager disabled")
}

// LoadModule loads a module
func (s *Service) LoadModule(name string) error {
	// Temporarily disabled
	return fmt.Errorf("plugin manager disabled")
}

// UnloadModule unloads a module
func (s *Service) UnloadModule(name string) error {
	// Temporarily disabled
	return fmt.Errorf("plugin manager disabled")
}

// EnableModule enables a module
func (s *Service) EnableModule(name string) error {
	// Temporarily disabled
	return fmt.Errorf("plugin manager disabled")
}

// DisableModule disables a module
func (s *Service) DisableModule(name string) error {
	// Temporarily disabled
	return fmt.Errorf("plugin manager disabled")
}

// ReloadModule reloads a module
func (s *Service) ReloadModule(name string) error {
	// Temporarily disabled
	return fmt.Errorf("plugin manager disabled")

	// if err := s.pluginManager.ReloadPlugin(name); err != nil {
	// 	return err
	// }

	// // Notify via WebSocket
	// s.wsHub.Broadcast(websocket.MessageTypeModuleStatus, map[string]interface{}{
	// 	"module": name,
	// 	"action": "reloaded",
	// })

	// return nil
}

// GetModuleStats returns module statistics
func (s *Service) GetModuleStats() map[string]interface{} {
	// return s.pluginManager.GetStats()
	return map[string]interface{}{"error": "plugin manager disabled"}
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
