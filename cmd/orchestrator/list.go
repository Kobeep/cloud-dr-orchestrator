package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Kobeep/cloud-dr-orchestrator/pkg/oracle"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List backup files in Oracle Cloud Object Storage",
	Long: `List all backup files stored in Oracle Cloud Object Storage.
You can optionally filter by year and month.

Examples:
  orchestrator list                          # List all backups
  orchestrator list --year 2025 --month 12  # List backups from December 2025`,
	RunE: runList,
}

var (
	listYear  int
	listMonth int
	listAll   bool
)

func init() {
	rootCmd.AddCommand(listCmd)

	listCmd.Flags().IntVar(&listYear, "year", 0, "Filter backups by year")
	listCmd.Flags().IntVar(&listMonth, "month", 0, "Filter backups by month (requires --year)")
	listCmd.Flags().BoolVar(&listAll, "all", false, "List all objects in bucket (not just backups)")
	listCmd.Flags().StringVar(&ociConfigFile, "oci-config", "", "Path to OCI config file (default: ~/.oci/config)")
	listCmd.Flags().StringVar(&ociProfile, "oci-profile", "DEFAULT", "OCI config profile to use")
	listCmd.Flags().StringVar(&ociBucket, "bucket", "", "OCI Object Storage bucket name (required)")
	listCmd.Flags().StringVar(&ociNamespace, "namespace", "", "OCI namespace (auto-detected if not provided)")
	listCmd.Flags().StringVar(&ociCompartment, "compartment", "", "OCI compartment ID (required)")

	listCmd.MarkFlagRequired("bucket")
	listCmd.MarkFlagRequired("compartment")
}

func runList(cmd *cobra.Command, args []string) error {
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
	fmt.Printf("📋 Listing backups from bucket: %s\n\n", client.GetBucketName())

	// List objects
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	var objects []oracle.ObjectInfo
	if listAll {
		objects, err = client.ListObjects(ctx, "")
	} else if listYear > 0 && listMonth > 0 {
		objects, err = client.ListBackupsByDate(ctx, listYear, listMonth)
	} else if listYear > 0 {
		prefix := fmt.Sprintf("backups/%d/", listYear)
		objects, err = client.ListObjects(ctx, prefix)
	} else {
		objects, err = client.ListBackups(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to list objects: %w", err)
	}

	// Display results
	if len(objects) == 0 {
		fmt.Println("No backups found.")
		return nil
	}

	fmt.Printf("Found %d backup(s):\n\n", len(objects))

	var totalSize int64
	for i, obj := range objects {
		fmt.Printf("%d. %s\n", i+1, obj.Name)
		fmt.Printf("   Size: %.2f MB\n", float64(obj.Size)/1024/1024)
		fmt.Printf("   Modified: %s\n", obj.LastModified.Format("2006-01-02 15:04:05"))
		fmt.Printf("   ETag: %s\n\n", obj.ETag)
		totalSize += obj.Size
	}

	fmt.Printf("Total: %d file(s), %.2f MB\n", len(objects), float64(totalSize)/1024/1024)

	return nil
}
