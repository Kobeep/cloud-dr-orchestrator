package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Kobeep/cloud-dr-orchestrator/pkg/oracle"
	"github.com/spf13/cobra"
)

var uploadCmd = &cobra.Command{
	Use:   "upload",
	Short: "Upload a backup file to Oracle Cloud Object Storage",
	Long: `Upload a local backup file to Oracle Cloud Object Storage.
The file will be organized in a date-based folder structure (backups/YYYY/MM/filename).

Example:
  orchestrator upload --file backup-20251209.tar.gz`,
	RunE: runUpload,
}

var (
	uploadFile       string
	uploadObjectName string
	ociConfigFile    string
	ociProfile       string
	ociBucket        string
	ociNamespace     string
	ociCompartment   string
)

func init() {
	rootCmd.AddCommand(uploadCmd)

	uploadCmd.Flags().StringVar(&uploadFile, "file", "", "Path to the backup file to upload (required)")
	uploadCmd.Flags().StringVar(&uploadObjectName, "object-name", "", "Custom object name in Object Storage (optional, uses filename if not set)")
	uploadCmd.Flags().StringVar(&ociConfigFile, "oci-config", "", "Path to OCI config file (default: ~/.oci/config)")
	uploadCmd.Flags().StringVar(&ociProfile, "oci-profile", "DEFAULT", "OCI config profile to use")
	uploadCmd.Flags().StringVar(&ociBucket, "bucket", "", "OCI Object Storage bucket name (required)")
	uploadCmd.Flags().StringVar(&ociNamespace, "namespace", "", "OCI namespace (auto-detected if not provided)")
	uploadCmd.Flags().StringVar(&ociCompartment, "compartment", "", "OCI compartment ID (required)")

	uploadCmd.MarkFlagRequired("file")
	uploadCmd.MarkFlagRequired("bucket")
	uploadCmd.MarkFlagRequired("compartment")
}

func runUpload(cmd *cobra.Command, args []string) error {
	// Validate file exists
	if _, err := os.Stat(uploadFile); os.IsNotExist(err) {
		return fmt.Errorf("file does not exist: %s", uploadFile)
	}

	fmt.Printf("🔗 Connecting to Oracle Cloud...\n")

	// Create OCI client
	config := oracle.Config{
		ConfigFilePath: ociConfigFile,
		Profile:        ociProfile,
		Namespace:      ociNamespace,
		BucketName:     ociBucket,
		CompartmentID:  ociCompartment,
	}

	client, err := oracle.NewClient(config)
	if err != nil {
		return fmt.Errorf("failed to create OCI client: %w", err)
	}

	fmt.Printf("✓ Connected to namespace: %s\n", client.GetNamespace())
	fmt.Printf("📤 Uploading file: %s\n", uploadFile)

	// Upload the file
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	var result *oracle.UploadResult
	if uploadObjectName != "" {
		result, err = client.UploadFile(ctx, uploadFile, uploadObjectName)
	} else {
		result, err = client.UploadBackup(ctx, uploadFile)
	}

	if err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Print success message
	fmt.Printf("\n✓ Upload successful!\n")
	fmt.Printf("  Object: %s\n", result.ObjectName)
	fmt.Printf("  Bucket: %s\n", result.BucketName)
	fmt.Printf("  Size: %.2f MB\n", float64(result.Size)/1024/1024)
	fmt.Printf("  Duration: %s\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("  ETag: %s\n", result.ETag)

	return nil
}
