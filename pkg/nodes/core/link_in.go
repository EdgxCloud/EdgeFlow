package core

import (
	"context"
	"fmt"
	"sync"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// LinkInNode receives messages from Link Out nodes with matching link ID
// This provides virtual wiring to avoid visual clutter in complex flows
type LinkInNode struct {
	linkID      string
	scope       string // "global" or "flow"
	flowID      string
	outputChan  chan node.Message
	ctx         context.Context
	cancel      context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.RWMutex
}

// NewLinkInNode creates a new Link In node
func NewLinkInNode() *LinkInNode {
	return &LinkInNode{
		scope:      "global",
		outputChan: make(chan node.Message, 100),
	}
}

// Init initializes the Link In node
func (n *LinkInNode) Init(config map[string]interface{}) error {
	// Parse link ID (required)
	if linkID, ok := config["linkId"].(string); ok {
		n.linkID = linkID
	} else {
		return fmt.Errorf("linkId is required for Link In node")
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

	// Validate flow scope has flowID
	if n.scope == "flow" && n.flowID == "" {
		return fmt.Errorf("flowId is required for flow-scoped links")
	}

	return nil
}

// Start begins listening for linked messages
func (n *LinkInNode) Start(ctx context.Context) error {
	n.mu.Lock()
	if n.ctx != nil {
		n.mu.Unlock()
		return fmt.Errorf("Link In node already started")
	}

	n.ctx, n.cancel = context.WithCancel(ctx)
	n.mu.Unlock()

	// Register this node with the link registry
	if err := RegisterLinkIn(n); err != nil {
		return fmt.Errorf("failed to register link in: %w", err)
	}

	return nil
}

// Execute is not used for Link In - messages come from Link Out nodes
func (n *LinkInNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Link In nodes don't process incoming messages from wires
	// They only receive from Link Out nodes via the registry
	return node.Message{}, nil
}

// ReceiveMessage is called by Link Out nodes to send messages
func (n *LinkInNode) ReceiveMessage(msg node.Message) error {
	n.mu.RLock()
	if n.ctx == nil {
		n.mu.RUnlock()
		return fmt.Errorf("link in node not started")
	}
	n.mu.RUnlock()

	select {
	case n.outputChan <- msg:
		return nil
	case <-n.ctx.Done():
		return fmt.Errorf("link in node stopped")
	default:
		// Channel full, drop message or log warning
		return fmt.Errorf("link in output channel full")
	}
}

// GetOutputChannel returns the channel for linked messages
func (n *LinkInNode) GetOutputChannel() <-chan node.Message {
	return n.outputChan
}

// GetLinkID returns the link identifier
func (n *LinkInNode) GetLinkID() string {
	return n.linkID
}

// GetScope returns the link scope
func (n *LinkInNode) GetScope() string {
	return n.scope
}

// GetFlowID returns the flow identifier
func (n *LinkInNode) GetFlowID() string {
	return n.flowID
}

// Cleanup stops the node and unregisters from link registry
func (n *LinkInNode) Cleanup() error {
	if n.cancel != nil {
		n.cancel()
	}

	n.wg.Wait()

	// Unregister from link registry
	UnregisterLinkIn(n)

	if n.outputChan != nil {
		close(n.outputChan)
	}

	return nil
}
