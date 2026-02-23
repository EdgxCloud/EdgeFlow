package core

import (
	"context"
	"fmt"
	"math"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// MathNode performs mathematical operations
type MathNode struct {
	operation string  // add, subtract, multiply, divide, mod, pow, sqrt, abs, etc.
	operand   float64 // Second operand for binary operations
	property  string  // Property to operate on
	precision int     // Decimal precision for rounding
}

// Init initializes the math node
func (n *MathNode) Init(config map[string]interface{}) error {
	// Operation
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	} else {
		n.operation = "add"
	}

	// Operand for binary operations
	if operand, ok := config["operand"].(float64); ok {
		n.operand = operand
	} else if operand, ok := config["operand"].(int); ok {
		n.operand = float64(operand)
	} else {
		n.operand = 0
	}

	// Property to operate on
	if prop, ok := config["property"].(string); ok {
		n.property = prop
	} else {
		n.property = "value"
	}

	// Precision for rounding
	if prec, ok := config["precision"].(float64); ok {
		n.precision = int(prec)
	} else if prec, ok := config["precision"].(int); ok {
		n.precision = prec
	} else {
		n.precision = -1 // No rounding by default
	}

	return nil
}

// Execute performs the mathematical operation
func (n *MathNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get input value
	var input float64
	if val, ok := msg.Payload[n.property]; ok {
		input = n.toFloat64(val)
	} else if val, ok := msg.Payload["value"]; ok {
		input = n.toFloat64(val)
	} else {
		return msg, fmt.Errorf("property %s not found in payload", n.property)
	}

	// Perform operation
	var result float64
	var err error

	switch n.operation {
	// Basic arithmetic
	case "add", "+":
		result = input + n.operand
	case "subtract", "-":
		result = input - n.operand
	case "multiply", "*":
		result = input * n.operand
	case "divide", "/":
		if n.operand == 0 {
			return msg, fmt.Errorf("division by zero")
		}
		result = input / n.operand
	case "mod", "%":
		result = math.Mod(input, n.operand)
	case "pow", "power":
		result = math.Pow(input, n.operand)

	// Unary operations
	case "abs":
		result = math.Abs(input)
	case "sqrt":
		if input < 0 {
			return msg, fmt.Errorf("cannot compute square root of negative number")
		}
		result = math.Sqrt(input)
	case "cbrt":
		result = math.Cbrt(input)
	case "exp":
		result = math.Exp(input)
	case "log", "ln":
		if input <= 0 {
			return msg, fmt.Errorf("cannot compute log of non-positive number")
		}
		result = math.Log(input)
	case "log10":
		if input <= 0 {
			return msg, fmt.Errorf("cannot compute log10 of non-positive number")
		}
		result = math.Log10(input)
	case "log2":
		if input <= 0 {
			return msg, fmt.Errorf("cannot compute log2 of non-positive number")
		}
		result = math.Log2(input)
	case "negate", "neg":
		result = -input
	case "inverse", "reciprocal":
		if input == 0 {
			return msg, fmt.Errorf("cannot compute inverse of zero")
		}
		result = 1.0 / input

	// Trigonometric (input in radians)
	case "sin":
		result = math.Sin(input)
	case "cos":
		result = math.Cos(input)
	case "tan":
		result = math.Tan(input)
	case "asin":
		if input < -1 || input > 1 {
			return msg, fmt.Errorf("asin input must be between -1 and 1")
		}
		result = math.Asin(input)
	case "acos":
		if input < -1 || input > 1 {
			return msg, fmt.Errorf("acos input must be between -1 and 1")
		}
		result = math.Acos(input)
	case "atan":
		result = math.Atan(input)
	case "sinh":
		result = math.Sinh(input)
	case "cosh":
		result = math.Cosh(input)
	case "tanh":
		result = math.Tanh(input)

	// Rounding
	case "round":
		result = math.Round(input)
	case "floor":
		result = math.Floor(input)
	case "ceil":
		result = math.Ceil(input)
	case "trunc":
		result = math.Trunc(input)

	// Conversions
	case "deg2rad":
		result = input * math.Pi / 180.0
	case "rad2deg":
		result = input * 180.0 / math.Pi

	// Min/Max with operand
	case "min":
		result = math.Min(input, n.operand)
	case "max":
		result = math.Max(input, n.operand)

	// Clamp between 0 and operand
	case "clamp":
		result = math.Max(0, math.Min(input, n.operand))

	default:
		return msg, fmt.Errorf("unknown operation: %s", n.operation)
	}

	if err != nil {
		return msg, err
	}

	// Apply precision if specified
	if n.precision >= 0 {
		multiplier := math.Pow(10, float64(n.precision))
		result = math.Round(result*multiplier) / multiplier
	}

	// Check for NaN or Inf
	if math.IsNaN(result) {
		return msg, fmt.Errorf("result is NaN")
	}
	if math.IsInf(result, 0) {
		return msg, fmt.Errorf("result is Infinity")
	}

	// Set result
	msg.Payload["value"] = result
	msg.Payload["_math"] = map[string]interface{}{
		"operation": n.operation,
		"input":     input,
		"operand":   n.operand,
		"result":    result,
	}

	return msg, nil
}

// toFloat64 converts various types to float64
func (n *MathNode) toFloat64(val interface{}) float64 {
	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case uint32:
		return float64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	default:
		return 0
	}
}

// Cleanup releases resources
func (n *MathNode) Cleanup() error {
	return nil
}

// NewMathExecutor creates a new math node executor
func NewMathExecutor() node.Executor {
	return &MathNode{}
}

// init registers the math node
func init() {
	registry := node.GetGlobalRegistry()
	registry.Register(&node.NodeInfo{
		Type:        "math",
		Name:        "Math",
		Category:    node.NodeTypeFunction,
		Description: "Perform mathematical operations (arithmetic, trigonometry, rounding)",
		Icon:        "calculator",
		Color:       "#FF6B6B",
		Properties: []node.PropertySchema{
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "add",
				Required:    true,
				Description: "Mathematical operation to perform",
				Options: []string{
					"add", "subtract", "multiply", "divide", "mod", "pow",
					"abs", "sqrt", "cbrt", "exp", "log", "log10", "log2", "negate", "inverse",
					"sin", "cos", "tan", "asin", "acos", "atan", "sinh", "cosh", "tanh",
					"round", "floor", "ceil", "trunc",
					"deg2rad", "rad2deg",
					"min", "max", "clamp",
				},
			},
			{
				Name:        "operand",
				Label:       "Operand",
				Type:        "number",
				Default:     0,
				Required:    false,
				Description: "Second operand for binary operations (add, subtract, multiply, divide, mod, pow, min, max, clamp)",
			},
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "value",
				Required:    false,
				Description: "Property containing the input value",
			},
			{
				Name:        "precision",
				Label:       "Precision",
				Type:        "number",
				Default:     -1,
				Required:    false,
				Description: "Decimal places for rounding (-1 for no rounding)",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "number", Description: "Numeric input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "number", Description: "Result in msg.payload.value"},
		},
		Factory: NewMathExecutor,
	})
}
