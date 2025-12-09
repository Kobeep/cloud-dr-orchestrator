package oracle

import (
	"context"
	"fmt"
	"time"

	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// ObjectInfo contains information about an object in Object Storage
type ObjectInfo struct {
	Name         string
	Size         int64
	LastModified time.Time
	ETag         string
}

// ListObjects lists all objects in the bucket with an optional prefix
func (c *Client) ListObjects(ctx context.Context, prefix string) ([]ObjectInfo, error) {
	var objects []ObjectInfo

	request := objectstorage.ListObjectsRequest{
		NamespaceName: &c.namespace,
		BucketName:    &c.bucketName,
		Prefix:        &prefix,
	}

	response, err := c.objectStorageClient.ListObjects(ctx, request)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	for _, obj := range response.Objects {
		if obj.Name != nil {
			info := ObjectInfo{
				Name:         *obj.Name,
				Size:         *obj.Size,
				LastModified: obj.TimeModified.Time,
			}
			if obj.Etag != nil {
				info.ETag = *obj.Etag
			}
			objects = append(objects, info)
		}
	}

	return objects, nil
}

// ListBackups lists all backup files (objects with 'backups/' prefix)
func (c *Client) ListBackups(ctx context.Context) ([]ObjectInfo, error) {
	return c.ListObjects(ctx, "backups/")
}

// ListBackupsByDate lists backups for a specific year and month
func (c *Client) ListBackupsByDate(ctx context.Context, year int, month int) ([]ObjectInfo, error) {
	prefix := fmt.Sprintf("backups/%d/%02d/", year, month)
	return c.ListObjects(ctx, prefix)
}
