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
			{Name: "token", Label: "Token", Type: "string", Default: "", Required: false},
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
			{Name: "accessKey", Label: "Access Key", Type: "string", Default: "", Required: true},
			{Name: "secretKey", Label: "Secret Key", Type: "string", Default: "", Required: true},
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
			{Name: "password", Label: "Password", Type: "string", Default: "", Required: false},
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
