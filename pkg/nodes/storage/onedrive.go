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

	"github.com/EdgxCloud/EdgeFlow/internal/node"
)

// OneDriveNode handles Microsoft OneDrive operations via Graph API
type OneDriveNode struct {
	accessToken string
	driveId     string
	httpClient  *http.Client
}

// NewOneDriveNode creates a new OneDrive node
func NewOneDriveNode() *OneDriveNode {
	return &OneDriveNode{
		driveId: "me",
	}
}

// Init initializes the OneDrive node
func (n *OneDriveNode) Init(config map[string]interface{}) error {
	if token, ok := config["accessToken"].(string); ok && token != "" {
		n.accessToken = token
	} else {
		return fmt.Errorf("OneDrive access token is required")
	}
	if driveId, ok := config["driveId"].(string); ok && driveId != "" {
		n.driveId = driveId
	}
	n.httpClient = &http.Client{Timeout: 60 * time.Second}
	return nil
}

func (n *OneDriveNode) doGraphAPI(method, path string, body io.Reader) (map[string]interface{}, error) {
	url := "https://graph.microsoft.com/v1.0/me/drive" + path
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+n.accessToken)
	if body != nil && method != "GET" {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := n.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Graph API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("Graph API error (%d): %s", resp.StatusCode, string(respBody))
	}

	if len(respBody) == 0 {
		return map[string]interface{}{"success": true}, nil
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}
	return result, nil
}

// Execute performs OneDrive operations
func (n *OneDriveNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	operation := "list"
	if op, ok := msg.Payload["operation"].(string); ok {
		operation = op
	}

	var result interface{}
	var err error

	switch operation {
	case "list":
		path := "/root/children"
		if folderId, ok := msg.Payload["folderId"].(string); ok && folderId != "" {
			path = "/items/" + folderId + "/children"
		}
		result, err = n.doGraphAPI("GET", path, nil)

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
		url := "https://graph.microsoft.com/v1.0/me/drive/root:/" + remotePath + ":/content"
		req, _ := http.NewRequest("PUT", url, bytes.NewReader(fileData))
		req.Header.Set("Authorization", "Bearer "+n.accessToken)
		req.Header.Set("Content-Type", "application/octet-stream")
		resp, e := n.httpClient.Do(req)
		if e != nil {
			err = e
		} else {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				var r map[string]interface{}
				json.Unmarshal(body, &r)
				result = r
			} else {
				err = fmt.Errorf("upload failed (%d): %s", resp.StatusCode, string(body))
			}
		}

	case "download":
		itemId, _ := msg.Payload["itemId"].(string)
		localPath, _ := msg.Payload["localPath"].(string)
		if itemId == "" {
			return node.Message{}, fmt.Errorf("itemId is required")
		}
		url := "https://graph.microsoft.com/v1.0/me/drive/items/" + itemId + "/content"
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", "Bearer "+n.accessToken)
		resp, e := n.httpClient.Do(req)
		if e != nil {
			err = e
		} else {
			defer resp.Body.Close()
			data, _ := io.ReadAll(resp.Body)
			if resp.StatusCode >= 200 && resp.StatusCode < 300 {
				if localPath != "" {
					os.WriteFile(localPath, data, 0644)
				}
				result = map[string]interface{}{
					"itemId":     itemId,
					"size":       len(data),
					"downloaded": true,
				}
			} else {
				err = fmt.Errorf("download failed (%d): %s", resp.StatusCode, string(data))
			}
		}

	case "delete":
		itemId, _ := msg.Payload["itemId"].(string)
		if itemId == "" {
			return node.Message{}, fmt.Errorf("itemId is required")
		}
		result, err = n.doGraphAPI("DELETE", "/items/"+itemId, nil)

	case "create_folder":
		parentId := "root"
		if pid, ok := msg.Payload["parentId"].(string); ok && pid != "" {
			parentId = pid
		}
		folderName, _ := msg.Payload["folderName"].(string)
		if folderName == "" {
			return node.Message{}, fmt.Errorf("folderName is required")
		}
		body, _ := json.Marshal(map[string]interface{}{
			"name":                              folderName,
			"folder":                            map[string]interface{}{},
			"@microsoft.graph.conflictBehavior": "rename",
		})
		result, err = n.doGraphAPI("POST", "/items/"+parentId+"/children", bytes.NewReader(body))

	case "get_metadata":
		itemId, _ := msg.Payload["itemId"].(string)
		if itemId == "" {
			return node.Message{}, fmt.Errorf("itemId is required")
		}
		result, err = n.doGraphAPI("GET", "/items/"+itemId, nil)

	case "share":
		itemId, _ := msg.Payload["itemId"].(string)
		if itemId == "" {
			return node.Message{}, fmt.Errorf("itemId is required")
		}
		shareType := "view"
		if st, ok := msg.Payload["shareType"].(string); ok {
			shareType = st
		}
		body, _ := json.Marshal(map[string]interface{}{
			"type":  shareType,
			"scope": "anonymous",
		})
		result, err = n.doGraphAPI("POST", "/items/"+itemId+"/createLink", bytes.NewReader(body))

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
func (n *OneDriveNode) Cleanup() error {
	return nil
}
