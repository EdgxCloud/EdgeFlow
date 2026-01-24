package network

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

type JSONParserNode struct {
	action   string // "parse" or "stringify"
	property string // Property to parse/stringify
	target   string // Where to store result
}

func NewJSONParserNode() *JSONParserNode {
	return &JSONParserNode{
		action:   "parse",
		property: "payload",
		target:   "payload",
	}
}

func (n *JSONParserNode) Init(config map[string]interface{}) error {
	if action, ok := config["action"].(string); ok {
		n.action = action
	}
	if property, ok := config["property"].(string); ok {
		n.property = property
	}
	if target, ok := config["target"].(string); ok {
		n.target = target
	}
	return nil
}

func (n *JSONParserNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	switch n.action {
	case "parse":
		return n.parseJSON(msg)
	case "stringify":
		return n.stringifyJSON(msg)
	default:
		return msg, fmt.Errorf("unknown action: %s", n.action)
	}
}

func (n *JSONParserNode) parseJSON(msg node.Message) (node.Message, error) {
	var jsonStr string

	// Handle data from payload
	if str, ok := msg.Payload["data"].(string); ok {
		jsonStr = str
	} else if bytes, ok := msg.Payload["data"].([]byte); ok {
		jsonStr = string(bytes)
	} else {
		return msg, fmt.Errorf("payload must contain 'data' as string or []byte")
	}

	var result interface{}
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return msg, fmt.Errorf("failed to parse JSON: %w", err)
	}

	msg.Payload = map[string]interface{}{"data": result}
	return msg, nil
}

func (n *JSONParserNode) stringifyJSON(msg node.Message) (node.Message, error) {
	var dataToStringify interface{}

	// Extract data from payload
	if data, ok := msg.Payload["data"]; ok {
		dataToStringify = data
	} else {
		dataToStringify = msg.Payload
	}

	data, err := json.Marshal(dataToStringify)
	if err != nil {
		return msg, fmt.Errorf("failed to stringify JSON: %w", err)
	}

	msg.Payload = map[string]interface{}{"data": string(data)}
	return msg, nil
}

func (n *JSONParserNode) Cleanup() error {
	return nil
}

func NewJSONParserExecutor() node.Executor {
	return NewJSONParserNode()
}
