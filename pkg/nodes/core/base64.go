package core

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// Base64Node encodes or decodes Base64 data
type Base64Node struct {
	operation string // encode, decode
	urlSafe   bool   // Use URL-safe Base64 encoding
	property  string // Property to encode/decode
}

// Init initializes the base64 node
func (n *Base64Node) Init(config map[string]interface{}) error {
	// Operation: encode or decode
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	} else {
		n.operation = "encode"
	}

	// URL-safe encoding
	if urlSafe, ok := config["urlSafe"].(bool); ok {
		n.urlSafe = urlSafe
	} else {
		n.urlSafe = false
	}

	// Property to encode/decode
	if prop, ok := config["property"].(string); ok {
		n.property = prop
	} else {
		n.property = "payload"
	}

	return nil
}

// Execute performs base64 encoding or decoding
func (n *Base64Node) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get input data
	var input interface{}
	if n.property == "payload" {
		if val, ok := msg.Payload["value"]; ok {
			input = val
		} else {
			// Use the entire payload as input
			input = msg.Payload
		}
	} else {
		var ok bool
		input, ok = msg.Payload[n.property]
		if !ok {
			return msg, fmt.Errorf("property %s not found in payload", n.property)
		}
	}

	// Get encoder/decoder based on URL-safe setting
	var encoding *base64.Encoding
	if n.urlSafe {
		encoding = base64.URLEncoding
	} else {
		encoding = base64.StdEncoding
	}

	switch n.operation {
	case "encode":
		// Convert input to bytes
		var data []byte
		switch v := input.(type) {
		case []byte:
			data = v
		case string:
			data = []byte(v)
		default:
			data = []byte(fmt.Sprintf("%v", v))
		}

		// Encode
		encoded := encoding.EncodeToString(data)
		msg.Payload["value"] = encoded
		msg.Payload["_base64"] = map[string]interface{}{
			"operation":  "encode",
			"urlSafe":    n.urlSafe,
			"inputBytes": len(data),
			"outputLen":  len(encoded),
		}

	case "decode":
		// Get base64 string
		var encodedStr string
		switch v := input.(type) {
		case string:
			encodedStr = v
		default:
			return msg, fmt.Errorf("decode requires string input, got %T", input)
		}

		// Decode
		decoded, err := encoding.DecodeString(encodedStr)
		if err != nil {
			// Try with padding if it fails
			if n.urlSafe {
				encoding = base64.RawURLEncoding
			} else {
				encoding = base64.RawStdEncoding
			}
			decoded, err = encoding.DecodeString(encodedStr)
			if err != nil {
				return msg, fmt.Errorf("failed to decode base64: %w", err)
			}
		}

		// Try to return as string if it's valid UTF-8, otherwise return bytes
		decodedStr := string(decoded)
		if isValidUTF8(decoded) {
			msg.Payload["value"] = decodedStr
		} else {
			msg.Payload["value"] = decoded
		}
		msg.Payload["_base64"] = map[string]interface{}{
			"operation":   "decode",
			"urlSafe":     n.urlSafe,
			"inputLen":    len(encodedStr),
			"outputBytes": len(decoded),
		}

	default:
		return msg, fmt.Errorf("unknown operation: %s", n.operation)
	}

	return msg, nil
}

// isValidUTF8 checks if bytes are valid UTF-8 text
func isValidUTF8(data []byte) bool {
	// Check for null bytes (binary data often has these)
	for _, b := range data {
		if b == 0 {
			return false
		}
	}
	// Check if it's valid UTF-8 by trying to convert
	s := string(data)
	for _, r := range s {
		if r == '\uFFFD' {
			return false
		}
	}
	return true
}

// Cleanup releases resources
func (n *Base64Node) Cleanup() error {
	return nil
}

// NewBase64Executor creates a new base64 node executor
func NewBase64Executor() node.Executor {
	return &Base64Node{}
}

// init registers the base64 node
func init() {
	registry := node.GetGlobalRegistry()
	registry.Register(&node.NodeInfo{
		Type:        "base64",
		Name:        "Base64",
		Category:    node.NodeTypeFunction,
		Description: "Encode or decode Base64 data",
		Icon:        "code",
		Color:       "#20B2AA",
		Properties: []node.PropertySchema{
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "encode",
				Required:    true,
				Description: "Encode or decode",
				Options:     []string{"encode", "decode"},
			},
			{
				Name:        "urlSafe",
				Label:       "URL Safe",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Use URL-safe Base64 encoding (- and _ instead of + and /)",
			},
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "payload",
				Required:    false,
				Description: "Property to encode/decode",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to encode/decode"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Encoded/decoded data in msg.payload.value"},
		},
		Factory: NewBase64Executor,
	})
}
