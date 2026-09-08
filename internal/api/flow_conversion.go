package api

import (
	"strconv"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/engine"
	"github.com/EdgxCloud/EdgeFlow/internal/logger"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	"github.com/EdgxCloud/EdgeFlow/internal/storage"
	"go.uber.org/zap"
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
			// Preserved so a save/load round-trip keeps switch/if branches wired
			// to the correct output port.
			"sourceOutput": conn.SourcePort,
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

	flowLog := logger.WithFlow(f.ID, f.Name)
	flowLog.Debug("Converting storage flow to engine",
		zap.Int("nodes", len(f.Nodes)), zap.Int("connections", len(f.Connections)))

	registry := node.GetGlobalRegistry()

	flow := engine.NewFlow(f.Name, f.Description)
	// Preserve the ID from storage
	flow.ID = f.ID
	// Always start as idle — actual status is determined by runtime, not storage
	// Storage status is only used for display purposes (corrected in handlers)
	flow.Status = engine.FlowStatusIdle

	// Reconstruct nodes from storage
	for _, nodeData := range f.Nodes {
		nodeID, _ := nodeData["id"].(string)
		nodeType, _ := nodeData["type"].(string)
		nodeName, _ := nodeData["name"].(string)
		if nodeID == "" || nodeType == "" {
			flowLog.Debug("Skipping node with empty id or type", zap.String("id", nodeID), zap.String("type", nodeType))
			continue
		}
		if nodeName == "" {
			nodeName = nodeType
		}

		flowLog.Debug("Creating node", zap.String("node_id", nodeID), zap.String("type", nodeType), zap.String("name", nodeName))

		// Create node from registry (gets the correct executor)
		n, err := registry.CreateNode(nodeType, nodeName)
		if err != nil {
			flowLog.Error("Failed to create node", zap.String("node_id", nodeID), zap.String("type", nodeType), zap.Error(err))
			continue
		}

		// Override auto-generated ID with the stored ID
		n.ID = nodeID

		// Apply config if present
		if config, ok := nodeData["config"].(map[string]interface{}); ok {
			flowLog.Debug("Applying config", zap.String("node_id", nodeID), zap.Any("config", config))
			n.UpdateConfig(config)
		} else {
			flowLog.Debug("No config found for node", zap.String("node_id", nodeID))
		}

		// Add to flow
		if err := flow.AddNode(n); err != nil {
			flowLog.Error("Failed to add node to flow", zap.String("node_id", nodeID), zap.Error(err))
		}
	}

	// Reconstruct connections
	for _, connData := range f.Connections {
		sourceID, _ := connData["source"].(string)
		targetID, _ := connData["target"].(string)
		flowLog.Debug("Connecting nodes", zap.String("source", sourceID), zap.String("target", targetID))
		if sourceID == "" || targetID == "" {
			flowLog.Debug("Skipping connection with empty source/target")
			continue
		}
		// The editor records which output port an edge leaves from; without it
		// every branch of a switch/if node would receive every message.
		sourcePort := connSourcePort(connData)
		if err := flow.ConnectPort(sourceID, targetID, sourcePort); err != nil {
			flowLog.Error("Failed to connect nodes", zap.String("source", sourceID), zap.String("target", targetID), zap.Error(err))
		}
	}

	flowLog.Debug("Flow conversion complete", zap.Int("nodes", len(flow.Nodes)), zap.Int("connections", len(flow.Connections)))
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

// connSourcePort extracts the source output port from a stored connection.
// The editor writes it as "sourceOutput"; "sourcePort" and the React Flow
// "sourceHandle" string are accepted as well. Anything unrecognised is port 0,
// which keeps single-output nodes working.
func connSourcePort(connData map[string]interface{}) int {
	for _, key := range []string{"sourceOutput", "sourcePort"} {
		if v, ok := connData[key]; ok {
			if p, ok := toPortInt(v); ok {
				return p
			}
		}
	}
	if v, ok := connData["sourceHandle"]; ok {
		if p, ok := toPortInt(v); ok {
			return p
		}
	}
	return 0
}

func toPortInt(v interface{}) (int, bool) {
	switch t := v.(type) {
	case float64: // JSON numbers decode as float64
		return int(t), true
	case int:
		return t, true
	case string:
		if t == "" {
			return 0, false
		}
		p, err := strconv.Atoi(t)
		if err != nil {
			return 0, false
		}
		return p, true
	default:
		return 0, false
	}
}
