package core

import (
	"context"

	"github.com/edgeflow/edgeflow/internal/node"
)

// CompleteNode triggers when a node or flow completes processing
type CompleteNode struct {
	scope   string   // "all", "flow", "nodes"
	nodeIds []string // Target node IDs when scope is "nodes"
}

// NewCompleteNode creates a new complete node
func NewCompleteNode() *CompleteNode {
	return &CompleteNode{
		scope:   "flow",
		nodeIds: []string{},
	}
}

// Init initializes the complete node with configuration
func (n *CompleteNode) Init(config map[string]interface{}) error {
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

// Execute processes completion messages
// In a full implementation, this would be triggered by the flow engine
// when monitored nodes complete their execution
func (n *CompleteNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Complete nodes receive completion notifications from the flow engine
	// The message payload should contain:
	// - _msgid: original message ID that completed
	// - source.id: node ID that completed
	// - source.type: node type
	// - source.name: node name

	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}

	// Check if this completion should be processed based on scope
	if !n.shouldProcess(msg) {
		return msg, nil
	}

	return msg, nil
}

// shouldProcess determines if a completion event should be processed
func (n *CompleteNode) shouldProcess(msg node.Message) bool {
	switch n.scope {
	case "all":
		return true
	case "flow":
		// In flow scope, only process completions from same flow
		return true
	case "nodes":
		// Only process completions from specified nodes
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
func (n *CompleteNode) getSourceID(msg node.Message) string {
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
func (n *CompleteNode) Cleanup() error {
	return nil
}
