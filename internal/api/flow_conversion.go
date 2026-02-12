package api

import (
	"log"
	"time"

	"github.com/edgeflow/edgeflow/internal/engine"
	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/edgeflow/edgeflow/internal/storage"
)

// engineFlowToStorage converts engine.Flow to storage.Flow
func engineFlowToStorage(f *engine.Flow) *storage.Flow {
	if f == nil {
		return nil
	}

	// Convert nodes map to slice of maps - preserve all node data including config and position
	nodes := make([]map[string]interface{}, 0)
	for id, node := range f.Nodes {
		nodeMap := map[string]interface{}{
			"id":   id,
			"type": node.Type,
			"name": node.Name,
		}
		// Preserve config if it exists
		if node.Config != nil {
			nodeMap["config"] = node.Config
		}
		// Note: position is stored in Config map by frontend as "position" key
		nodes = append(nodes, nodeMap)
	}

	// Convert connections
	edges := make([]map[string]interface{}, 0)
	for _, conn := range f.Connections {
		edges = append(edges, map[string]interface{}{
			"id":     conn.ID,
			"source": conn.SourceID,
			"target": conn.TargetID,
		})
	}

	return &storage.Flow{
		ID:          f.ID,
		Name:        f.Name,
		Description: f.Description,
		Status:      string(f.Status),
		Nodes:       nodes,
		Connections: edges,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// storageFlowToEngine converts storage.Flow to engine.Flow
// Reconstructs nodes from the registry and wires up connections
func storageFlowToEngine(f *storage.Flow) *engine.Flow {
	if f == nil {
		return nil
	}

	registry := node.GetGlobalRegistry()

	flow := engine.NewFlow(f.Name, f.Description)
	// Preserve the ID from storage
	flow.ID = f.ID
	// Preserve the status
	flow.Status = engine.FlowStatus(f.Status)

	// Reconstruct nodes from storage
	for _, nodeData := range f.Nodes {
		nodeID, _ := nodeData["id"].(string)
		nodeType, _ := nodeData["type"].(string)
		nodeName, _ := nodeData["name"].(string)
		if nodeID == "" || nodeType == "" {
			continue
		}
		if nodeName == "" {
			nodeName = nodeType
		}

		// Create node from registry (gets the correct executor)
		n, err := registry.CreateNode(nodeType, nodeName)
		if err != nil {
			log.Printf("Warning: failed to create node %s (%s): %v", nodeID, nodeType, err)
			continue
		}

		// Override auto-generated ID with the stored ID
		n.ID = nodeID

		// Apply config if present
		if config, ok := nodeData["config"].(map[string]interface{}); ok {
			n.UpdateConfig(config)
		}

		// Add to flow
		if err := flow.AddNode(n); err != nil {
			log.Printf("Warning: failed to add node %s to flow: %v", nodeID, err)
		}
	}

	// Reconstruct connections
	for _, connData := range f.Connections {
		sourceID, _ := connData["source"].(string)
		targetID, _ := connData["target"].(string)
		if sourceID == "" || targetID == "" {
			continue
		}
		if err := flow.Connect(sourceID, targetID); err != nil {
			log.Printf("Warning: failed to connect %s -> %s: %v", sourceID, targetID, err)
		}
	}

	return flow
}

// storageFlowsToEngine converts a slice of storage.Flow to engine.Flow
func storageFlowsToEngine(flows []*storage.Flow) []*engine.Flow {
	result := make([]*engine.Flow, len(flows))
	for i, f := range flows {
		result[i] = storageFlowToEngine(f)
	}
	return result
}
