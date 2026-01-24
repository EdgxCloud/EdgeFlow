package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/jlaffaye/ftp"
)

// FTPNode handles FTP operations
type FTPNode struct {
	host     string
	port     int
	username string
	password string
	conn     *ftp.ServerConn
}

// NewFTPNode creates a new FTP node
func NewFTPNode() *FTPNode {
	return &FTPNode{
		port: 21,
	}
}

// Init initializes the FTP node
func (n *FTPNode) Init(config map[string]interface{}) error {
	// Parse host
	if host, ok := config["host"].(string); ok {
		n.host = host
	} else {
		return fmt.Errorf("FTP host is required")
	}

	// Parse port
	if port, ok := config["port"].(float64); ok {
		n.port = int(port)
	}

	// Parse username
	if username, ok := config["username"].(string); ok {
		n.username = username
	} else {
		n.username = "anonymous"
	}

	// Parse password
	if password, ok := config["password"].(string); ok {
		n.password = password
	}

	// Connect to FTP server
	conn, err := ftp.Dial(fmt.Sprintf("%s:%d", n.host, n.port), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return fmt.Errorf("failed to connect to FTP server: %w", err)
	}

	// Login
	err = conn.Login(n.username, n.password)
	if err != nil {
		conn.Quit()
		return fmt.Errorf("failed to login to FTP server: %w", err)
	}

	n.conn = conn

	return nil
}

// Execute processes incoming messages
func (n *FTPNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if n.conn == nil {
		return node.Message{}, fmt.Errorf("FTP connection not established")
	}

	// Get operation type
	operation := "list"
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	var result interface{}
	var err error

	switch operation {
	case "upload":
		result, err = n.uploadFile(msg)
	case "download":
		result, err = n.downloadFile(msg)
	case "list":
		result, err = n.listFiles(msg)
	case "delete":
		result, err = n.deleteFile(msg)
	case "mkdir":
		result, err = n.makeDirectory(msg)
	case "rename":
		result, err = n.renameFile(msg)
	default:
		return node.Message{}, fmt.Errorf("unsupported operation: %s", operation)
	}

	if err != nil {
		return node.Message{}, err
	}

	// Create output message
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

// uploadFile uploads a file to FTP server
func (n *FTPNode) uploadFile(msg node.Message) (map[string]interface{}, error) {
	// Get local file path
	localPath, ok := msg.Payload["localPath"].(string)
	if !ok {
		return nil, fmt.Errorf("localPath is required for upload")
	}

	// Get remote path
	remotePath, ok := msg.Payload["remotePath"].(string)
	if !ok {
		return nil, fmt.Errorf("remotePath is required for upload")
	}

	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Upload file
	err = n.conn.Stor(remotePath, file)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return map[string]interface{}{
		"localPath":  localPath,
		"remotePath": remotePath,
		"size":       fileInfo.Size(),
		"uploaded":   true,
	}, nil
}

// downloadFile downloads a file from FTP server
func (n *FTPNode) downloadFile(msg node.Message) (map[string]interface{}, error) {
	// Get remote path
	remotePath, ok := msg.Payload["remotePath"].(string)
	if !ok {
		return nil, fmt.Errorf("remotePath is required for download")
	}

	// Get local path
	localPath, ok := msg.Payload["localPath"].(string)
	if !ok {
		return nil, fmt.Errorf("localPath is required for download")
	}

	// Download file
	response, err := n.conn.Retr(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve file: %w", err)
	}
	defer response.Close()

	// Create local file
	file, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file: %w", err)
	}
	defer file.Close()

	// Copy content
	size, err := io.Copy(file, response)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"remotePath": remotePath,
		"localPath":  localPath,
		"size":       size,
		"downloaded": true,
	}, nil
}

// listFiles lists files on FTP server
func (n *FTPNode) listFiles(msg node.Message) ([]map[string]interface{}, error) {
	// Get path (default to current directory)
	path := "."
	if p, ok := msg.Payload["path"].(string); ok {
		path = p
	}

	// List directory
	entries, err := n.conn.List(path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	// Convert to result
	files := make([]map[string]interface{}, 0, len(entries))
	for _, entry := range entries {
		files = append(files, map[string]interface{}{
			"name": entry.Name,
			"size": entry.Size,
			"type": entry.Type.String(),
			"time": entry.Time.Format(time.RFC3339),
		})
	}

	return files, nil
}

// deleteFile deletes a file from FTP server
func (n *FTPNode) deleteFile(msg node.Message) (map[string]interface{}, error) {
	// Get remote path
	remotePath, ok := msg.Payload["remotePath"].(string)
	if !ok {
		return nil, fmt.Errorf("remotePath is required for delete")
	}

	// Delete file
	err := n.conn.Delete(remotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return map[string]interface{}{
		"remotePath": remotePath,
		"deleted":    true,
	}, nil
}

// makeDirectory creates a directory on FTP server
func (n *FTPNode) makeDirectory(msg node.Message) (map[string]interface{}, error) {
	// Get path
	path, ok := msg.Payload["path"].(string)
	if !ok {
		return nil, fmt.Errorf("path is required for mkdir")
	}

	// Create directory
	err := n.conn.MakeDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return map[string]interface{}{
		"path":    path,
		"created": true,
	}, nil
}

// renameFile renames a file on FTP server
func (n *FTPNode) renameFile(msg node.Message) (map[string]interface{}, error) {
	// Get old path
	oldPath, ok := msg.Payload["oldPath"].(string)
	if !ok {
		return nil, fmt.Errorf("oldPath is required for rename")
	}

	// Get new path
	newPath, ok := msg.Payload["newPath"].(string)
	if !ok {
		return nil, fmt.Errorf("newPath is required for rename")
	}

	// Rename file
	err := n.conn.Rename(oldPath, newPath)
	if err != nil {
		return nil, fmt.Errorf("failed to rename file: %w", err)
	}

	return map[string]interface{}{
		"oldPath": oldPath,
		"newPath": newPath,
		"renamed": true,
	}, nil
}

// Cleanup closes the FTP connection
func (n *FTPNode) Cleanup() error {
	if n.conn != nil {
		return n.conn.Quit()
	}
	return nil
}
