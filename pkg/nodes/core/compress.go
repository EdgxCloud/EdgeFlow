package core

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"context"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/edgeflow/edgeflow/internal/node"
)

// CompressNode compresses or decompresses data
type CompressNode struct {
	operation string // compress, decompress
	algorithm string // gzip, zlib, deflate
	level     int    // Compression level (1-9)
	encoding  string // base64, raw
	property  string // Property to compress/decompress
}

// Init initializes the compress node
func (n *CompressNode) Init(config map[string]interface{}) error {
	// Operation
	if op, ok := config["operation"].(string); ok {
		n.operation = op
	} else {
		n.operation = "compress"
	}

	// Algorithm
	if algo, ok := config["algorithm"].(string); ok {
		n.algorithm = algo
	} else {
		n.algorithm = "gzip"
	}

	// Compression level
	if level, ok := config["level"].(float64); ok {
		n.level = int(level)
	} else if level, ok := config["level"].(int); ok {
		n.level = level
	} else {
		n.level = gzip.DefaultCompression
	}

	// Validate level
	if n.level < -1 || n.level > 9 {
		n.level = gzip.DefaultCompression
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

// Execute performs compression or decompression
func (n *CompressNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
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
	case "compress":
		return n.compress(msg, input)
	case "decompress":
		return n.decompress(msg, input)
	default:
		return msg, fmt.Errorf("unknown operation: %s", n.operation)
	}
}

func (n *CompressNode) compress(msg node.Message, input interface{}) (node.Message, error) {
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

	originalSize := len(data)

	// Compress
	var buf bytes.Buffer
	var err error

	switch n.algorithm {
	case "gzip":
		writer, err := gzip.NewWriterLevel(&buf, n.level)
		if err != nil {
			return msg, fmt.Errorf("failed to create gzip writer: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			return msg, fmt.Errorf("failed to write gzip data: %w", err)
		}
		if err := writer.Close(); err != nil {
			return msg, fmt.Errorf("failed to close gzip writer: %w", err)
		}

	case "zlib":
		writer, err := zlib.NewWriterLevel(&buf, n.level)
		if err != nil {
			return msg, fmt.Errorf("failed to create zlib writer: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			return msg, fmt.Errorf("failed to write zlib data: %w", err)
		}
		if err := writer.Close(); err != nil {
			return msg, fmt.Errorf("failed to close zlib writer: %w", err)
		}

	case "deflate":
		writer, err := flate.NewWriter(&buf, n.level)
		if err != nil {
			return msg, fmt.Errorf("failed to create deflate writer: %w", err)
		}
		if _, err := writer.Write(data); err != nil {
			return msg, fmt.Errorf("failed to write deflate data: %w", err)
		}
		if err := writer.Close(); err != nil {
			return msg, fmt.Errorf("failed to close deflate writer: %w", err)
		}

	default:
		return msg, fmt.Errorf("unsupported algorithm: %s", n.algorithm)
	}

	if err != nil {
		return msg, err
	}

	compressed := buf.Bytes()
	compressedSize := len(compressed)

	// Encode output
	var result interface{}
	switch n.encoding {
	case "base64":
		result = base64.StdEncoding.EncodeToString(compressed)
	case "base64url":
		result = base64.URLEncoding.EncodeToString(compressed)
	case "raw":
		result = compressed
	default:
		result = base64.StdEncoding.EncodeToString(compressed)
	}

	msg.Payload["value"] = result
	msg.Payload["_compress"] = map[string]interface{}{
		"operation":      "compress",
		"algorithm":      n.algorithm,
		"level":          n.level,
		"encoding":       n.encoding,
		"originalSize":   originalSize,
		"compressedSize": compressedSize,
		"ratio":          float64(compressedSize) / float64(originalSize),
		"savings":        fmt.Sprintf("%.1f%%", (1-float64(compressedSize)/float64(originalSize))*100),
	}

	return msg, nil
}

func (n *CompressNode) decompress(msg node.Message, input interface{}) (node.Message, error) {
	// Get compressed data
	var data []byte

	switch v := input.(type) {
	case []byte:
		data = v
	case string:
		// Decode if encoded
		var err error
		switch n.encoding {
		case "base64":
			data, err = base64.StdEncoding.DecodeString(v)
		case "base64url":
			data, err = base64.URLEncoding.DecodeString(v)
		case "raw":
			data = []byte(v)
		default:
			data, err = base64.StdEncoding.DecodeString(v)
		}
		if err != nil {
			return msg, fmt.Errorf("failed to decode input: %w", err)
		}
	default:
		return msg, fmt.Errorf("decompress requires bytes or string input")
	}

	compressedSize := len(data)

	// Decompress
	var buf bytes.Buffer
	reader := bytes.NewReader(data)

	switch n.algorithm {
	case "gzip":
		gzReader, err := gzip.NewReader(reader)
		if err != nil {
			return msg, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()

		if _, err := io.Copy(&buf, gzReader); err != nil {
			return msg, fmt.Errorf("failed to decompress gzip: %w", err)
		}

	case "zlib":
		zlibReader, err := zlib.NewReader(reader)
		if err != nil {
			return msg, fmt.Errorf("failed to create zlib reader: %w", err)
		}
		defer zlibReader.Close()

		if _, err := io.Copy(&buf, zlibReader); err != nil {
			return msg, fmt.Errorf("failed to decompress zlib: %w", err)
		}

	case "deflate":
		flateReader := flate.NewReader(reader)
		defer flateReader.Close()

		if _, err := io.Copy(&buf, flateReader); err != nil {
			return msg, fmt.Errorf("failed to decompress deflate: %w", err)
		}

	default:
		return msg, fmt.Errorf("unsupported algorithm: %s", n.algorithm)
	}

	decompressed := buf.Bytes()
	decompressedSize := len(decompressed)

	// Return as string
	msg.Payload["value"] = string(decompressed)
	msg.Payload["_compress"] = map[string]interface{}{
		"operation":        "decompress",
		"algorithm":        n.algorithm,
		"encoding":         n.encoding,
		"compressedSize":   compressedSize,
		"decompressedSize": decompressedSize,
		"ratio":            float64(decompressedSize) / float64(compressedSize),
	}

	return msg, nil
}

// Cleanup releases resources
func (n *CompressNode) Cleanup() error {
	return nil
}

// NewCompressExecutor creates a new compress node executor
func NewCompressExecutor() node.Executor {
	return &CompressNode{}
}

// init registers the compress node
func init() {
	registry := node.GetGlobalRegistry()
	registry.Register(&node.NodeInfo{
		Type:        "compress",
		Name:        "Compress/Decompress",
		Category:    node.NodeTypeFunction,
		Description: "Compress or decompress data using gzip, zlib, or deflate",
		Icon:        "archive",
		Color:       "#8B4513",
		Properties: []node.PropertySchema{
			{
				Name:        "operation",
				Label:       "Operation",
				Type:        "select",
				Default:     "compress",
				Required:    true,
				Description: "Compress or decompress",
				Options:     []string{"compress", "decompress"},
			},
			{
				Name:        "algorithm",
				Label:       "Algorithm",
				Type:        "select",
				Default:     "gzip",
				Required:    true,
				Description: "Compression algorithm",
				Options:     []string{"gzip", "zlib", "deflate"},
			},
			{
				Name:        "level",
				Label:       "Compression Level",
				Type:        "number",
				Default:     -1,
				Required:    false,
				Description: "Compression level (1=fastest, 9=best, -1=default)",
			},
			{
				Name:        "encoding",
				Label:       "Encoding",
				Type:        "select",
				Default:     "base64",
				Required:    true,
				Description: "Output encoding for compressed data",
				Options:     []string{"base64", "base64url", "raw"},
			},
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "value",
				Required:    false,
				Description: "Property to compress/decompress",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Data to compress/decompress"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Compressed/decompressed data"},
		},
		Factory: NewCompressExecutor,
	})
}
