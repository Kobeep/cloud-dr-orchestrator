package oracle

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// DownloadResult contains information about the downloaded file
type DownloadResult struct {
	ObjectName   string
	LocalPath    string
	Size         int64
	Duration     time.Duration
	LastModified time.Time
}

// DownloadFile downloads an object from Oracle Cloud Object Storage to a local file
func (c *Client) DownloadFile(ctx context.Context, objectName string, localPath string) (*DownloadResult, error) {
	startTime := time.Now()

	// Create the get object request
	request := objectstorage.GetObjectRequest{
		NamespaceName: &c.namespace,
		BucketName:    &c.bucketName,
		ObjectName:    &objectName,
	}

	// Download the object
	response, err := c.objectStorageClient.GetObject(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to download object %s: %w", objectName, err)
	}
	defer response.Content.Close()

	// Create the local file
	outFile, err := os.Create(localPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file %s: %w", localPath, err)
	}
	defer outFile.Close()

	// Copy the content to the local file
	bytesWritten, err := io.Copy(outFile, response.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to write content to file: %w", err)
	}

	duration := time.Since(startTime)

	result := &DownloadResult{
		ObjectName:   objectName,
		LocalPath:    localPath,
		Size:         bytesWritten,
		Duration:     duration,
		LastModified: response.LastModified.Time,
	}

	return result, nil
}

// DownloadBackup is a convenience function that downloads a backup file
// It downloads from the specified object name to the local path
func (c *Client) DownloadBackup(ctx context.Context, objectName string, localPath string) (*DownloadResult, error) {
	return c.DownloadFile(ctx, objectName, localPath)
}
