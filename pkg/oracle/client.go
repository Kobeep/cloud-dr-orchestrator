package oracle

import (
	"context"
	"fmt"
	"os"

	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/objectstorage"
)

// Client represents an Oracle Cloud Infrastructure client for Object Storage operations
type Client struct {
	objectStorageClient *objectstorage.ObjectStorageClient
	namespace           string
	bucketName          string
	compartmentID       string
}

// Config holds the configuration for OCI client
type Config struct {
	ConfigFilePath string
	Profile        string
	Namespace      string
	BucketName     string
	CompartmentID  string
}

// NewClient creates a new OCI Object Storage client
// It reads credentials from ~/.oci/config or uses environment variables
func NewClient(config Config) (*Client, error) {
	var configProvider common.ConfigurationProvider
	var err error

	// Try to load from config file first
	if config.ConfigFilePath == "" {
		// Use default OCI config location: ~/.oci/config
		homeDir, err := os.UserHomeDir()
		if err == nil {
			config.ConfigFilePath = homeDir + "/.oci/config"
		}
	}
	if config.Profile == "" {
		config.Profile = "DEFAULT"
	}

	configProvider, err = common.ConfigurationProviderFromFileWithProfile(
		config.ConfigFilePath,
		config.Profile,
		"",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load OCI config: %w", err)
	}

	// Create Object Storage client
	client, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(configProvider)
	if err != nil {
		return nil, fmt.Errorf("failed to create Object Storage client: %w", err)
	}

	// Get namespace if not provided
	namespace := config.Namespace
	if namespace == "" {
		ctx := context.Background()
		request := objectstorage.GetNamespaceRequest{}
		response, err := client.GetNamespace(ctx, request)
		if err != nil {
			return nil, fmt.Errorf("failed to get namespace: %w", err)
		}
		namespace = *response.Value
	}

	return &Client{
		objectStorageClient: &client,
		namespace:           namespace,
		bucketName:          config.BucketName,
		compartmentID:       config.CompartmentID,
	}, nil
}

// GetBucketName returns the configured bucket name
func (c *Client) GetBucketName() string {
	return c.bucketName
}

// GetNamespace returns the OCI namespace
func (c *Client) GetNamespace() string {
	return c.namespace
}
