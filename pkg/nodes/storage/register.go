package storage

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes registers all storage nodes
func RegisterAllNodes(registry *node.Registry) {
	// Google Drive
	registry.Register(&node.NodeInfo{
		Type:        "google-drive",
		Name:        "Google Drive",
		Category:    node.NodeTypeProcessing,
		Description: "Upload, download, list, share files on Google Drive",
		Icon:        "cloud",
		Properties: []node.PropertySchema{
			{Name: "credentials", Label: "Credentials", Type: "string", Default: "", Required: true},
			{Name: "token", Label: "Token", Type: "password", Default: "", Required: false},
			{Name: "folderId", Label: "Folder ID", Type: "string", Default: "", Required: false},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewGoogleDriveNode()
		},
	})

	// AWS S3
	registry.Register(&node.NodeInfo{
		Type:        "aws-s3",
		Name:        "AWS S3",
		Category:    node.NodeTypeProcessing,
		Description: "Upload, download, list objects in AWS S3",
		Icon:        "cloud",
		Properties: []node.PropertySchema{
			{Name: "region", Label: "Region", Type: "string", Default: "us-east-1", Required: true},
			{Name: "accessKey", Label: "Access Key", Type: "password", Default: "", Required: true},
			{Name: "secretKey", Label: "Secret Key", Type: "password", Default: "", Required: true},
			{Name: "bucket", Label: "Bucket", Type: "string", Default: "", Required: true},
			{Name: "prefix", Label: "Prefix", Type: "string", Default: "", Required: false},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewAWSS3Node()
		},
	})

	// SFTP
	registry.Register(&node.NodeInfo{
		Type:        "sftp",
		Name:        "SFTP",
		Category:    node.NodeTypeProcessing,
		Description: "Upload, download files via SFTP (SSH) protocol",
		Icon:        "lock",
		Color:       "#6366f1",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true, Description: "SSH server hostname or IP"},
			{Name: "port", Label: "Port", Type: "number", Default: 22, Required: true, Description: "SSH port number"},
			{Name: "username", Label: "Username", Type: "string", Default: "", Required: true, Description: "SSH username"},
			{Name: "password", Label: "Password", Type: "password", Default: "", Description: "SSH password"},
			{Name: "privateKey", Label: "Private Key", Type: "string", Default: "", Description: "SSH private key (PEM format)"},
			{Name: "passphrase", Label: "Key Passphrase", Type: "password", Default: "", Description: "Passphrase for encrypted private key"},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewSFTPNode()
		},
	})

	// Dropbox
	registry.Register(&node.NodeInfo{
		Type:        "dropbox",
		Name:        "Dropbox",
		Category:    node.NodeTypeProcessing,
		Description: "Upload, download, manage files on Dropbox",
		Icon:        "cloud",
		Color:       "#0061FF",
		Properties: []node.PropertySchema{
			{Name: "accessToken", Label: "Access Token", Type: "password", Default: "", Required: true, Description: "Dropbox API access token"},
			{Name: "rootPath", Label: "Root Path", Type: "string", Default: "", Description: "Root path for operations"},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewDropboxNode()
		},
	})

	// OneDrive
	registry.Register(&node.NodeInfo{
		Type:        "onedrive",
		Name:        "OneDrive",
		Category:    node.NodeTypeProcessing,
		Description: "Upload, download, manage files on Microsoft OneDrive",
		Icon:        "cloud",
		Color:       "#0078D4",
		Properties: []node.PropertySchema{
			{Name: "accessToken", Label: "Access Token", Type: "password", Default: "", Required: true, Description: "Microsoft Graph API access token"},
			{Name: "driveId", Label: "Drive ID", Type: "string", Default: "", Description: "Drive ID (empty for default)"},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewOneDriveNode()
		},
	})

	// FTP
	registry.Register(&node.NodeInfo{
		Type:        "ftp",
		Name:        "FTP",
		Category:    node.NodeTypeProcessing,
		Description: "Upload, download files via FTP protocol",
		Icon:        "server",
		Properties: []node.PropertySchema{
			{Name: "host", Label: "Host", Type: "string", Default: "localhost", Required: true},
			{Name: "port", Label: "Port", Type: "number", Default: 21, Required: true},
			{Name: "username", Label: "Username", Type: "string", Default: "anonymous", Required: true},
			{Name: "password", Label: "Password", Type: "password", Default: "", Required: false},
		},
		Inputs: []node.PortSchema{
			{Name: "in", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "out", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewFTPNode()
		},
	})
}
