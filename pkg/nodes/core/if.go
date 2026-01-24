package core

import (
	"context"
	"fmt"

	"github.com/edgeflow/edgeflow/internal/node"
)

// IfNode conditionally routes messages based on a condition
type IfNode struct {
	condition string // JavaScript expression or simple comparison
	operator  string // eq, ne, gt, lt, gte, lte, contains, etc.
	value     interface{}
	field     string
}

// NewIfNode creates a new if node
func NewIfNode() *IfNode {
	return &IfNode{
		operator: "eq",
	}
}

// Init initializes the if node with configuration
func (n *IfNode) Init(config map[string]interface{}) error {
	if field, ok := config["field"].(string); ok {
		n.field = field
	}

	if operator, ok := config["operator"].(string); ok {
		n.operator = operator
	}

	if value, ok := config["value"]; ok {
		n.value = value
	}

	if condition, ok := config["condition"].(string); ok {
		n.condition = condition
	}

	return nil
}

// Execute evaluates the condition and routes the message
func (n *IfNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	result, err := n.evaluateCondition(msg)
	if err != nil {
		return msg, err
	}

	// Add condition result to payload
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}
	msg.Payload["_condition_result"] = result
	msg.Payload["_branch"] = map[string]bool{
		"true":  result,
		"false": !result,
	}

	return msg, nil
}

// Cleanup stops the if node
func (n *IfNode) Cleanup() error {
	return nil
}

// evaluateCondition evaluates the configured condition
func (n *IfNode) evaluateCondition(msg node.Message) (bool, error) {
	// Get the field value from payload
	var fieldValue interface{}
	if n.field != "" {
		var ok bool
		fieldValue, ok = msg.Payload[n.field]
		if !ok {
			return false, fmt.Errorf("field %s not found in payload", n.field)
		}
	}

	// Evaluate based on operator
	switch n.operator {
	case "eq": // Equal
		return compareEqual(fieldValue, n.value), nil
	case "ne": // Not equal
		return !compareEqual(fieldValue, n.value), nil
	case "gt": // Greater than
		return compareGreaterThan(fieldValue, n.value)
	case "lt": // Less than
		return compareLessThan(fieldValue, n.value)
	case "gte": // Greater than or equal
		return compareGreaterThanOrEqual(fieldValue, n.value)
	case "lte": // Less than or equal
		return compareLessThanOrEqual(fieldValue, n.value)
	case "contains": // String contains
		return compareContains(fieldValue, n.value)
	case "exists": // Field exists
		return fieldValue != nil, nil
	default:
		return false, fmt.Errorf("unknown operator: %s", n.operator)
	}
}

// Helper comparison functions
func compareEqual(a, b interface{}) bool {
	return fmt.Sprintf("%v", a) == fmt.Sprintf("%v", b)
}

func compareGreaterThan(a, b interface{}) (bool, error) {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}
	return aFloat > bFloat, nil
}

func compareLessThan(a, b interface{}) (bool, error) {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}
	return aFloat < bFloat, nil
}

func compareGreaterThanOrEqual(a, b interface{}) (bool, error) {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}
	return aFloat >= bFloat, nil
}

func compareLessThanOrEqual(a, b interface{}) (bool, error) {
	aFloat, aOk := toFloat64(a)
	bFloat, bOk := toFloat64(b)
	if !aOk || !bOk {
		return false, fmt.Errorf("cannot compare non-numeric values")
	}
	return aFloat <= bFloat, nil
}

func compareContains(a, b interface{}) (bool, error) {
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)
	return len(aStr) > 0 && len(bStr) > 0 && contains(aStr, bStr), nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// toFloat64 moved to utils.go to avoid duplication
