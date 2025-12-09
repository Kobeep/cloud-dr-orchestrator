package oracle

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// UploadResult contains information about the uploaded file
type UploadResult struct {
	ObjectName string
	BucketName string
	Namespace  string
	Size       int64
	Duration   time.Duration
	ETag       string
}

// UploadFile uploads a local file to Oracle Cloud Object Storage
// It reads the file from localPath and uploads it with the given objectName
func (c *Client) UploadFile(ctx context.Context, localPath string, objectName string) (*UploadResult, error) {
	startTime := time.Now()

	// Open the file
	file, err := os.Open(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", localPath, err)
	}
	defer file.Close()

	// Get file info for size
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// If object name is not provided, use the filename
	if objectName == "" {
		objectName = filepath.Base(localPath)
	}

	// Get file size for ContentLength
	fileSize := fileInfo.Size()

	// Create the put object request
	request := objectstorage.PutObjectRequest{
		NamespaceName: &c.namespace,
		BucketName:    &c.bucketName,
		ObjectName:    &objectName,
		ContentLength: &fileSize,
		PutObjectBody: file,
	}

	// Upload the file
	response, err := c.objectStorageClient.PutObject(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file to Object Storage: %w", err)
	}

	duration := time.Since(startTime)

	result := &UploadResult{
		ObjectName: objectName,
		BucketName: c.bucketName,
		Namespace:  c.namespace,
		Size:       fileInfo.Size(),
		Duration:   duration,
		ETag:       *response.ETag,
	}

	return result, nil
}

// UploadBackup is a convenience function that uploads a backup file
// It automatically generates the object name from the local file path
func (c *Client) UploadBackup(ctx context.Context, backupPath string) (*UploadResult, error) {
	// Extract filename from path for object name
	filename := filepath.Base(backupPath)

	// Create a folder structure: backups/YYYY/MM/filename
	now := time.Now()
	objectName := fmt.Sprintf("backups/%d/%02d/%s", now.Year(), now.Month(), filename)

	return c.UploadFile(ctx, backupPath, objectName)
}
