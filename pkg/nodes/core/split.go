package core

import (
	"context"
	"strings"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// SplitNode splits arrays, objects, or strings into separate messages
type SplitNode struct {
	arraySplt      bool   // Whether to split arrays
	arraySplitType string // "each" or "len"
	arraySpltLen   int    // Length of each segment when using "len"
	strSplit       string // Delimiter for string splitting
}

// NewSplitNode creates a new split node
func NewSplitNode() *SplitNode {
	return &SplitNode{
		arraySplt:      true,
		arraySplitType: "each",
		arraySpltLen:   1,
		strSplit:       "\n",
	}
}

// Init initializes the split node with configuration
func (n *SplitNode) Init(config map[string]interface{}) error {
	if arraySplt, ok := config["arraySplt"].(bool); ok {
		n.arraySplt = arraySplt
	}
	if arraySplitType, ok := config["arraySplitType"].(string); ok {
		n.arraySplitType = arraySplitType
	}
	if arraySpltLen, ok := config["arraySpltLen"].(float64); ok {
		n.arraySpltLen = int(arraySpltLen)
	}
	if strSplit, ok := config["strSplit"].(string); ok {
		n.strSplit = strSplit
	}
	return nil
}

// Execute splits the message payload
// Note: This returns the first part; in a real implementation,
// the engine would need to handle multiple output messages
func (n *SplitNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if msg.Payload == nil {
		return msg, nil
	}

	// Check for array in payload
	if arr, ok := msg.Payload["value"].([]interface{}); ok {
		return n.splitArray(msg, arr)
	}

	// Check for string in payload
	if str, ok := msg.Payload["value"].(string); ok {
		return n.splitString(msg, str)
	}

	// If payload itself is usable as array/string
	return msg, nil
}

// splitArray splits an array into separate messages
func (n *SplitNode) splitArray(msg node.Message, arr []interface{}) (node.Message, error) {
	if !n.arraySplt || len(arr) == 0 {
		return msg, nil
	}

	// Store split metadata for potential join operation
	partCount := len(arr)
	if n.arraySplitType == "len" && n.arraySpltLen > 0 {
		partCount = (len(arr) + n.arraySpltLen - 1) / n.arraySpltLen
	}

	if n.arraySplitType == "each" {
		// Split into individual elements
		// Return first element, mark with split metadata
		msg.Payload = map[string]interface{}{
			"value":       arr[0],
			"_splitIndex": 0,
			"_splitParts": partCount,
			"_splitTotal": len(arr),
		}
		return msg, nil
	}

	// Split into chunks of specified length
	if n.arraySpltLen <= 0 {
		n.arraySpltLen = 1
	}

	chunk := make([]interface{}, 0, n.arraySpltLen)
	for i := 0; i < n.arraySpltLen && i < len(arr); i++ {
		chunk = append(chunk, arr[i])
	}

	msg.Payload = map[string]interface{}{
		"value":       chunk,
		"_splitIndex": 0,
		"_splitParts": partCount,
		"_splitTotal": len(arr),
	}
	return msg, nil
}

// splitString splits a string by delimiter
func (n *SplitNode) splitString(msg node.Message, str string) (node.Message, error) {
	delimiter := n.strSplit
	if delimiter == "" {
		delimiter = "\n"
	}

	parts := strings.Split(str, delimiter)
	if len(parts) == 0 {
		return msg, nil
	}

	msg.Payload = map[string]interface{}{
		"value":       parts[0],
		"_splitIndex": 0,
		"_splitParts": len(parts),
		"_splitTotal": len(parts),
	}
	return msg, nil
}

// Cleanup cleans up resources
func (n *SplitNode) Cleanup() error {
	return nil
}
