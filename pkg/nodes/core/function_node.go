package core

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/edgeflow/edgeflow/internal/node"
)

// FunctionNodeExecutor provides enhanced function execution with expression evaluation
type FunctionNodeExecutor struct {
	code    string
	outputs int
	noerr   bool
}

// NewFunctionNodeExecutor creates a new enhanced function node executor
func NewFunctionNodeExecutor() node.Executor {
	return &FunctionNodeExecutor{
		outputs: 1,
	}
}

// Init initializes the function node executor
func (n *FunctionNodeExecutor) Init(config map[string]interface{}) error {
	if code, ok := config["code"].(string); ok {
		n.code = code
	}
	if outputs, ok := config["outputs"].(float64); ok {
		n.outputs = int(outputs)
		if n.outputs < 1 {
			n.outputs = 1
		}
	}
	if noerr, ok := config["noerr"].(bool); ok {
		n.noerr = noerr
	}

	if n.code == "" {
		return fmt.Errorf("function-node code cannot be empty")
	}
	return nil
}

// Execute processes the message through the function code
func (n *FunctionNodeExecutor) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if msg.Payload == nil {
		msg.Payload = make(map[string]interface{})
	}

	// Process code lines
	lines := strings.Split(n.code, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		if err := n.processLine(line, msg.Payload); err != nil {
			if n.noerr {
				continue
			}
			return msg, fmt.Errorf("function-node error: %w", err)
		}
	}

	msg.Payload["_functionNodeProcessed"] = true
	return msg, nil
}

// processLine handles a single line of function code
func (n *FunctionNodeExecutor) processLine(line string, payload map[string]interface{}) error {
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

// Cleanup releases resources
func (n *FunctionNodeExecutor) Cleanup() error {
	return nil
}
