package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/edgeflow/edgeflow/internal/node"
)

// AWSS3Node handles AWS S3 operations
type AWSS3Node struct {
	client      *s3.S3
	bucket      string
	region      string
	accessKey   string
	secretKey   string
	prefix      string
}

// NewAWSS3Node creates a new AWS S3 node
func NewAWSS3Node() *AWSS3Node {
	return &AWSS3Node{}
}

// Init initializes the AWS S3 node
func (n *AWSS3Node) Init(config map[string]interface{}) error {
	// Parse region
	if region, ok := config["region"].(string); ok {
		n.region = region
	} else {
		n.region = "us-east-1" // Default region
	}

	// Parse access key
	if accessKey, ok := config["accessKey"].(string); ok {
		n.accessKey = accessKey
	} else {
		return fmt.Errorf("AWS access key is required")
	}

	// Parse secret key
	if secretKey, ok := config["secretKey"].(string); ok {
		n.secretKey = secretKey
	} else {
		return fmt.Errorf("AWS secret key is required")
	}

	// Parse bucket name
	if bucket, ok := config["bucket"].(string); ok {
		n.bucket = bucket
	} else {
		return fmt.Errorf("bucket name is required")
	}

	// Parse optional prefix
	if prefix, ok := config["prefix"].(string); ok {
		n.prefix = prefix
	}

	// Create AWS session
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(n.region),
		Credentials: credentials.NewStaticCredentials(
			n.accessKey,
			n.secretKey,
			"", // token
		),
	})

	if err != nil {
		return fmt.Errorf("failed to create AWS session: %w", err)
	}

	// Create S3 client
	n.client = s3.New(sess)

	// Verify bucket exists
	_, err = n.client.HeadBucket(&s3.HeadBucketInput{
		Bucket: aws.String(n.bucket),
	})

	if err != nil {
		return fmt.Errorf("failed to access bucket: %w", err)
	}

	return nil
}

// Execute processes incoming messages
func (n *AWSS3Node) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	if n.client == nil {
		return node.Message{}, fmt.Errorf("AWS S3 client not initialized")
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
		result, err = n.listObjects(ctx, msg)
	case "delete":
		result, err = n.deleteObject(ctx, msg)
	case "getSignedUrl":
		result, err = n.getSignedURL(ctx, msg)
	case "copy":
		result, err = n.copyObject(ctx, msg)
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

// uploadFile uploads a file to S3
func (n *AWSS3Node) uploadFile(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get file path or content
	var fileReader io.Reader
	var contentLength int64

	if filePath, ok := msg.Payload["filePath"].(string); ok {
		// Upload from file
		file, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}

		fileReader = file
		contentLength = fileInfo.Size()
	} else if content, ok := msg.Payload["content"].(string); ok {
		// Upload from string content
		fileReader = bytes.NewReader([]byte(content))
		contentLength = int64(len(content))
	} else {
		return nil, fmt.Errorf("either filePath or content is required for upload")
	}

	// Get key (object name)
	key, ok := msg.Payload["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key is required for upload")
	}

	if n.prefix != "" {
		key = n.prefix + "/" + key
	}

	// Get optional content type
	contentType := "application/octet-stream"
	if ct, ok := msg.Payload["contentType"].(string); ok {
		contentType = ct
	}

	// Get optional ACL
	acl := "private"
	if a, ok := msg.Payload["acl"].(string); ok {
		acl = a
	}

	// Upload object
	_, err := n.client.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(n.bucket),
		Key:           aws.String(key),
		Body:          aws.ReadSeekCloser(fileReader),
		ContentLength: aws.Int64(contentLength),
		ContentType:   aws.String(contentType),
		ACL:           aws.String(acl),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to upload object: %w", err)
	}

	return map[string]interface{}{
		"bucket":   n.bucket,
		"key":      key,
		"size":     contentLength,
		"uploaded": true,
	}, nil
}

// downloadFile downloads a file from S3
func (n *AWSS3Node) downloadFile(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get key
	key, ok := msg.Payload["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key is required for download")
	}

	if n.prefix != "" {
		key = n.prefix + "/" + key
	}

	// Get destination path
	destPath, ok := msg.Payload["destPath"].(string)
	if !ok {
		return nil, fmt.Errorf("destPath is required for download")
	}

	// Download object
	result, err := n.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: aws.String(n.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}
	defer result.Body.Close()

	// Create destination file
	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	// Copy content
	size, err := io.Copy(destFile, result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return map[string]interface{}{
		"bucket":     n.bucket,
		"key":        key,
		"destPath":   destPath,
		"size":       size,
		"downloaded": true,
	}, nil
}

// listObjects lists objects in S3 bucket
func (n *AWSS3Node) listObjects(ctx context.Context, msg node.Message) ([]map[string]interface{}, error) {
	// Get prefix
	prefix := n.prefix
	if p, ok := msg.Payload["prefix"].(string); ok {
		if n.prefix != "" {
			prefix = n.prefix + "/" + p
		} else {
			prefix = p
		}
	}

	// Get max keys
	maxKeys := int64(100)
	if mk, ok := msg.Payload["maxKeys"].(float64); ok {
		maxKeys = int64(mk)
	}

	// List objects
	result, err := n.client.ListObjectsV2WithContext(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(n.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int64(maxKeys),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	// Convert to result
	objects := make([]map[string]interface{}, 0, len(result.Contents))
	for _, obj := range result.Contents {
		objects = append(objects, map[string]interface{}{
			"key":          *obj.Key,
			"size":         *obj.Size,
			"lastModified": obj.LastModified.Format(time.RFC3339),
			"etag":         *obj.ETag,
		})
	}

	return objects, nil
}

// deleteObject deletes an object from S3
func (n *AWSS3Node) deleteObject(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get key
	key, ok := msg.Payload["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key is required for delete")
	}

	if n.prefix != "" {
		key = n.prefix + "/" + key
	}

	// Delete object
	_, err := n.client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(n.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to delete object: %w", err)
	}

	return map[string]interface{}{
		"bucket":  n.bucket,
		"key":     key,
		"deleted": true,
	}, nil
}

// getSignedURL generates a presigned URL for sharing
func (n *AWSS3Node) getSignedURL(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get key
	key, ok := msg.Payload["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key is required for getSignedUrl")
	}

	if n.prefix != "" {
		key = n.prefix + "/" + key
	}

	// Get expiration (default 1 hour)
	expiration := 1 * time.Hour
	if exp, ok := msg.Payload["expirationMinutes"].(float64); ok {
		expiration = time.Duration(exp) * time.Minute
	}

	// Create presigned URL request
	req, _ := n.client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(n.bucket),
		Key:    aws.String(key),
	})

	// Generate presigned URL
	url, err := req.Presign(expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return map[string]interface{}{
		"bucket":     n.bucket,
		"key":        key,
		"url":        url,
		"expiration": expiration.String(),
	}, nil
}

// copyObject copies an object within S3
func (n *AWSS3Node) copyObject(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get source key
	sourceKey, ok := msg.Payload["sourceKey"].(string)
	if !ok {
		return nil, fmt.Errorf("sourceKey is required for copy")
	}

	// Get destination key
	destKey, ok := msg.Payload["destKey"].(string)
	if !ok {
		return nil, fmt.Errorf("destKey is required for copy")
	}

	if n.prefix != "" {
		sourceKey = n.prefix + "/" + sourceKey
		destKey = n.prefix + "/" + destKey
	}

	// Copy object
	_, err := n.client.CopyObjectWithContext(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(n.bucket),
		CopySource: aws.String(n.bucket + "/" + sourceKey),
		Key:        aws.String(destKey),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to copy object: %w", err)
	}

	return map[string]interface{}{
		"bucket":    n.bucket,
		"sourceKey": sourceKey,
		"destKey":   destKey,
		"copied":    true,
	}, nil
}

// getMetadata retrieves metadata for an object
func (n *AWSS3Node) getMetadata(ctx context.Context, msg node.Message) (map[string]interface{}, error) {
	// Get key
	key, ok := msg.Payload["key"].(string)
	if !ok {
		return nil, fmt.Errorf("key is required for getMetadata")
	}

	if n.prefix != "" {
		key = n.prefix + "/" + key
	}

	// Get object metadata
	result, err := n.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(n.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	metadata := make(map[string]interface{})
	for k, v := range result.Metadata {
		metadata[k] = *v
	}

	return map[string]interface{}{
		"bucket":       n.bucket,
		"key":          key,
		"contentType":  *result.ContentType,
		"contentLength": *result.ContentLength,
		"lastModified": result.LastModified.Format(time.RFC3339),
		"etag":         *result.ETag,
		"metadata":     metadata,
	}, nil
}

// Cleanup closes the AWS S3 connection
func (n *AWSS3Node) Cleanup() error {
	// No explicit cleanup needed for S3 client
	return nil
}
