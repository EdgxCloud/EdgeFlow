package core

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// JoinNode joins sequences of messages into array, object, or string
type JoinNode struct {
	mode    string // "auto", "manual", "reduce", "merge"
	build   string // "array", "object", "string", "buffer"
	count   int    // Message count for manual mode
	joiner  string // Join character for string mode
	timeout int    // Timeout in seconds

	// Internal state for accumulating messages
	mu       sync.Mutex
	messages []interface{}
	keys     map[string]interface{}
}

// NewJoinNode creates a new join node
func NewJoinNode() *JoinNode {
	return &JoinNode{
		mode:     "auto",
		build:    "array",
		count:    0,
		joiner:   "\n",
		timeout:  0,
		messages: make([]interface{}, 0),
		keys:     make(map[string]interface{}),
	}
}

// Init initializes the join node with configuration
func (n *JoinNode) Init(config map[string]interface{}) error {
	if mode, ok := config["mode"].(string); ok {
		n.mode = mode
	}
	if build, ok := config["build"].(string); ok {
		n.build = build
	}
	if count, ok := config["count"].(float64); ok {
		n.count = int(count)
	}
	if joiner, ok := config["joiner"].(string); ok {
		// Handle escaped newline
		if joiner == "\\n" {
			n.joiner = "\n"
		} else {
			n.joiner = joiner
		}
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = int(timeout)
	}
	return nil
}

// Execute accumulates messages and joins them based on mode
func (n *JoinNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	// Check for split metadata
	splitParts := n.getSplitParts(msg)

	switch n.mode {
	case "auto":
		return n.autoJoin(msg, splitParts)
	case "manual":
		return n.manualJoin(msg)
	case "reduce":
		return n.reduceJoin(msg)
	case "merge":
		return n.mergeJoin(msg)
	default:
		return n.autoJoin(msg, splitParts)
	}
}

// getSplitParts extracts split metadata from message
func (n *JoinNode) getSplitParts(msg node.Message) int {
	if msg.Payload != nil {
		if parts, ok := msg.Payload["_splitParts"].(int); ok {
			return parts
		}
		if parts, ok := msg.Payload["_splitParts"].(float64); ok {
			return int(parts)
		}
	}
	return 0
}

// autoJoin automatically detects when to join based on split metadata
func (n *JoinNode) autoJoin(msg node.Message, splitParts int) (node.Message, error) {
	// Extract actual payload (remove split metadata if present)
	payload := n.extractPayload(msg.Payload)
	n.messages = append(n.messages, payload)

	// Determine expected count
	expectedCount := splitParts
	if expectedCount == 0 {
		expectedCount = n.count
	}

	// If we have enough messages or no count is set, join
	if expectedCount > 0 && len(n.messages) >= expectedCount {
		return n.buildOutput(msg)
	}

	// Not ready yet - return message as-is for now
	// In a full implementation, the engine would hold the message
	return msg, nil
}

// manualJoin joins after receiving specified count of messages
func (n *JoinNode) manualJoin(msg node.Message) (node.Message, error) {
	payload := n.extractPayload(msg.Payload)
	n.messages = append(n.messages, payload)

	if n.count > 0 && len(n.messages) >= n.count {
		return n.buildOutput(msg)
	}

	return msg, nil
}

// reduceJoin reduces messages using an accumulator
func (n *JoinNode) reduceJoin(msg node.Message) (node.Message, error) {
	// For reduce mode, accumulate into the first message
	payload := n.extractPayload(msg.Payload)

	if len(n.messages) == 0 {
		n.messages = append(n.messages, payload)
		return msg, nil
	}

	// Merge payloads if both are maps
	if firstMap, ok := n.messages[0].(map[string]interface{}); ok {
		if currentMap, ok := payload.(map[string]interface{}); ok {
			for k, v := range currentMap {
				firstMap[k] = v
			}
		}
	}

	// Check if we should output
	if n.count > 0 && len(n.messages) >= n.count {
		msg.Payload = map[string]interface{}{"value": n.messages[0]}
		n.messages = make([]interface{}, 0)
		return msg, nil
	}

	return msg, nil
}

// mergeJoin merges object properties from multiple messages
func (n *JoinNode) mergeJoin(msg node.Message) (node.Message, error) {
	payload := n.extractPayload(msg.Payload)

	// Try to extract key-value pair
	if kvMap, ok := payload.(map[string]interface{}); ok {
		if key, hasKey := kvMap["key"].(string); hasKey {
			if value, hasValue := kvMap["value"]; hasValue {
				n.keys[key] = value
			}
		} else {
			// Merge all keys
			for k, v := range kvMap {
				if k != "_splitIndex" && k != "_splitParts" && k != "_splitTotal" {
					n.keys[k] = v
				}
			}
		}
	}

	n.messages = append(n.messages, payload)

	// Check if we should output
	splitParts := n.getSplitParts(msg)
	expectedCount := splitParts
	if expectedCount == 0 {
		expectedCount = n.count
	}

	if expectedCount > 0 && len(n.messages) >= expectedCount {
		msg.Payload = n.keys
		n.messages = make([]interface{}, 0)
		n.keys = make(map[string]interface{})
		return msg, nil
	}

	return msg, nil
}

// extractPayload removes split metadata from payload
func (n *JoinNode) extractPayload(payload map[string]interface{}) interface{} {
	if payload == nil {
		return nil
	}

	// Check if this has split metadata with actual value
	if value, hasValue := payload["value"]; hasValue {
		_, hasSplit := payload["_splitParts"]
		if hasSplit {
			return value
		}
	}

	// Return a copy without split metadata
	result := make(map[string]interface{})
	for k, v := range payload {
		if k != "_splitIndex" && k != "_splitParts" && k != "_splitTotal" {
			result[k] = v
		}
	}
	return result
}

// buildOutput constructs the joined output based on build type
func (n *JoinNode) buildOutput(msg node.Message) (node.Message, error) {
	var payload interface{}

	switch n.build {
	case "array":
		payload = n.messages
	case "string":
		payload = n.joinAsString()
	case "object":
		payload = n.keys
	case "buffer":
		payload = n.joinAsBuffer()
	default:
		payload = n.messages
	}

	msg.Payload = map[string]interface{}{"value": payload}

	// Reset state
	n.messages = make([]interface{}, 0)
	n.keys = make(map[string]interface{})

	return msg, nil
}

// joinAsString joins messages as a string with delimiter
func (n *JoinNode) joinAsString() string {
	parts := make([]string, 0, len(n.messages))
	for _, msg := range n.messages {
		parts = append(parts, fmt.Sprintf("%v", msg))
	}
	return strings.Join(parts, n.joiner)
}

// joinAsBuffer joins messages as a byte buffer
func (n *JoinNode) joinAsBuffer() []byte {
	var buffer []byte
	for _, msg := range n.messages {
		switch v := msg.(type) {
		case []byte:
			buffer = append(buffer, v...)
		case string:
			buffer = append(buffer, []byte(v)...)
		default:
			buffer = append(buffer, []byte(fmt.Sprintf("%v", v))...)
		}
	}
	return buffer
}

// Cleanup cleans up resources
func (n *JoinNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.messages = nil
	n.keys = nil
	return nil
}
