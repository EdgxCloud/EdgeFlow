package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/edgeflow/edgeflow/internal/node"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
)

// GoogleDriveNode handles Google Drive operations
type GoogleDriveNode struct {
	credentialsJSON string
	service         *drive.Service
	folderId        string
}

// NewGoogleDriveNode creates a new Google Drive node
func NewGoogleDriveNode() *GoogleDriveNode {
	return &GoogleDriveNode{}
}

// Init initializes the Google Drive node
func (n *GoogleDriveNode) Init(config map[string]interface{}) error {
	// Parse credentials JSON
	if creds, ok := config["credentials"].(string); ok {
		n.credentialsJSON = creds
	} else {
		return fmt.Errorf("Google Drive credentials are required")
	}

	// Parse optional folder ID
	if folderId, ok := config["folderId"].(string); ok {
		n.folderId = folderId
	}

	// Create OAuth2 config
	oauthConfig, err := google.ConfigFromJSON([]byte(n.credentialsJSON), drive.DriveFileScope)
	if err != nil {
		return fmt.Errorf("failed to parse credentials: %w", err)
	}

	// Get token (in production, this should be stored securely)
	token := &oauth2.Token{}
	if tokenStr, ok := config["token"].(string); ok {
		if err := json.Unmarshal([]byte(tokenStr), token); err != nil {
			return fmt.Errorf("failed to parse token: %w", err)
		}
	} else {
		return fmt.Errorf("OAuth2 token is required")
	}

	// Create HTTP client
	client := oauthConfig.Client(context.Background(), token)

	// Create Drive service
	service, err := drive.NewService(context.Background(), option.WithHTTPClient(client))
	if err != nil {
		return fmt.Errorf("failed to create Drive service: %w", err)
	}

	n.service = service

	return nil
}

// Execute processes incoming messages
func (n *GoogleDriveNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if n.service == nil {
		return node.Message{}, fmt.Errorf("Google Drive service not initialized")
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
		result, err = n.uploadFile(ctx, msg)
	case "download":
		result, err = n.downloadFile(ctx, msg)
	case "list":
		result, err = n.listFiles(ctx, msg)
	case "delete":
		result, err = n.deleteFile(ctx, msg)
	case "share":
		result, err = n.shareFile(ctx, msg)
	case "createFolder":
		result, err = n.createFolder(ctx, msg)
	case "getMetadata":
		result, err = n.getMetadata(ctx, msg)
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

// uploadFile uploads a file to Google Drive
func (n *GoogleDriveNode) uploadFile(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get file path
	filePath, ok := msg.Payload["filePath"].(string)
	if !ok {
		return nil, fmt.Errorf("filePath is required for upload")
	}

	// Get file name (optional, defaults to base name)
	fileName := ""
	if name, ok := msg.Payload["fileName"].(string); ok {
		fileName = name
	}

	// Get parent folder ID (optional)
	parentId := n.folderId
	if parent, ok := msg.Payload["parentId"].(string); ok {
		parentId = parent
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	if fileName == "" {
		fileName = fileInfo.Name()
	}

	// Create file metadata
	driveFile := &drive.File{
		Name: fileName,
	}

	if parentId != "" {
		driveFile.Parents = []string{parentId}
	}

	// Upload file
	uploadedFile, err := n.service.Files.Create(driveFile).
		Media(file).
		Context(ctx).
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return map[string]interface{}{
		"id":       uploadedFile.Id,
		"name":     uploadedFile.Name,
		"mimeType": uploadedFile.MimeType,
		"size":     uploadedFile.Size,
		"webViewLink": uploadedFile.WebViewLink,
		"uploaded": true,
	}, nil
}

// downloadFile downloads a file from Google Drive
func (n *GoogleDriveNode) downloadFile(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get file ID
	fileId, ok := msg.Payload["fileId"].(string)
	if !ok {
		return nil, fmt.Errorf("fileId is required for download")
	}

	// Get destination path
	destPath, ok := msg.Payload["destPath"].(string)
	if !ok {
		return nil, fmt.Errorf("destPath is required for download")
	}

	// Download file
	response, err := n.service.Files.Get(fileId).
		Context(ctx).
		Download()

	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer response.Body.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy content
	size, err := io.Copy(destFile, response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"fileId":     fileId,
		"destPath":   destPath,
		"size":       size,
		"downloaded": true,
	}, nil
}

// listFiles lists files in Google Drive
func (n *GoogleDriveNode) listFiles(ctx context.Context, msg node.Message) ([]map[string]interface{}, error) {
	// Build query
	query := ""
	if q, ok := msg.Payload["query"].(string); ok {
		query = q
	} else if n.folderId != "" {
		query = fmt.Sprintf("'%s' in parents", n.folderId)
	}

	// Get page size (default 100)
	pageSize := int64(100)
	if ps, ok := msg.Payload["pageSize"].(float64); ok {
		pageSize = int64(ps)
	}

	// List files
	fileList, err := n.service.Files.List().
		Q(query).
		PageSize(pageSize).
		Fields("files(id, name, mimeType, size, modifiedTime, webViewLink)").
		Context(ctx).
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Convert to result
	files := make([]map[string]interface{}, 0, len(fileList.Files))
	for _, file := range fileList.Files {
		files = append(files, map[string]interface{}{
			"id":           file.Id,
			"name":         file.Name,
			"mimeType":     file.MimeType,
			"size":         file.Size,
			"modifiedTime": file.ModifiedTime,
			"webViewLink":  file.WebViewLink,
		})
	}

	return files, nil
}

// deleteFile deletes a file from Google Drive
func (n *GoogleDriveNode) deleteFile(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get file ID
	fileId, ok := msg.Payload["fileId"].(string)
	if !ok {
		return nil, fmt.Errorf("fileId is required for delete")
	}

	// Delete file
	err := n.service.Files.Delete(fileId).
		Context(ctx).
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to delete file: %w", err)
	}

	return map[string]interface{}{
		"fileId":  fileId,
		"deleted": true,
	}, nil
}

// shareFile creates a shareable link for a file
func (n *GoogleDriveNode) shareFile(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get file ID
	fileId, ok := msg.Payload["fileId"].(string)
	if !ok {
		return nil, fmt.Errorf("fileId is required for share")
	}

	// Get permission type (default: anyone with link can view)
	permType := "anyone"
	if pt, ok := msg.Payload["type"].(string); ok {
		permType = pt
	}

	role := "reader"
	if r, ok := msg.Payload["role"].(string); ok {
		role = r
	}

	// Create permission
	permission := &drive.Permission{
		Type: permType,
		Role: role,
	}

	_, err := n.service.Permissions.Create(fileId, permission).
		Context(ctx).
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to create permission: %w", err)
	}

	// Get file to retrieve the web link
	file, err := n.service.Files.Get(fileId).
		Fields("webViewLink, webContentLink").
		Context(ctx).
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to get file link: %w", err)
	}

	return map[string]interface{}{
		"fileId":         fileId,
		"webViewLink":    file.WebViewLink,
		"webContentLink": file.WebContentLink,
		"shared":         true,
	}, nil
}

// createFolder creates a folder in Google Drive
func (n *GoogleDriveNode) createFolder(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get folder name
	folderName, ok := msg.Payload["name"].(string)
	if !ok {
		return nil, fmt.Errorf("folder name is required")
	}

	// Get parent folder ID (optional)
	parentId := n.folderId
	if parent, ok := msg.Payload["parentId"].(string); ok {
		parentId = parent
	}

	// Create folder metadata
	folder := &drive.File{
		Name:     folderName,
		MimeType: "application/vnd.google-apps.folder",
	}

	if parentId != "" {
		folder.Parents = []string{parentId}
	}

	// Create folder
	createdFolder, err := n.service.Files.Create(folder).
		Context(ctx).
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to create folder: %w", err)
	}

	return map[string]interface{}{
		"id":          createdFolder.Id,
		"name":        createdFolder.Name,
		"webViewLink": createdFolder.WebViewLink,
		"created":     true,
	}, nil
}

// getMetadata retrieves metadata for a file
func (n *GoogleDriveNode) getMetadata(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get file ID
	fileId, ok := msg.Payload["fileId"].(string)
	if !ok {
		return nil, fmt.Errorf("fileId is required for getMetadata")
	}

	// Get file metadata
	file, err := n.service.Files.Get(fileId).
		Fields("id, name, mimeType, size, createdTime, modifiedTime, owners, webViewLink, webContentLink").
		Context(ctx).
		Do()

	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	owners := make([]string, 0, len(file.Owners))
	for _, owner := range file.Owners {
		owners = append(owners, owner.EmailAddress)
	}

	return map[string]interface{}{
		"id":              file.Id,
		"name":            file.Name,
		"mimeType":        file.MimeType,
		"size":            file.Size,
		"createdTime":     file.CreatedTime,
		"modifiedTime":    file.ModifiedTime,
		"owners":          owners,
		"webViewLink":     file.WebViewLink,
		"webContentLink":  file.WebContentLink,
	}, nil
}

// Cleanup closes the Google Drive connection
func (n *GoogleDriveNode) Cleanup() error {
	// No explicit cleanup needed for Drive service
	return nil
}
