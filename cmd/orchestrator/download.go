package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Kobeep/cloud-dr-orchestrator/pkg/oracle"
	"github.com/spf13/cobra"
)

var downloadCmd = &cobra.Command{
	Use:   "download",
	Short: "Download a backup file from Oracle Cloud Object Storage",
	Long: `Download a backup file from Oracle Cloud Object Storage to a local path.

Example:
  orchestrator download --object backups/2025/12/backup-20251209.tar.gz --output ./backup.tar.gz`,
	RunE: runDownload,
}

var (
	downloadObjectName string
	downloadOutput     string
)

func init() {
	rootCmd.AddCommand(downloadCmd)

	downloadCmd.Flags().StringVar(&downloadObjectName, "object", "", "Object name in Object Storage to download (required)")
	downloadCmd.Flags().StringVar(&downloadOutput, "output", "", "Local path to save the downloaded file (required)")
	downloadCmd.Flags().StringVar(&ociConfigFile, "oci-config", "", "Path to OCI config file (default: ~/.oci/config)")
	downloadCmd.Flags().StringVar(&ociProfile, "oci-profile", "DEFAULT", "OCI config profile to use")
	downloadCmd.Flags().StringVar(&ociBucket, "bucket", "", "OCI Object Storage bucket name (required)")
	downloadCmd.Flags().StringVar(&ociNamespace, "namespace", "", "OCI namespace (auto-detected if not provided)")
	downloadCmd.Flags().StringVar(&ociCompartment, "compartment", "", "OCI compartment ID (required)")

	downloadCmd.MarkFlagRequired("object")
	downloadCmd.MarkFlagRequired("output")
	downloadCmd.MarkFlagRequired("bucket")
	downloadCmd.MarkFlagRequired("compartment")
}

func runDownload(cmd *cobra.Command, args []string) error {
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
	fmt.Printf("📥 Downloading object: %s\n", downloadObjectName)

	// Download the file
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := client.DownloadFile(ctx, downloadObjectName, downloadOutput)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	// Print success message
	fmt.Printf("\n✓ Download successful!\n")
	fmt.Printf("  Object: %s\n", result.ObjectName)
	fmt.Printf("  Local path: %s\n", result.LocalPath)
	fmt.Printf("  Size: %.2f MB\n", float64(result.Size)/1024/1024)
	fmt.Printf("  Duration: %s\n", result.Duration.Round(time.Millisecond))
	fmt.Printf("  Last modified: %s\n", result.LastModified.Format(time.RFC3339))

	return nil
}
