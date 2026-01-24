package core

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"

	"github.com/edgeflow/edgeflow/internal/node"
)

// EncryptNode encrypts or decrypts data
type EncryptNode struct {
	operation string // encrypt, decrypt
	algorithm string // aes-128-gcm, aes-256-gcm, aes-128-cbc, aes-256-cbc
	key       []byte // Encryption key
	encoding  string // hex, base64, raw
	property  string // Property to encrypt/decrypt
}

// Init initializes the encrypt node
func (n *EncryptNode) Init(config map[string]interface{}) error {
	// Operation
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	} else {
		n.operation = "encrypt"
	}

	// Algorithm
	if algo, ok := config["algorithm"].(string); ok {
		n.algorithm = algo
	} else {
		n.algorithm = "aes-256-gcm"
	}

	// Key (hex or base64 encoded)
	if keyStr, ok := config["key"].(string); ok {
		// Try hex first
		key, err := hex.DecodeString(keyStr)
		if err != nil {
			// Try base64
			key, err = base64.StdEncoding.DecodeString(keyStr)
			if err != nil {
				// Use raw string as key (will be padded/truncated)
				key = []byte(keyStr)
			}
		}
		n.key = key
	} else {
		return fmt.Errorf("encryption key is required")
	}

	// Validate key length
	keyLen := n.getKeyLength()
	if len(n.key) < keyLen {
		// Pad key
		padded := make([]byte, keyLen)
		copy(padded, n.key)
		n.key = padded
	} else if len(n.key) > keyLen {
		n.key = n.key[:keyLen]
	}

	// Output encoding
	if enc, ok := config["encoding"].(string); ok {
		n.encoding = enc
	} else {
		n.encoding = "base64"
	}

	// Property
	if prop, ok := config["property"].(string); ok {
		n.property = prop
	} else {
		n.property = "value"
	}

	return nil
}

// getKeyLength returns required key length for algorithm
func (n *EncryptNode) getKeyLength() int {
	switch n.algorithm {
	case "aes-128-gcm", "aes-128-cbc":
		return 16
	case "aes-256-gcm", "aes-256-cbc":
		return 32
	default:
		return 32
	}
}

// Execute performs encryption or decryption
func (n *EncryptNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	// Get input data
	var input interface{}
	if val, ok := msg.Payload[n.property]; ok {
		input = val
	} else if val, ok := msg.Payload["value"]; ok {
		input = val
	} else {
		return msg, fmt.Errorf("property %s not found in payload", n.property)
	}

	switch n.operation {
	case "encrypt":
		return n.encrypt(msg, input)
	case "decrypt":
		return n.decrypt(msg, input)
	default:
		return msg, fmt.Errorf("unknown operation: %s", n.operation)
	}
}

func (n *EncryptNode) encrypt(msg node.Message, input interface{}) (node.Message, error) {
	// Convert input to bytes
	var plaintext []byte
	switch v := input.(type) {
	case []byte:
		plaintext = v
	case string:
		plaintext = []byte(v)
	default:
		plaintext = []byte(fmt.Sprintf("%v", v))
	}

	// Create cipher block
	block, err := aes.NewCipher(n.key)
	if err != nil {
		return msg, fmt.Errorf("failed to create cipher: %w", err)
	}

	var ciphertext []byte

	switch n.algorithm {
	case "aes-128-gcm", "aes-256-gcm":
		// GCM mode
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return msg, fmt.Errorf("failed to create GCM: %w", err)
		}

		// Generate nonce
		nonce := make([]byte, gcm.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return msg, fmt.Errorf("failed to generate nonce: %w", err)
		}

		// Encrypt and prepend nonce
		ciphertext = gcm.Seal(nonce, nonce, plaintext, nil)

	case "aes-128-cbc", "aes-256-cbc":
		// CBC mode with PKCS7 padding
		plaintext = pkcs7Pad(plaintext, aes.BlockSize)

		ciphertext = make([]byte, aes.BlockSize+len(plaintext))
		iv := ciphertext[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			return msg, fmt.Errorf("failed to generate IV: %w", err)
		}

		mode := cipher.NewCBCEncrypter(block, iv)
		mode.CryptBlocks(ciphertext[aes.BlockSize:], plaintext)

	default:
		return msg, fmt.Errorf("unsupported algorithm: %s", n.algorithm)
	}

	// Encode output
	var encoded string
	switch n.encoding {
	case "hex":
		encoded = hex.EncodeToString(ciphertext)
	case "base64":
		encoded = base64.StdEncoding.EncodeToString(ciphertext)
	case "base64url":
		encoded = base64.URLEncoding.EncodeToString(ciphertext)
	default:
		encoded = base64.StdEncoding.EncodeToString(ciphertext)
	}

	msg.Payload["value"] = encoded
	msg.Payload["_encrypt"] = map[string]interface{}{
		"operation": "encrypt",
		"algorithm": n.algorithm,
		"encoding":  n.encoding,
		"inputLen":  len(plaintext),
		"outputLen": len(ciphertext),
	}

	return msg, nil
}

func (n *EncryptNode) decrypt(msg node.Message, input interface{}) (node.Message, error) {
	// Get ciphertext
	var encoded string
	switch v := input.(type) {
	case string:
		encoded = v
	default:
		return msg, fmt.Errorf("decrypt requires string input")
	}

	// Decode input
	var ciphertext []byte
	var err error
	switch n.encoding {
	case "hex":
		ciphertext, err = hex.DecodeString(encoded)
	case "base64":
		ciphertext, err = base64.StdEncoding.DecodeString(encoded)
	case "base64url":
		ciphertext, err = base64.URLEncoding.DecodeString(encoded)
	default:
		ciphertext, err = base64.StdEncoding.DecodeString(encoded)
	}
	if err != nil {
		return msg, fmt.Errorf("failed to decode ciphertext: %w", err)
	}

	// Create cipher block
	block, err := aes.NewCipher(n.key)
	if err != nil {
		return msg, fmt.Errorf("failed to create cipher: %w", err)
	}

	var plaintext []byte

	switch n.algorithm {
	case "aes-128-gcm", "aes-256-gcm":
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return msg, fmt.Errorf("failed to create GCM: %w", err)
		}

		nonceSize := gcm.NonceSize()
		if len(ciphertext) < nonceSize {
			return msg, fmt.Errorf("ciphertext too short")
		}

		nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
		plaintext, err = gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return msg, fmt.Errorf("decryption failed: %w", err)
		}

	case "aes-128-cbc", "aes-256-cbc":
		if len(ciphertext) < aes.BlockSize {
			return msg, fmt.Errorf("ciphertext too short")
		}

		iv := ciphertext[:aes.BlockSize]
		ciphertext = ciphertext[aes.BlockSize:]

		if len(ciphertext)%aes.BlockSize != 0 {
			return msg, fmt.Errorf("ciphertext not a multiple of block size")
		}

		mode := cipher.NewCBCDecrypter(block, iv)
		plaintext = make([]byte, len(ciphertext))
		mode.CryptBlocks(plaintext, ciphertext)

		// Remove PKCS7 padding
		plaintext, err = pkcs7Unpad(plaintext)
		if err != nil {
			return msg, fmt.Errorf("padding error: %w", err)
		}

	default:
		return msg, fmt.Errorf("unsupported algorithm: %s", n.algorithm)
	}

	// Return as string if valid UTF-8
	result := string(plaintext)
	msg.Payload["value"] = result
	msg.Payload["_encrypt"] = map[string]interface{}{
		"operation": "decrypt",
		"algorithm": n.algorithm,
		"encoding":  n.encoding,
		"inputLen":  len(ciphertext),
		"outputLen": len(plaintext),
	}

	return msg, nil
}

// PKCS7 padding functions

func pkcs7Pad(data []byte, blockSize int) []byte {
	padding := blockSize - len(data)%blockSize
	padtext := make([]byte, padding)
	for i := range padtext {
		padtext[i] = byte(padding)
	}
	return append(data, padtext...)
}

func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padding := int(data[len(data)-1])
	if padding > len(data) || padding == 0 {
		return nil, fmt.Errorf("invalid padding")
	}
	for i := len(data) - padding; i < len(data); i++ {
		if data[i] != byte(padding) {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	return data[:len(data)-padding], nil
}

// Cleanup releases resources
func (n *EncryptNode) Cleanup() error {
	return nil
}

// NewEncryptExecutor creates a new encrypt node executor
func NewEncryptExecutor() node.Executor {
	return &EncryptNode{}
}

// init registers the encrypt node
func init() {
	registry := node.GetGlobalRegistry()
	registry.Register(&node.NodeInfo{
		Type:        "encrypt",
		Name:        "Encrypt/Decrypt",
		Category:    node.NodeTypeFunction,
		Description: "Encrypt or decrypt data using AES (GCM or CBC mode)",
		Icon:        "lock",
		Color:       "#DC143C",
		Properties: []node.PropertySchema{
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "encrypt",
				Required:    true,
				Description: "Encrypt or decrypt",
				Options:     []string{"encrypt", "decrypt"},
			},
			{
				Name:        "algorithm",
				Label:       "Algorithm",
				Type:        "select",
				Default:     "aes-256-gcm",
				Required:    true,
				Description: "Encryption algorithm and mode",
				Options:     []string{"aes-128-gcm", "aes-256-gcm", "aes-128-cbc", "aes-256-cbc"},
			},
			{
				Name:        "key",
				Label:       "Key",
				Type:        "password",
				Default:     "",
				Required:    true,
				Description: "Encryption key (hex or base64 encoded, or raw string)",
			},
			{
				Name:        "encoding",
				Label:       "Output Encoding",
				Type:        "select",
				Default:     "base64",
				Required:    true,
				Description: "Output encoding for ciphertext",
				Options:     []string{"base64", "base64url", "hex"},
			},
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "value",
				Required:    false,
				Description: "Property to encrypt/decrypt",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to encrypt/decrypt"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "string", Description: "Encrypted/decrypted data"},
		},
		Factory: NewEncryptExecutor,
	})
}
