package network

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

type FileOutNode struct {
	filename       string
	action         string // "write", "append"
	createDir      bool
	encoding       string
	addNewline     bool
	overwriteFile  bool
}

func NewFileOutNode() *FileOutNode {
	return &FileOutNode{
		action:        "write",
		createDir:     true,
		encoding:      "utf-8",
		addNewline:    false,
		overwriteFile: true,
	}
}

func (n *FileOutNode) Init(config map[string]interface{}) error {
	if filename, ok := config["filename"].(string); ok {
		n.filename = filename
	}
	if action, ok := config["action"].(string); ok {
		n.action = action
	}
	if createDir, ok := config["createDir"].(bool); ok {
		n.createDir = createDir
	}
	if encoding, ok := config["encoding"].(string); ok {
		n.encoding = encoding
	}
	if addNewline, ok := config["addNewline"].(bool); ok {
		n.addNewline = addNewline
	}
	if overwrite, ok := config["overwriteFile"].(bool); ok {
		n.overwriteFile = overwrite
	}
	return nil
}

func (n *FileOutNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	filename := n.filename
	payload := msg.Payload
	if fn, ok := payload["filename"].(string); ok && fn != "" {
		filename = fn
	}

	if filename == "" {
		return msg, fmt.Errorf("no filename specified")
	}

	if n.createDir {
		dir := filepath.Dir(filename)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return msg, fmt.Errorf("failed to create directory: %w", err)
		}
	}

	var data []byte
	if content, ok := payload["content"].(string); ok {
		data = []byte(content)
	} else if content, ok := payload["content"].([]byte); ok {
		data = content
	} else {
		return msg, fmt.Errorf("no content in payload")
	}

	if n.addNewline {
		data = append(data, '\n')
	}

	switch n.action {
	case "write":
		flag := os.O_CREATE | os.O_WRONLY
		if n.overwriteFile {
			flag |= os.O_TRUNC
		}
		file, err := os.OpenFile(filename, flag, 0644)
		if err != nil {
			return msg, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		if _, err := file.Write(data); err != nil {
			return msg, fmt.Errorf("failed to write file: %w", err)
		}

	case "append":
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			return msg, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		if _, err := file.Write(data); err != nil {
			return msg, fmt.Errorf("failed to append to file: %w", err)
		}

	default:
		return msg, fmt.Errorf("unknown action: %s", n.action)
	}

	msg.Topic = "file-written"
	return msg, nil
}

func (n *FileOutNode) Cleanup() error {
	return nil
}

func NewFileOutExecutor() node.Executor {
	return NewFileOutNode()
}
