package api

import (
	"time"

	"github.com/edgeflow/edgeflow/internal/engine"
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
func storageFlowToEngine(f *storage.Flow) *engine.Flow {
	if f == nil {
		return nil
	}

	flow := engine.NewFlow(f.Name, f.Description)
	// Preserve the ID from storage
	flow.ID = f.ID
	// Preserve the status
	flow.Status = engine.FlowStatus(f.Status)

	// TODO: Reconstruct nodes and connections from storage format
	// For now, we keep empty nodes/connections but preserve metadata

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
