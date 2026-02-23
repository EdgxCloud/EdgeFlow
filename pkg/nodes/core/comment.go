package core

import (
	"context"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// CommentNode is a display-only node for adding documentation to flows
// It does not process messages - it simply passes them through unchanged
type CommentNode struct {
	text  string // Comment text
	color string // Display color
}

// NewCommentNode creates a new comment node
func NewCommentNode() *CommentNode {
	return &CommentNode{
		text:  "",
		color: "#fbbf24", // Default yellow/amber
	}
}

// Init initializes the comment node with configuration
func (n *CommentNode) Init(config map[string]interface{}) error {
	if text, ok := config["text"].(string); ok {
		n.text = text
	}
	if color, ok := config["color"].(string); ok {
		n.color = color
	}
	return nil
}

// Execute passes the message through unchanged
// Comment nodes are display-only and don't modify messages
func (n *CommentNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Pass through unchanged - comments don't process messages
	return msg, nil
}

// Cleanup cleans up resources
func (n *CommentNode) Cleanup() error {
	return nil
}
