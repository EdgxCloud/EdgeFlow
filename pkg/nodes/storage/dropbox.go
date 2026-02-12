package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
)

// DropboxNode handles Dropbox cloud storage operations via REST API
type DropboxNode struct {
	accessToken string
	rootPath    string
	httpClient  *http.Client
}

// NewDropboxNode creates a new Dropbox node
func NewDropboxNode() *DropboxNode {
	return &DropboxNode{
		rootPath: "",
	}
}

// Init initializes the Dropbox node
func (n *DropboxNode) Init(config map[string]interface{}) error {
	if token, ok := config["accessToken"].(string); ok && token != "" {
		n.accessToken = token
	} else {
		return fmt.Errorf("Dropbox access token is required")
	}
	if rootPath, ok := config["rootPath"].(string); ok {
		n.rootPath = rootPath
	}
	n.httpClient = &http.Client{Timeout: 60 * time.Second}
	return nil
}

func (n *DropboxNode) doAPI(endpoint string, body interface{}) (map[string]interface{}, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.dropboxapi.com/2"+endpoint, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+n.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Dropbox API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

// Execute performs Dropbox operations
func (n *DropboxNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	operation := "list"
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	var result interface{}
	var err error

	switch operation {
	case "list":
		path := n.rootPath
		if p, ok := msg.Payload["path"].(string); ok {
			path = p
		}
		if path == "" {
			path = ""
		}
		result, err = n.doAPI("/files/list_folder", map[string]interface{}{
			"path":                            path,
			"recursive":                       false,
			"include_media_info":              false,
			"include_deleted":                 false,
			"include_has_explicit_shared_members": false,
		})

	case "upload":
		localPath, _ := msg.Payload["localPath"].(string)
		remotePath, _ := msg.Payload["remotePath"].(string)
		if localPath == "" || remotePath == "" {
			return node.Message{}, fmt.Errorf("localPath and remotePath are required")
		}
		fileData, e := os.ReadFile(localPath)
		if e != nil {
			return node.Message{}, fmt.Errorf("failed to read file: %w", e)
		}
		apiArg, _ := json.Marshal(map[string]interface{}{
			"path":       remotePath,
			"mode":       "overwrite",
			"autorename": false,
			"mute":       false,
		})
		req, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/upload", bytes.NewReader(fileData))
		req.Header.Set("Authorization", "Bearer "+n.accessToken)
		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Dropbox-API-Arg", string(apiArg))
		resp, e := n.httpClient.Do(req)
		if e != nil {
			err = e
		} else {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode == 200 {
				var r map[string]interface{}
				json.Unmarshal(body, &r)
				result = r
			} else {
				err = fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(body))
			}
		}

	case "download":
		remotePath, _ := msg.Payload["remotePath"].(string)
		localPath, _ := msg.Payload["localPath"].(string)
		if remotePath == "" {
			return node.Message{}, fmt.Errorf("remotePath is required")
		}
		apiArg, _ := json.Marshal(map[string]interface{}{"path": remotePath})
		req, _ := http.NewRequest("POST", "https://content.dropboxapi.com/2/files/download", nil)
		req.Header.Set("Authorization", "Bearer "+n.accessToken)
		req.Header.Set("Dropbox-API-Arg", string(apiArg))
		resp, e := n.httpClient.Do(req)
		if e != nil {
			err = e
		} else {
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			if resp.StatusCode == 200 {
				if localPath != "" {
					os.WriteFile(localPath, data, 0644)
				}
				result = map[string]interface{}{
					"remotePath": remotePath,
					"size":       len(data),
					"downloaded": true,
				}
			} else {
				err = fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(data))
			}
		}

	case "delete":
		path, _ := msg.Payload["path"].(string)
		if path == "" {
			return node.Message{}, fmt.Errorf("path is required")
		}
		result, err = n.doAPI("/files/delete_v2", map[string]interface{}{"path": path})

	case "create_folder":
		path, _ := msg.Payload["path"].(string)
		if path == "" {
			return node.Message{}, fmt.Errorf("path is required")
		}
		result, err = n.doAPI("/files/create_folder_v2", map[string]interface{}{
			"path":       path,
			"autorename": false,
		})

	case "get_metadata":
		path, _ := msg.Payload["path"].(string)
		if path == "" {
			return node.Message{}, fmt.Errorf("path is required")
		}
		result, err = n.doAPI("/files/get_metadata", map[string]interface{}{"path": path})

	case "share":
		path, _ := msg.Payload["path"].(string)
		if path == "" {
			return node.Message{}, fmt.Errorf("path is required")
		}
		result, err = n.doAPI("/sharing/create_shared_link_with_settings", map[string]interface{}{
			"path": path,
			"settings": map[string]interface{}{
				"requested_visibility": "public",
			},
		})

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

// Cleanup releases resources
func (n *DropboxNode) Cleanup() error {
	return nil
}
