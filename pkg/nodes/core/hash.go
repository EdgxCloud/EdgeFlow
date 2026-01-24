package core

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"

	"github.com/edgeflow/edgeflow/internal/node"
)

// HashNode computes cryptographic hashes
type HashNode struct {
	algorithm string // md5, sha1, sha256, sha512, sha384, hmac-sha256, hmac-sha512
	encoding  string // hex, base64, binary
	key       string // HMAC key (for HMAC algorithms)
	property  string // Property to hash (default: payload)
}

// Init initializes the hash node
func (n *HashNode) Init(config map[string]interface{}) error {
	// Algorithm
	if algo, ok := config["algorithm"].(string); ok {
		n.algorithm = algo
	} else {
		n.algorithm = "sha256" // Default SHA-256
	}

	// Encoding
	if enc, ok := config["encoding"].(string); ok {
		n.encoding = enc
	} else {
		n.encoding = "hex" // Default hex encoding
	}

	// HMAC key
	if key, ok := config["key"].(string); ok {
		n.key = key
	}

	// Property to hash
	if prop, ok := config["property"].(string); ok {
		n.property = prop
	} else {
		n.property = "payload"
	}

	return nil
}

// Execute computes the hash of the input
func (n *HashNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get input data
	var data []byte
	var err error

	if n.property == "payload" {
		// Hash the entire payload
		data, err = n.getPayloadBytes(msg.Payload)
	} else {
		// Hash a specific property
		if val, ok := msg.Payload[n.property]; ok {
			data, err = n.valueToBytes(val)
		} else {
			return msg, fmt.Errorf("property %s not found in payload", n.property)
		}
	}

	if err != nil {
		return msg, fmt.Errorf("failed to get data for hashing: %w", err)
	}

	// Compute hash
	var hashBytes []byte
	switch n.algorithm {
	case "md5":
		h := md5.Sum(data)
		hashBytes = h[:]
	case "sha1":
		h := sha1.Sum(data)
		hashBytes = h[:]
	case "sha256":
		h := sha256.Sum256(data)
		hashBytes = h[:]
	case "sha384":
		h := sha512.Sum384(data)
		hashBytes = h[:]
	case "sha512":
		h := sha512.Sum512(data)
		hashBytes = h[:]
	case "hmac-md5":
		hashBytes = n.computeHMAC(md5.New, data)
	case "hmac-sha1":
		hashBytes = n.computeHMAC(sha1.New, data)
	case "hmac-sha256":
		hashBytes = n.computeHMAC(sha256.New, data)
	case "hmac-sha384":
		hashBytes = n.computeHMAC(sha512.New384, data)
	case "hmac-sha512":
		hashBytes = n.computeHMAC(sha512.New, data)
	default:
		return msg, fmt.Errorf("unsupported algorithm: %s", n.algorithm)
	}

	// Encode result
	var result interface{}
	switch n.encoding {
	case "hex":
		result = hex.EncodeToString(hashBytes)
	case "base64":
		result = base64.StdEncoding.EncodeToString(hashBytes)
	case "base64url":
		result = base64.URLEncoding.EncodeToString(hashBytes)
	case "binary":
		result = hashBytes
	default:
		result = hex.EncodeToString(hashBytes)
	}

	// Set result
	msg.Payload["hash"] = result
	msg.Payload["_hashInfo"] = map[string]interface{}{
		"algorithm":  n.algorithm,
		"encoding":   n.encoding,
		"inputBytes": len(data),
		"hashBytes":  len(hashBytes),
	}

	return msg, nil
}

// computeHMAC computes HMAC with the given hash function
func (n *HashNode) computeHMAC(hashFunc func() hash.Hash, data []byte) []byte {
	h := hmac.New(hashFunc, []byte(n.key))
	h.Write(data)
	return h.Sum(nil)
}

// getPayloadBytes converts payload to bytes for hashing
func (n *HashNode) getPayloadBytes(payload map[string]interface{}) ([]byte, error) {
	// Check for specific value types in payload
	if val, ok := payload["value"]; ok {
		return n.valueToBytes(val)
	}
	// Otherwise serialize the whole payload as JSON-like string
	return []byte(fmt.Sprintf("%v", payload)), nil
}

// valueToBytes converts a value to bytes
func (n *HashNode) valueToBytes(val interface{}) ([]byte, error) {
	switch v := val.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case int:
		return []byte(fmt.Sprintf("%d", v)), nil
	case int64:
		return []byte(fmt.Sprintf("%d", v)), nil
	case float64:
		return []byte(fmt.Sprintf("%f", v)), nil
	case bool:
		return []byte(fmt.Sprintf("%t", v)), nil
	default:
		return []byte(fmt.Sprintf("%v", v)), nil
	}
}

// Cleanup releases resources
func (n *HashNode) Cleanup() error {
	return nil
}

// NewHashExecutor creates a new hash node executor
func NewHashExecutor() node.Executor {
	return &HashNode{}
}

// init registers the hash node
func init() {
	registry := node.GetGlobalRegistry()
	registry.Register(&node.NodeInfo{
		Type:        "hash",
		Name:        "Hash",
		Category:    node.NodeTypeFunction,
		Description: "Compute cryptographic hash (MD5, SHA1, SHA256, SHA512, HMAC)",
		Icon:        "key",
		Color:       "#9370DB",
		Properties: []node.PropertySchema{
			{
				Name:        "algorithm",
				Label:       "Algorithm",
				Type:        "select",
				Default:     "sha256",
				Required:    true,
				Description: "Hash algorithm to use",
				Options:     []string{"md5", "sha1", "sha256", "sha384", "sha512", "hmac-md5", "hmac-sha1", "hmac-sha256", "hmac-sha384", "hmac-sha512"},
			},
			{
				Name:        "encoding",
				Label:       "Output Encoding",
				Type:        "select",
				Default:     "hex",
				Required:    true,
				Description: "Output encoding format",
				Options:     []string{"hex", "base64", "base64url", "binary"},
			},
			{
				Name:        "key",
				Label:       "HMAC Key",
				Type:        "password",
				Default:     "",
				Required:    false,
				Description: "Secret key for HMAC algorithms",
			},
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "payload",
				Required:    false,
				Description: "Property to hash (default: entire payload)",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to hash"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "object", Description: "Hash result in msg.payload.hash"},
		},
		Factory: NewHashExecutor,
	})
}
