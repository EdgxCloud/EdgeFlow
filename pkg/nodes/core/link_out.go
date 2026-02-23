package core

import (
	"context"
	"fmt"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// LinkOutNode sends messages to Link In nodes with matching link ID
// This provides virtual wiring to avoid visual clutter in complex flows
type LinkOutNode struct {
	linkIDs    []string // Can send to multiple link IDs
	scope      string   // "global" or "flow"
	flowID     string
	mode       string   // "return" or "continue"
}

// NewLinkOutNode creates a new Link Out node
func NewLinkOutNode() *LinkOutNode {
	return &LinkOutNode{
		linkIDs: []string{},
		scope:   "global",
		mode:    "continue",
	}
}

// Init initializes the Link Out node
func (n *LinkOutNode) Init(config map[string]interface{}) error {
	// Parse link IDs (required, can be array or single string)
	if linkIDsRaw, ok := config["linkIds"]; ok {
		switch v := linkIDsRaw.(type) {
		case []interface{}:
			for _, id := range v {
				if strID, ok := id.(string); ok {
					n.linkIDs = append(n.linkIDs, strID)
				}
			}
		case []string:
			n.linkIDs = v
		case string:
			n.linkIDs = []string{v}
		default:
			return fmt.Errorf("linkIds must be string or array of strings")
		}
	} else if linkID, ok := config["linkId"].(string); ok {
		// Support single linkId for backward compatibility
		n.linkIDs = []string{linkID}
	} else {
		return fmt.Errorf("linkIds or linkId is required for Link Out node")
	}

	if len(n.linkIDs) == 0 {
		return fmt.Errorf("at least one linkId is required")
	}

	// Parse scope (global or flow)
	if scope, ok := config["scope"].(string); ok {
		if scope != "global" && scope != "flow" {
			return fmt.Errorf("scope must be 'global' or 'flow'")
		}
		n.scope = scope
	}

	// Parse flow ID (for flow-scoped links)
	if flowID, ok := config["flowId"].(string); ok {
		n.flowID = flowID
	}

	// Parse mode (return or continue)
	if mode, ok := config["mode"].(string); ok {
		if mode != "return" && mode != "continue" {
			return fmt.Errorf("mode must be 'return' or 'continue'")
		}
		n.mode = mode
	}

	// Validate flow scope has flowID
	if n.scope == "flow" && n.flowID == "" {
		return fmt.Errorf("flowId is required for flow-scoped links")
	}

	return nil
}

// Execute sends the message to all matching Link In nodes
func (n *LinkOutNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get all matching Link In nodes from registry
	linkInNodes := GetLinkInNodes(n.linkIDs, n.scope, n.flowID)

	if len(linkInNodes) == 0 {
		// No matching Link In nodes found
		// This is not necessarily an error - the Link In might not be deployed yet
		// Log a warning but don't fail
	}

	// Send message to all matching Link In nodes
	var lastErr error
	for _, linkIn := range linkInNodes {
		if err := linkIn.ReceiveMessage(msg); err != nil {
			lastErr = err
			// Continue sending to other nodes even if one fails
		}
	}

	// Return behavior depends on mode
	if n.mode == "return" {
		// Return mode: don't pass message to output wires
		return node.Message{}, nil
	}

	// Continue mode: pass message through to output wires
	if lastErr != nil {
		return msg, fmt.Errorf("some link deliveries failed: %w", lastErr)
	}

	return msg, nil
}

// Cleanup performs cleanup
func (n *LinkOutNode) Cleanup() error {
	// Nothing to clean up for Link Out
	return nil
}
