package network

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/edgeflow/edgeflow/internal/node"
)

type FileInNode struct {
	filename string
	format   string // "utf8", "binary", "lines"
	encoding string
	watch    bool
}

func NewFileInNode() *FileInNode {
	return &FileInNode{
		format:   "utf8",
		encoding: "utf-8",
		watch:    false,
	}
}

func (n *FileInNode) Init(config map[string]interface{}) error {
	if filename, ok := config["filename"].(string); ok {
		n.filename = filename
	}
	if format, ok := config["format"].(string); ok {
		n.format = format
	}
	if encoding, ok := config["encoding"].(string); ok {
		n.encoding = encoding
	}
	if watch, ok := config["watch"].(bool); ok {
		n.watch = watch
	}
	return nil
}

func (n *FileInNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	filename := n.filename

	// Check if filename is provided in payload
	payload := msg.Payload
	if fn, ok := payload["filename"].(string); ok && fn != "" {
		filename = fn
	}

	if filename == "" {
		return msg, fmt.Errorf("no filename specified")
	}

	switch n.format {
	case "utf8":
		data, err := os.ReadFile(filename)
		if err != nil {
			return msg, fmt.Errorf("failed to read file: %w", err)
		}
		msg.Payload = map[string]interface{}{"data": string(data)}

	case "binary":
		data, err := os.ReadFile(filename)
		if err != nil {
			return msg, fmt.Errorf("failed to read file: %w", err)
		}
		msg.Payload = map[string]interface{}{"data": data}

	case "lines":
		file, err := os.Open(filename)
		if err != nil {
			return msg, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		var lines []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return msg, fmt.Errorf("failed to read lines: %w", err)
		}
		msg.Payload = map[string]interface{}{"data": lines}

	default:
		return msg, fmt.Errorf("unknown format: %s", n.format)
	}

	msg.Topic = "file-read"
	return msg, nil
}

func (n *FileInNode) Cleanup() error {
	return nil
}

func NewFileInExecutor() node.Executor {
	return NewFileInNode()
}
