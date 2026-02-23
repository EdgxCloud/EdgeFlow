package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// SetRule represents a single set/delete/move rule
type SetRule struct {
	Type     string      `json:"t"`  // "set", "delete", "move"
	Property string      `json:"p"`  // Target property path
	To       interface{} `json:"to"` // Value to set or destination for move
	ToType   string      `json:"tot"`
}

// SetNode sets, deletes, or moves message properties
// This is a simplified version of Change node focused on set operations
type SetNode struct {
	rules []SetRule
}

// NewSetNode creates a new set node
func NewSetNode() *SetNode {
	return &SetNode{
		rules: []SetRule{},
	}
}

// Init initializes the set node with configuration
func (n *SetNode) Init(config map[string]interface{}) error {
	if rulesRaw, ok := config["rules"].([]interface{}); ok {
		for _, ruleRaw := range rulesRaw {
			if ruleMap, ok := ruleRaw.(map[string]interface{}); ok {
				rule := SetRule{}
				if t, ok := ruleMap["t"].(string); ok {
					rule.Type = t
				}
				if p, ok := ruleMap["p"].(string); ok {
					rule.Property = p
				}
				if to, ok := ruleMap["to"]; ok {
					rule.To = to
				}
				if tot, ok := ruleMap["tot"].(string); ok {
					rule.ToType = tot
				}
				n.rules = append(n.rules, rule)
			}
		}
	}
	return nil
}

// Execute applies set rules to the message
func (n *SetNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}

	for _, rule := range n.rules {
		switch rule.Type {
		case "set":
			n.setProperty(&msg, rule)
		case "delete":
			n.deleteProperty(&msg, rule)
		case "move":
			n.moveProperty(&msg, rule)
		}
	}

	return msg, nil
}

// setProperty sets a property value
func (n *SetNode) setProperty(msg *node.Message, rule SetRule) {
	value := n.resolveValue(rule.To, rule.ToType, msg)
	n.setPath(msg, rule.Property, value)
}

// deleteProperty deletes a property
func (n *SetNode) deleteProperty(msg *node.Message, rule SetRule) {
	path := rule.Property
	if strings.HasPrefix(path, "msg.payload.") {
		path = strings.TrimPrefix(path, "msg.payload.")
		n.deletePath(msg.Payload, path)
	} else if path == "msg.payload" || path == "payload" {
		msg.Payload = make(map[string]interface{})
	} else if strings.HasPrefix(path, "msg.") {
		prop := strings.TrimPrefix(path, "msg.")
		if prop == "topic" {
			msg.Topic = ""
		}
	} else {
		n.deletePath(msg.Payload, path)
	}
}

// moveProperty moves a property to another location
func (n *SetNode) moveProperty(msg *node.Message, rule SetRule) {
	value := n.getPathFull(msg, rule.Property)
	n.deletePropertyPath(msg, rule.Property)
	if toPath, ok := rule.To.(string); ok {
		n.setPath(msg, toPath, value)
	}
}

// resolveValue resolves the value based on type
func (n *SetNode) resolveValue(value interface{}, valueType string, msg *node.Message) interface{} {
	switch valueType {
	case "msg":
		if path, ok := value.(string); ok {
			return n.getPathFull(msg, path)
		}
	case "str":
		return fmt.Sprintf("%v", value)
	case "num":
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case string:
			var f float64
			fmt.Sscanf(v, "%f", &f)
			return f
		}
	case "bool":
		switch v := value.(type) {
		case bool:
			return v
		case string:
			return v == "true"
		}
	case "json":
		return value
	}
	return value
}

// getPathFull gets a value from the full msg context
func (n *SetNode) getPathFull(msg *node.Message, path string) interface{} {
	if strings.HasPrefix(path, "msg.payload.") {
		return n.getPath(msg.Payload, strings.TrimPrefix(path, "msg.payload."))
	} else if path == "msg.payload" || path == "payload" {
		return msg.Payload
	} else if strings.HasPrefix(path, "msg.") {
		prop := strings.TrimPrefix(path, "msg.")
		if prop == "topic" {
			return msg.Topic
		}
	}
	return n.getPath(msg.Payload, path)
}

// setPath sets a value at the given path
func (n *SetNode) setPath(msg *node.Message, path string, value interface{}) {
	if strings.HasPrefix(path, "msg.payload.") {
		n.setPathPayload(msg.Payload, strings.TrimPrefix(path, "msg.payload."), value)
	} else if path == "msg.payload" || path == "payload" {
		if m, ok := value.(map[string]interface{}); ok {
			for k, v := range m {
				msg.Payload[k] = v
			}
		} else {
			msg.Payload["value"] = value
		}
	} else if strings.HasPrefix(path, "msg.") {
		prop := strings.TrimPrefix(path, "msg.")
		if prop == "topic" {
			if s, ok := value.(string); ok {
				msg.Topic = s
			}
		}
	} else {
		n.setPathPayload(msg.Payload, path, value)
	}
}

// getPath gets a value from a nested map
func (n *SetNode) getPath(data map[string]interface{}, path string) interface{} {
	parts := strings.Split(path, ".")
	var current interface{} = data

	for _, part := range parts {
		if m, ok := current.(map[string]interface{}); ok {
			current = m[part]
		} else {
			return nil
		}
	}
	return current
}

// setPathPayload sets a value in a nested map
func (n *SetNode) setPathPayload(data map[string]interface{}, path string, value interface{}) {
	parts := strings.Split(path, ".")
	current := data

	for i, part := range parts {
		if i == len(parts)-1 {
			current[part] = value
		} else {
			if next, ok := current[part].(map[string]interface{}); ok {
				current = next
			} else {
				next := make(map[string]interface{})
				current[part] = next
				current = next
			}
		}
	}
}

// deletePath deletes a value from a nested map
func (n *SetNode) deletePath(data map[string]interface{}, path string) {
	parts := strings.Split(path, ".")
	if len(parts) == 1 {
		delete(data, path)
		return
	}

	current := data
	for i, part := range parts[:len(parts)-1] {
		if next, ok := current[part].(map[string]interface{}); ok {
			if i == len(parts)-2 {
				delete(next, parts[len(parts)-1])
				return
			}
			current = next
		} else {
			return
		}
	}
}

// deletePropertyPath deletes from the full msg context
func (n *SetNode) deletePropertyPath(msg *node.Message, path string) {
	if strings.HasPrefix(path, "msg.payload.") {
		n.deletePath(msg.Payload, strings.TrimPrefix(path, "msg.payload."))
	} else if path == "msg.payload" || path == "payload" {
		msg.Payload = make(map[string]interface{})
	} else if !strings.HasPrefix(path, "msg.") {
		n.deletePath(msg.Payload, path)
	}
}

// Cleanup cleans up resources
func (n *SetNode) Cleanup() error {
	return nil
}
