package network

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
	"gopkg.in/yaml.v3"
)

type YAMLParserNode struct {
	action string
}

func NewYAMLParserNode() *YAMLParserNode {
	return &YAMLParserNode{
		action: "parse",
	}
}

func (n *YAMLParserNode) Init(config map[string]interface{}) error {
	if action, ok := config["action"].(string); ok {
		n.action = action
	}
	return nil
}

func (n *YAMLParserNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	switch n.action {
	case "parse":
		return n.parseYAML(msg)
	case "stringify":
		return n.stringifyYAML(msg)
	default:
		return msg, fmt.Errorf("unknown action: %s", n.action)
	}
}

func (n *YAMLParserNode) parseYAML(msg node.Message) (node.Message, error) {
	var yamlStr string

	// Handle data from payload
	if str, ok := msg.Payload["data"].(string); ok {
		yamlStr = str
	} else if bytes, ok := msg.Payload["data"].([]byte); ok {
		yamlStr = string(bytes)
	} else {
		return msg, fmt.Errorf("payload must contain 'data' as string or []byte")
	}

	var result interface{}
	if err := yaml.Unmarshal([]byte(yamlStr), &result); err != nil {
		return msg, fmt.Errorf("failed to parse YAML: %w", err)
	}

	msg.Payload = map[string]interface{}{"data": result}
	return msg, nil
}

func (n *YAMLParserNode) stringifyYAML(msg node.Message) (node.Message, error) {
	var dataToStringify interface{}

	// Extract data from payload
	if data, ok := msg.Payload["data"]; ok {
		dataToStringify = data
	} else {
		dataToStringify = msg.Payload
	}

	data, err := yaml.Marshal(dataToStringify)
	if err != nil {
		return msg, fmt.Errorf("failed to stringify YAML: %w", err)
	}

	msg.Payload = map[string]interface{}{"data": string(data)}
	return msg, nil
}

func (n *YAMLParserNode) Cleanup() error {
	return nil
}

func NewYAMLParserExecutor() node.Executor {
	return NewYAMLParserNode()
}
