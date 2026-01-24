package core

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

// SwitchNode routes messages based on property values
// Supports multiple rule types: ==, !=, <, >, between, contains, regex
type SwitchNode struct {
	property     string
	propertyType string // "msg", "flow", "global", "jsonata"
	rules        []SwitchRule
	checkAll     bool   // If true, check all rules; if false, stop at first match
	repair       bool   // If true, don't forward messages that don't match any rule
	outputs      int    // Number of outputs (one per rule + optional "otherwise")
}

// SwitchRule defines a single routing rule
type SwitchRule struct {
	Type  string      // "eq", "neq", "lt", "lte", "gt", "gte", "btwn", "cont", "regex", "true", "false", "null", "nnull", "empty", "nempty", "istype", "head", "tail", "index", "jsonata", "else"
	Value interface{} // Comparison value
	Value2 interface{} // Second value for "between" rule
	Case  bool        // Case sensitive for string comparisons
}

// NewSwitchNode creates a new switch node
func NewSwitchNode() *SwitchNode {
	return &SwitchNode{
		property:     "payload",
		propertyType: "msg",
		rules:        make([]SwitchRule, 0),
		checkAll:     true,
		repair:       false,
		outputs:      1,
	}
}

// Init initializes the switch node with configuration
func (n *SwitchNode) Init(config map[string]interface{}) error {
	if property, ok := config["property"].(string); ok {
		n.property = property
	}

	if propertyType, ok := config["propertyType"].(string); ok {
		n.propertyType = propertyType
	}

	if checkAll, ok := config["checkall"].(bool); ok {
		n.checkAll = checkAll
	}

	if repair, ok := config["repair"].(bool); ok {
		n.repair = repair
	}

	// Parse rules
	if rulesConfig, ok := config["rules"].([]interface{}); ok {
		for _, ruleConfig := range rulesConfig {
			if ruleMap, ok := ruleConfig.(map[string]interface{}); ok {
				rule := SwitchRule{
					Type: getStringFromMap(ruleMap, "t", "eq"),
					Case: getBoolFromMap(ruleMap, "case", true),
				}

				if val, ok := ruleMap["v"]; ok {
					rule.Value = val
				}

				if val2, ok := ruleMap["v2"]; ok {
					rule.Value2 = val2
				}

				n.rules = append(n.rules, rule)
			}
		}
	}

	if len(n.rules) == 0 {
		return fmt.Errorf("switch node must have at least one rule")
	}

	n.outputs = len(n.rules)

	return nil
}

// Execute evaluates all rules and routes message to matching outputs
func (n *SwitchNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get property value to test
	testValue, err := n.getPropertyValue(msg)
	if err != nil {
		return msg, fmt.Errorf("failed to get property value: %w", err)
	}

	// Track which outputs should receive the message
	matchedOutputs := make([]bool, n.outputs)

	// Evaluate each rule
	for i, rule := range n.rules {
		matched, err := n.evaluateRule(testValue, rule)
		if err != nil {
			// Log error but continue evaluating other rules
			continue
		}

		if matched {
			matchedOutputs[i] = true
			if !n.checkAll {
				break // Stop at first match
			}
		}
	}

	// Check if any rule matched
	anyMatch := false
	for _, matched := range matchedOutputs {
		if matched {
			anyMatch = true
			break
		}
	}

	// If repair mode and no matches, don't forward
	if n.repair && !anyMatch {
		return msg, nil // Don't forward message
	}

	// TODO: In actual implementation, send to multiple outputs based on matchedOutputs
	// For now, just store matched outputs in message metadata
	if em, ok := interface{}(msg).(*node.EnhancedMessage); ok {
		em.Metadata["_switchOutputs"] = matchedOutputs
	}

	return msg, nil
}

// Cleanup stops the switch node
func (n *SwitchNode) Cleanup() error {
	return nil
}

// evaluateRule checks if a value matches a rule
func (n *SwitchNode) evaluateRule(value interface{}, rule SwitchRule) (bool, error) {
	switch rule.Type {
	case "eq":
		return n.equals(value, rule.Value, rule.Case), nil
	case "neq":
		return !n.equals(value, rule.Value, rule.Case), nil
	case "lt":
		return n.lessThan(value, rule.Value), nil
	case "lte":
		return n.lessThanOrEqual(value, rule.Value), nil
	case "gt":
		return n.greaterThan(value, rule.Value), nil
	case "gte":
		return n.greaterThanOrEqual(value, rule.Value), nil
	case "btwn":
		return n.between(value, rule.Value, rule.Value2), nil
	case "cont":
		return n.contains(value, rule.Value, rule.Case), nil
	case "regex":
		return n.matchesRegex(value, rule.Value), nil
	case "true":
		return n.isTrue(value), nil
	case "false":
		return n.isFalse(value), nil
	case "null":
		return value == nil, nil
	case "nnull":
		return value != nil, nil
	case "empty":
		return n.isEmpty(value), nil
	case "nempty":
		return !n.isEmpty(value), nil
	case "istype":
		return n.isType(value, rule.Value), nil
	case "else":
		return true, nil // Always matches (for "otherwise" output)
	default:
		return false, fmt.Errorf("unknown rule type: %s", rule.Type)
	}
}

// Comparison functions

func (n *SwitchNode) equals(a, b interface{}, caseSensitive bool) bool {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	if !caseSensitive {
		aStr = strings.ToLower(aStr)
		bStr = strings.ToLower(bStr)
	}

	return aStr == bStr
}

func (n *SwitchNode) lessThan(a, b interface{}) bool {
	aNum, aOk := toNumber(a)
	bNum, bOk := toNumber(b)
	if aOk && bOk {
		return aNum < bNum
	}
	return false
}

func (n *SwitchNode) lessThanOrEqual(a, b interface{}) bool {
	aNum, aOk := toNumber(a)
	bNum, bOk := toNumber(b)
	if aOk && bOk {
		return aNum <= bNum
	}
	return false
}

func (n *SwitchNode) greaterThan(a, b interface{}) bool {
	aNum, aOk := toNumber(a)
	bNum, bOk := toNumber(b)
	if aOk && bOk {
		return aNum > bNum
	}
	return false
}

func (n *SwitchNode) greaterThanOrEqual(a, b interface{}) bool {
	aNum, aOk := toNumber(a)
	bNum, bOk := toNumber(b)
	if aOk && bOk {
		return aNum >= bNum
	}
	return false
}

func (n *SwitchNode) between(value, min, max interface{}) bool {
	vNum, vOk := toNumber(value)
	minNum, minOk := toNumber(min)
	maxNum, maxOk := toNumber(max)

	if vOk && minOk && maxOk {
		return vNum >= minNum && vNum <= maxNum
	}
	return false
}

func (n *SwitchNode) contains(haystack, needle interface{}, caseSensitive bool) bool {
	haystackStr := fmt.Sprintf("%v", haystack)
	needleStr := fmt.Sprintf("%v", needle)

	if !caseSensitive {
		haystackStr = strings.ToLower(haystackStr)
		needleStr = strings.ToLower(needleStr)
	}

	return strings.Contains(haystackStr, needleStr)
}

func (n *SwitchNode) matchesRegex(value, pattern interface{}) bool {
	valueStr := fmt.Sprintf("%v", value)
	patternStr := fmt.Sprintf("%v", pattern)

	re, err := regexp.Compile(patternStr)
	if err != nil {
		return false
	}

	return re.MatchString(valueStr)
}

func (n *SwitchNode) isTrue(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	valueStr := strings.ToLower(fmt.Sprintf("%v", value))
	return valueStr == "true" || valueStr == "1"
}

func (n *SwitchNode) isFalse(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return !b
	}
	valueStr := strings.ToLower(fmt.Sprintf("%v", value))
	return valueStr == "false" || valueStr == "0" || valueStr == ""
}

func (n *SwitchNode) isEmpty(value interface{}) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case []interface{}:
		return len(v) == 0
	case map[string]interface{}:
		return len(v) == 0
	default:
		return false
	}
}

func (n *SwitchNode) isType(value, typeName interface{}) bool {
	typeStr := fmt.Sprintf("%v", typeName)

	switch typeStr {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		_, ok := toNumber(value)
		return ok
	case "boolean":
		_, ok := value.(bool)
		return ok
	case "array":
		_, ok := value.([]interface{})
		return ok
	case "object":
		_, ok := value.(map[string]interface{})
		return ok
	case "null":
		return value == nil
	default:
		return false
	}
}

// getPropertyValue retrieves the value to test
func (n *SwitchNode) getPropertyValue(msg node.Message) (interface{}, error) {
	// Handle different property types
	switch n.propertyType {
	case "msg":
		return n.getMessageProperty(msg, n.property)
	case "flow", "global":
		// TODO: Integrate with context store
		return nil, fmt.Errorf("flow/global context not yet implemented")
	default:
		return n.getMessageProperty(msg, n.property)
	}
}

// getMessageProperty extracts a property from the message
func (n *SwitchNode) getMessageProperty(msg node.Message, property string) (interface{}, error) {
	parts := strings.Split(property, ".")

	if len(parts) == 0 {
		return nil, fmt.Errorf("invalid property path: %s", property)
	}

	// Handle top-level properties
	switch parts[0] {
	case "payload":
		if len(parts) == 1 {
			return msg.Payload, nil
		}
		return getNestedProperty(msg.Payload, parts[1:])
	case "topic":
		return msg.Topic, nil
	default:
		// Assume it's a payload property
		return getNestedProperty(msg.Payload, parts)
	}
}

// Helper functions

func getNestedProperty(obj interface{}, path []string) (interface{}, error) {
	if len(path) == 0 {
		return obj, nil
	}

	if m, ok := obj.(map[string]interface{}); ok {
		val, exists := m[path[0]]
		if !exists {
			return nil, fmt.Errorf("property not found: %s", path[0])
		}
		if len(path) == 1 {
			return val, nil
		}
		return getNestedProperty(val, path[1:])
	}

	return nil, fmt.Errorf("cannot traverse non-map type: %T", obj)
}

func toNumber(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func getStringFromMap(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return defaultValue
}

func getBoolFromMap(m map[string]interface{}, key string, defaultValue bool) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return defaultValue
}
