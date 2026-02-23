package core

import (
	"context"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// CatchNode captures errors from other nodes
// Supports scope filtering: all flows, current flow, or specific nodes
type CatchNode struct {
	scope      string   // "all", "flow", "nodes"
	uncaught   bool     // Catch only uncaught errors
	nodeIDs    []string // Specific node IDs to catch errors from
	flowID     string   // Current flow ID (set at runtime)
}

// NewCatchNode creates a new catch node
func NewCatchNode() *CatchNode {
	return &CatchNode{
		scope:    "flow",
		uncaught: false,
		nodeIDs:  make([]string, 0),
	}
}

// Init initializes the catch node with configuration
func (n *CatchNode) Init(config map[string]interface{}) error {
	if scope, ok := config["scope"].(string); ok {
		n.scope = scope
	}

	if uncaught, ok := config["uncaught"].(bool); ok {
		n.uncaught = uncaught
	}

	// Parse node IDs for "nodes" scope
	if nodeIDs, ok := config["nodeIds"].([]interface{}); ok {
		for _, id := range nodeIDs {
			if idStr, ok := id.(string); ok {
				n.nodeIDs = append(n.nodeIDs, idStr)
			}
		}
	}

	return nil
}

// Execute forwards error messages if they match the scope
func (n *CatchNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// This node doesn't process normal messages
	// It only receives messages that were routed here by the error handling system

	// Check if this is an error message
	em, ok := interface{}(msg).(*node.EnhancedMessage)
	if !ok || em.Error == nil {
		// Not an error message, don't forward
		return node.Message{}, nil
	}

	// Check if we should catch this error based on scope
	if !n.shouldCatchError(em) {
		// Don't catch this error
		return node.Message{}, nil
	}

	// Forward the error message
	return msg, nil
}

// Cleanup stops the catch node
func (n *CatchNode) Cleanup() error {
	return nil
}

// shouldCatchError determines if this error should be caught
func (n *CatchNode) shouldCatchError(msg *node.EnhancedMessage) bool {
	if msg.Error == nil {
		return false
	}

	// Check uncaught flag
	if n.uncaught {
		// Only catch if error hasn't been caught by another catch node
		if caught, ok := msg.Metadata["_errorCaught"].(bool); ok && caught {
			return false
		}
	}

	// Check scope
	switch n.scope {
	case "all":
		// Catch all errors
		return true

	case "flow":
		// Catch errors from current flow only
		if msg.Error.Source != nil {
			// TODO: Compare flow IDs when flow context is available
			// For now, catch all errors (same as "all" scope)
			return true
		}
		return true

	case "nodes":
		// Catch errors from specific nodes only
		if msg.Error.Source != nil {
			sourceID := msg.Error.Source.ID
			for _, nodeID := range n.nodeIDs {
				if sourceID == nodeID {
					return true
				}
			}
		}
		return false

	default:
		return false
	}
}

// SetFlowID sets the current flow ID (called by engine at runtime)
func (n *CatchNode) SetFlowID(flowID string) {
	n.flowID = flowID
}
