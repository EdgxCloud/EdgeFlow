package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"golang.org/x/crypto/ssh"
)

// SFTPNode handles SFTP operations over SSH
type SFTPNode struct {
	host       string
	port       int
	username   string
	password   string
	privateKey string
	timeout    time.Duration
	client     *ssh.Client
}

// NewSFTPNode creates a new SFTP node
func NewSFTPNode() *SFTPNode {
	return &SFTPNode{
		port:    22,
		timeout: 30 * time.Second,
	}
}

// Init initializes the SFTP node
func (n *SFTPNode) Init(config map[string]interface{}) error {
	if host, ok := config["host"].(string); ok {
		n.host = host
	} else {
		return fmt.Errorf("SFTP host is required")
	}
	if port, ok := config["port"].(float64); ok {
		n.port = int(port)
	}
	if username, ok := config["username"].(string); ok {
		n.username = username
	} else {
		return fmt.Errorf("SFTP username is required")
	}
	if password, ok := config["password"].(string); ok {
		n.password = password
	}
	if privateKey, ok := config["privateKey"].(string); ok {
		n.privateKey = privateKey
	}
	if timeout, ok := config["timeout"].(float64); ok {
		n.timeout = time.Duration(timeout) * time.Millisecond
	}
	return nil
}

func (n *SFTPNode) connect() error {
	if n.client != nil {
		return nil
	}

	config := &ssh.ClientConfig{
		User:            n.username,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         n.timeout,
	}

	if n.privateKey != "" {
		signer, err := ssh.ParsePrivateKey([]byte(n.privateKey))
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
		config.Auth = []ssh.AuthMethod{ssh.PublicKeys(signer)}
	} else if n.password != "" {
		config.Auth = []ssh.AuthMethod{ssh.Password(n.password)}
	} else {
		return fmt.Errorf("either password or privateKey is required")
	}

	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", n.host, n.port), config)
	if err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	n.client = client
	return nil
}

func (n *SFTPNode) runCommand(cmd string) (string, error) {
	session, err := n.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", fmt.Errorf("command failed: %s: %w", stderr.String(), err)
	}
	return stdout.String(), nil
}

// Execute performs SFTP operations
func (n *SFTPNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if err := n.connect(); err != nil {
		return node.Message{}, err
	}

	operation := "list"
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	var result interface{}
	var err error

	switch operation {
	case "list":
		path := "."
		if p, ok := msg.Payload["path"].(string); ok {
			path = p
		}
		output, e := n.runCommand(fmt.Sprintf("ls -la %s", path))
		if e != nil {
			err = e
		} else {
			lines := strings.Split(strings.TrimSpace(output), "\n")
			var files []map[string]interface{}
			for _, line := range lines {
				if line == "" || strings.HasPrefix(line, "total") {
					continue
				}
				files = append(files, map[string]interface{}{"raw": line})
			}
			result = files
		}
	case "download":
		remotePath, _ := msg.Payload["remotePath"].(string)
		if remotePath == "" {
			return node.Message{}, fmt.Errorf("remotePath is required")
		}
		output, e := n.runCommand(fmt.Sprintf("cat %s", remotePath))
		if e != nil {
			err = e
		} else {
			result = map[string]interface{}{
				"remotePath": remotePath,
				"content":    output,
				"size":       len(output),
			}
		}
	case "upload":
		remotePath, _ := msg.Payload["remotePath"].(string)
		content, _ := msg.Payload["content"].(string)
		if remotePath == "" || content == "" {
			return node.Message{}, fmt.Errorf("remotePath and content are required")
		}
		session, e := n.client.NewSession()
		if e != nil {
			err = e
		} else {
			defer session.Close()
			session.Stdin = strings.NewReader(content)
			e = session.Run(fmt.Sprintf("cat > %s", remotePath))
			if e != nil {
				err = e
			} else {
				result = map[string]interface{}{
					"remotePath": remotePath,
					"size":       len(content),
					"uploaded":   true,
				}
			}
		}
	case "delete":
		remotePath, _ := msg.Payload["remotePath"].(string)
		if remotePath == "" {
			return node.Message{}, fmt.Errorf("remotePath is required")
		}
		_, e := n.runCommand(fmt.Sprintf("rm -f %s", remotePath))
		if e != nil {
			err = e
		} else {
			result = map[string]interface{}{"remotePath": remotePath, "deleted": true}
		}
	case "mkdir":
		path, _ := msg.Payload["path"].(string)
		if path == "" {
			return node.Message{}, fmt.Errorf("path is required")
		}
		_, e := n.runCommand(fmt.Sprintf("mkdir -p %s", path))
		if e != nil {
			err = e
		} else {
			result = map[string]interface{}{"path": path, "created": true}
		}
	case "rename":
		oldPath, _ := msg.Payload["oldPath"].(string)
		newPath, _ := msg.Payload["newPath"].(string)
		if oldPath == "" || newPath == "" {
			return node.Message{}, fmt.Errorf("oldPath and newPath are required")
		}
		_, e := n.runCommand(fmt.Sprintf("mv %s %s", oldPath, newPath))
		if e != nil {
			err = e
		} else {
			result = map[string]interface{}{"oldPath": oldPath, "newPath": newPath, "renamed": true}
		}
	case "stat":
		remotePath, _ := msg.Payload["remotePath"].(string)
		if remotePath == "" {
			return node.Message{}, fmt.Errorf("remotePath is required")
		}
		output, e := n.runCommand(fmt.Sprintf("stat %s", remotePath))
		if e != nil {
			err = e
		} else {
			result = map[string]interface{}{"remotePath": remotePath, "stat": output}
		}
	default:
		return node.Message{}, fmt.Errorf("unsupported operation: %s", operation)
	}

	if err != nil {
		return node.Message{}, err
	}

	outputPayload := make(map[string]interface{})
	for k, v := range msg.Payload {
		outputPayload[k] = v
	}
	outputPayload["result"] = result
	outputPayload["operation"] = operation

	return node.Message{
		Type:    node.MessageTypeData,
		Payload: outputPayload,
		Topic:   msg.Topic,
	}, nil
}

// Cleanup closes the SSH connection
func (n *SFTPNode) Cleanup() error {
	if n.client != nil {
		n.client.Close()
		n.client = nil
	}
	return nil
}
