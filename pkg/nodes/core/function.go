package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

// FunctionRule represents a single transformation rule
type FunctionRule struct {
	Action    string      `json:"action"`    // "set" or "delete"
	Property  string      `json:"property"`  // target key in msg.Payload
	ValueType string      `json:"valueType"` // "string","number","boolean","json","msg","expression"
	Value     interface{} `json:"value"`     // the value to set
}

// FunctionNode transforms message payloads using rules or legacy DSL code.
// New config uses a rules array; legacy config with "code" string is still supported.
type FunctionNode struct {
	rules    []FunctionRule
	code     string // legacy backward compat
	noerr    bool
	useRules bool
}

// NewFunctionNode creates a new function node
func NewFunctionNode() *FunctionNode {
	return &FunctionNode{}
}

// Init initializes the function node from configuration.
// Accepts either {"rules": [...]} (new) or {"code": "..."} (legacy).
func (n *FunctionNode) Init(config map[string]interface{}) error {
	// Check for noerr flag
	if noerr, ok := config["noerr"].(bool); ok {
		n.noerr = noerr
	}

	// Try rules-based config first
	if rulesRaw, ok := config["rules"]; ok {
		n.useRules = true
		return n.parseRules(rulesRaw)
	}

	// Fall back to legacy DSL code
	if code, ok := config["code"].(string); ok && code != "" {
		n.code = code
		n.useRules = false
		return nil
	}

	// Empty config = pass-through (no error)
	n.useRules = true
	n.rules = []FunctionRule{}
	return nil
}

// parseRules converts raw config rules into typed FunctionRule structs
func (n *FunctionNode) parseRules(rulesRaw interface{}) error {
	rulesSlice, ok := rulesRaw.([]interface{})
	if !ok {
		// Could be pre-marshaled []FunctionRule via JSON round-trip
		data, err := json.Marshal(rulesRaw)
		if err != nil {
			return fmt.Errorf("invalid rules format: %w", err)
		}
		return json.Unmarshal(data, &n.rules)
	}

	n.rules = make([]FunctionRule, 0, len(rulesSlice))
	for _, ruleRaw := range rulesSlice {
		ruleMap, ok := ruleRaw.(map[string]interface{})
		if !ok {
			continue
		}

		rule := FunctionRule{}
		if action, ok := ruleMap["action"].(string); ok {
			rule.Action = action
		}
		if prop, ok := ruleMap["property"].(string); ok {
			rule.Property = prop
		}
		if vt, ok := ruleMap["valueType"].(string); ok {
			rule.ValueType = vt
		}
		if val, exists := ruleMap["value"]; exists {
			rule.Value = val
		}

		if rule.Action != "" && rule.Property != "" {
			n.rules = append(n.rules, rule)
		}
	}
	return nil
}

// Execute processes the message through rules or legacy DSL code
func (n *FunctionNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}

	if n.useRules {
		return n.executeRules(msg)
	}
	return n.executeLegacyCode(msg)
}

// executeRules applies typed rules to the message payload
func (n *FunctionNode) executeRules(msg node.Message) (node.Message, error) {
	for _, rule := range n.rules {
		switch rule.Action {
		case "set":
			val := n.resolveRuleValue(rule, msg.Payload)
			msg.Payload[rule.Property] = val
		case "delete":
			delete(msg.Payload, rule.Property)
		}
	}
	return msg, nil
}

// resolveRuleValue converts a rule's value based on its type
func (n *FunctionNode) resolveRuleValue(rule FunctionRule, payload map[string]interface{}) interface{} {
	switch rule.ValueType {
	case "string":
		return fmt.Sprintf("%v", rule.Value)
	case "number":
		return fnToFloat64(rule.Value)
	case "boolean":
		return fnToBool(rule.Value)
	case "json":
		// Already parsed by JSON unmarshaling; return as-is
		return rule.Value
	case "msg":
		// Reference another payload key
		if key, ok := rule.Value.(string); ok {
			if val, exists := payload[key]; exists {
				return val
			}
		}
		return nil
	case "expression":
		// Use legacy DSL parser for arithmetic expressions
		if expr, ok := rule.Value.(string); ok {
			return fnParseValue(expr, payload)
		}
		return rule.Value
	default:
		return rule.Value
	}
}

// executeLegacyCode processes legacy DSL code line by line
func (n *FunctionNode) executeLegacyCode(msg node.Message) (node.Message, error) {
	lines := strings.Split(n.code, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if err := processLine(line, msg.Payload); err != nil {
			if n.noerr {
				continue
			}
			return msg, fmt.Errorf("function error: %w", err)
		}
	}
	return msg, nil
}

// Cleanup releases resources
func (n *FunctionNode) Cleanup() error {
	return nil
}

// processLine handles a single line of legacy DSL code
func processLine(line string, payload map[string]interface{}) error {
	// Handle: msg.payload.key = value
	if strings.HasPrefix(line, "msg.payload.") && strings.Contains(line, "=") {
		parts := strings.SplitN(line[len("msg.payload."):], "=", 2)
		if len(parts) != 2 {
			return nil
		}
		key := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])
		valueStr = strings.TrimSuffix(valueStr, ";")
		payload[key] = fnParseValue(valueStr, payload)
		return nil
	}

	// Handle: set key = value
	if strings.HasPrefix(line, "set ") {
		rest := strings.TrimPrefix(line, "set ")
		parts := strings.SplitN(rest, "=", 2)
		if len(parts) != 2 {
			return nil
		}
		key := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])
		valueStr = strings.TrimSuffix(valueStr, ";")
		payload[key] = fnParseValue(valueStr, payload)
		return nil
	}

	// Handle: delete key
	if strings.HasPrefix(line, "delete ") {
		key := strings.TrimSpace(strings.TrimPrefix(line, "delete "))
		key = strings.TrimSuffix(key, ";")
		key = strings.TrimPrefix(key, "msg.payload.")
		delete(payload, key)
		return nil
	}

	// Handle: return msg (pass-through)
	if strings.HasPrefix(line, "return") {
		return nil
	}

	return nil
}

// fnParseValue parses a string value into the appropriate Go type
func fnParseValue(s string, payload map[string]interface{}) interface{} {
	s = strings.TrimSpace(s)

	// Boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Null
	if s == "null" || s == "nil" {
		return nil
	}

	// Reference to another payload value: msg.payload.X
	if strings.HasPrefix(s, "msg.payload.") {
		key := strings.TrimPrefix(s, "msg.payload.")
		if val, ok := payload[key]; ok {
			return val
		}
		return nil
	}

	// String literal
	if (strings.HasPrefix(s, "\"") && strings.HasSuffix(s, "\"")) ||
		(strings.HasPrefix(s, "'") && strings.HasSuffix(s, "'")) {
		return s[1 : len(s)-1]
	}

	// Try JSON (number, object, array)
	var jsonVal interface{}
	if err := json.Unmarshal([]byte(s), &jsonVal); err == nil {
		return jsonVal
	}

	// Simple arithmetic: value + N or value - N or value * N or value / N
	for _, op := range []string{" + ", " - ", " * ", " / "} {
		if strings.Contains(s, op) {
			parts := strings.SplitN(s, op, 2)
			left := fnParseValue(parts[0], payload)
			right := fnParseValue(parts[1], payload)
			lf, lok := fnToFloat(left)
			rf, rok := fnToFloat(right)
			if lok && rok {
				switch op {
				case " + ":
					return lf + rf
				case " - ":
					return lf - rf
				case " * ":
					return lf * rf
				case " / ":
					if rf != 0 {
						return lf / rf
					}
					return 0.0
				}
			}
		}
	}

	// Return as string
	return s
}

func fnToFloat(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case json.Number:
		f, err := val.Float64()
		return f, err == nil
	default:
		return 0, false
	}
}

// fnToFloat64 converts an interface value to float64
func fnToFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case json.Number:
		f, _ := val.Float64()
		return f
	case string:
		var f float64
		if err := json.Unmarshal([]byte(val), &f); err == nil {
			return f
		}
		return 0
	default:
		return 0
	}
}

// fnToBool converts an interface value to bool
func fnToBool(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true" || val == "1"
	case float64:
		return val != 0
	case int:
		return val != 0
	default:
		return false
	}
}

// Helper function to safely get values from payload
func getPayloadValue(payload map[string]interface{}, key string) (interface{}, error) {
	if value, ok := payload[key]; ok {
		return value, nil
	}
	return nil, fmt.Errorf("key %s not found in payload", key)
}

// Helper function to convert to JSON string
func toJSON(v interface{}) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}
