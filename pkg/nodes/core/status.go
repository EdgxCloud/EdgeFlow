package core

import (
	"context"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// StatusNode monitors node status changes
type StatusNode struct {
	scope   string   // "all", "flow", "nodes"
	nodeIds []string // Target node IDs when scope is "nodes"
}

// NewStatusNode creates a new status node
func NewStatusNode() *StatusNode {
	return &StatusNode{
		scope:   "all",
		nodeIds: []string{},
	}
}

// Init initializes the status node with configuration
func (n *StatusNode) Init(config map[string]interface{}) error {
	if scope, ok := config["scope"].(string); ok {
		n.scope = scope
	}
	if nodeIds, ok := config["nodeIds"].([]interface{}); ok {
		for _, id := range nodeIds {
			if idStr, ok := id.(string); ok {
				n.nodeIds = append(n.nodeIds, idStr)
			}
		}
	}
	return nil
}

// Execute processes status messages
// In a full implementation, this would be triggered by the flow engine
// when monitored nodes change status
func (n *StatusNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Status nodes typically receive status updates from the flow engine
	// The message payload should contain status information:
	// - status.fill: color (red, green, yellow, blue, grey)
	// - status.shape: dot or ring
	// - status.text: status text
	// - source.id: node ID that generated status
	// - source.type: node type
	// - source.name: node name

	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}

	// Check if this status update should be processed based on scope
	if !n.shouldProcess(msg) {
		return msg, nil
	}

	return msg, nil
}

// shouldProcess determines if a status update should be processed
func (n *StatusNode) shouldProcess(msg node.Message) bool {
	switch n.scope {
	case "all":
		return true
	case "flow":
		// In flow scope, only process status from same flow
		// This would require flow context to be passed
		return true
	case "nodes":
		// Only process status from specified nodes
		if sourceID := n.getSourceID(msg); sourceID != "" {
			for _, id := range n.nodeIds {
				if id == sourceID {
					return true
				}
			}
			return false
		}
		return true
	default:
		return true
	}
}

// getSourceID extracts the source node ID from the message
func (n *StatusNode) getSourceID(msg node.Message) string {
	if msg.Payload == nil {
		return ""
	}
	if source, ok := msg.Payload["source"].(map[string]interface{}); ok {
		if id, ok := source["id"].(string); ok {
			return id
		}
	}
	return ""
}

// Cleanup cleans up resources
func (n *StatusNode) Cleanup() error {
	return nil
}
