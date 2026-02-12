package core

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

// FilterNode evaluates a condition on incoming messages and routes them
// to either "match" or "no-match" output using the _matchedFilter flag.
type FilterNode struct {
	property      string
	operator      string
	value         interface{}
	valueType     string
	compiledRegex *regexp.Regexp
	value2        interface{}
}

// NewFilterNode creates a new filter node
func NewFilterNode() *FilterNode {
	return &FilterNode{
		property:  "payload",
		operator:  "equals",
		valueType: "string",
	}
}

// Init initializes the filter node with configuration
func (n *FilterNode) Init(config map[string]interface{}) error {
	if property, ok := config["property"].(string); ok {
		n.property = property
	}
	if operator, ok := config["operator"].(string); ok {
		n.operator = operator
	}
	if value, ok := config["value"]; ok {
		n.value = value
	}
	if valueType, ok := config["valueType"].(string); ok {
		n.valueType = valueType
	}
	if value2, ok := config["value2"]; ok {
		n.value2 = value2
	}

	if n.property == "" {
		return fmt.Errorf("filter node requires a property to check")
	}
	if n.operator == "" {
		return fmt.Errorf("filter node requires an operator")
	}

	validOperators := map[string]bool{
		"equals": true, "not_equals": true,
		"gt": true, "lt": true, "gte": true, "lte": true,
		"contains": true, "not_contains": true,
		"regex":      true,
		"is_true":    true, "is_false": true,
		"is_null":    true, "is_not_null": true,
		"between":    true,
	}
	if !validOperators[n.operator] {
		return fmt.Errorf("unknown filter operator: %s", n.operator)
	}

	if n.operator == "regex" {
		pattern, ok := n.value.(string)
		if !ok {
			return fmt.Errorf("regex operator requires a string pattern")
		}
		compiled, err := regexp.Compile(pattern)
		if err != nil {
			return fmt.Errorf("invalid regex pattern %q: %w", pattern, err)
		}
		n.compiledRegex = compiled
	}

	if n.operator == "between" {
		if n.value == nil || n.value2 == nil {
			return fmt.Errorf("between operator requires both value and value2")
		}
	}

	return nil
}

// Execute evaluates the filter condition and sets the _matchedFilter flag
func (n *FilterNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}

	propValue, found := filterGetNestedValue(msg.Payload, n.property)

	matched, err := n.evaluate(propValue, found)
	if err != nil {
		return msg, fmt.Errorf("filter evaluation error: %w", err)
	}

	msg.Payload["_matchedFilter"] = matched
	return msg, nil
}

// Cleanup releases resources
func (n *FilterNode) Cleanup() error {
	return nil
}

func (n *FilterNode) evaluate(propValue interface{}, found bool) (bool, error) {
	switch n.operator {
	case "equals":
		return n.evalEquals(propValue, n.value), nil
	case "not_equals":
		return !n.evalEquals(propValue, n.value), nil
	case "gt":
		return n.evalNumericCompare(propValue, n.value, func(a, b float64) bool { return a > b })
	case "lt":
		return n.evalNumericCompare(propValue, n.value, func(a, b float64) bool { return a < b })
	case "gte":
		return n.evalNumericCompare(propValue, n.value, func(a, b float64) bool { return a >= b })
	case "lte":
		return n.evalNumericCompare(propValue, n.value, func(a, b float64) bool { return a <= b })
	case "contains":
		return n.evalContains(propValue, n.value), nil
	case "not_contains":
		return !n.evalContains(propValue, n.value), nil
	case "regex":
		return n.evalRegex(propValue), nil
	case "is_true":
		return n.evalIsTrue(propValue), nil
	case "is_false":
		return n.evalIsFalse(propValue), nil
	case "is_null":
		return !found || propValue == nil, nil
	case "is_not_null":
		return found && propValue != nil, nil
	case "between":
		return n.evalBetween(propValue, n.value, n.value2)
	default:
		return false, fmt.Errorf("unknown operator: %s", n.operator)
	}
}

func (n *FilterNode) evalEquals(a, b interface{}) bool {
	aNum, aOk := filterToFloat64(a)
	bNum, bOk := filterToFloat64(b)
	if aOk && bOk {
		return aNum == bNum
	}
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func (n *FilterNode) evalNumericCompare(a, b interface{}, cmp func(float64, float64) bool) (bool, error) {
	aNum, aOk := filterToFloat64(a)
	bNum, bOk := filterToFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("cannot perform numeric comparison: values are not numeric")
	}
	return cmp(aNum, bNum), nil
}

func (n *FilterNode) evalContains(a, b interface{}) bool {
	return strings.Contains(fmt.Sprintf("%v", a), fmt.Sprintf("%v", b))
}

func (n *FilterNode) evalRegex(value interface{}) bool {
	if n.compiledRegex == nil {
		return false
	}
	return n.compiledRegex.MatchString(fmt.Sprintf("%v", value))
}

func (n *FilterNode) evalIsTrue(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	s := strings.ToLower(fmt.Sprintf("%v", value))
	return s == "true" || s == "1"
}

func (n *FilterNode) evalIsFalse(value interface{}) bool {
	if b, ok := value.(bool); ok {
		return !b
	}
	s := strings.ToLower(fmt.Sprintf("%v", value))
	return s == "false" || s == "0" || s == ""
}

func (n *FilterNode) evalBetween(value, min, max interface{}) (bool, error) {
	vNum, vOk := filterToFloat64(value)
	minNum, minOk := filterToFloat64(min)
	maxNum, maxOk := filterToFloat64(max)
	if !vOk || !minOk || !maxOk {
		return false, fmt.Errorf("between operator requires numeric values")
	}
	return vNum >= minNum && vNum <= maxNum, nil
}

func filterGetNestedValue(payload map[string]interface{}, path string) (interface{}, bool) {
	parts := strings.Split(path, ".")
	var current interface{} = payload
	for _, part := range parts {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		val, exists := m[part]
		if !exists {
			return nil, false
		}
		current = val
	}
	return current, true
}

func filterToFloat64(v interface{}) (float64, bool) {
	switch val := v.(type) {
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int32:
		return float64(val), true
	case int64:
		return float64(val), true
	case uint:
		return float64(val), true
	case uint32:
		return float64(val), true
	case uint64:
		return float64(val), true
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}
