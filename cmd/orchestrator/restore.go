package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Kobeep/cloud-dr-orchestrator/pkg/backup"
	"github.com/Kobeep/cloud-dr-orchestrator/pkg/oracle"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore PostgreSQL database from backup",
	Long: `Restore a PostgreSQL database from a local .tar.gz backup file or
download from Oracle Cloud Object Storage and restore.

Examples:
  # Restore from local backup file
  orchestrator restore --file backup-20251209.tar.gz --db-name mydb --db-host localhost --db-user postgres --db-password secret

  # Download from cloud and restore
  orchestrator restore --from-cloud backups/2025/12/backup-20251209.tar.gz --bucket my-bucket --compartment ocid1... --db-name mydb --db-host localhost --db-user postgres --db-password secret

  # Restore to different target database
  orchestrator restore --file backup.tar.gz --db-name mydb --target-db mydb_restored --db-host localhost --db-user postgres --db-password secret
`,
	RunE: runRestore,
}

var (
	restoreFile        string
	restoreFromCloud   string
	restoreTargetDB    string
	restoreDBName      string
	restoreDBHost      string
	restoreDBPort      int
	restoreDBUser      string
	restoreDBPassword  string
	restoreBucket      string
	restoreCompartment string
	restoreOCIConfig   string
	restoreOCIProfile  string
	restoreSkipConfirm bool
)

func init() {
	rootCmd.AddCommand(restoreCmd)

	// Backup file flags
	restoreCmd.Flags().StringVar(&restoreFile, "file", "", "Local backup file path (.tar.gz)")
	restoreCmd.Flags().StringVar(&restoreFromCloud, "from-cloud", "", "Download backup from cloud (object path in bucket)")

	// Database connection flags
	restoreCmd.Flags().StringVar(&restoreDBName, "db-name", "", "Database name to restore to (required)")
	restoreCmd.Flags().StringVar(&restoreTargetDB, "target-db", "", "Target database name (if different from source)")
	restoreCmd.Flags().StringVar(&restoreDBHost, "db-host", "localhost", "Database host")
	restoreCmd.Flags().IntVar(&restoreDBPort, "db-port", 5432, "Database port")
	restoreCmd.Flags().StringVar(&restoreDBUser, "db-user", "postgres", "Database user")
	restoreCmd.Flags().StringVar(&restoreDBPassword, "db-password", "", "Database password")

	// Oracle Cloud flags (only needed if --from-cloud is used)
	restoreCmd.Flags().StringVar(&restoreBucket, "bucket", "", "OCI Object Storage bucket name")
	restoreCmd.Flags().StringVar(&restoreCompartment, "compartment", "", "OCI compartment OCID")
	restoreCmd.Flags().StringVar(&restoreOCIConfig, "oci-config", "", "OCI config file path (default: ~/.oci/config)")
	restoreCmd.Flags().StringVar(&restoreOCIProfile, "oci-profile", "DEFAULT", "OCI config profile")

	// Safety flag
	restoreCmd.Flags().BoolVar(&restoreSkipConfirm, "yes", false, "Skip confirmation prompt")

	// Required flags
	restoreCmd.MarkFlagRequired("db-name")
}

func runRestore(cmd *cobra.Command, args []string) error {
	// Validate flags
	if restoreFile == "" && restoreFromCloud == "" {
		return fmt.Errorf("either --file or --from-cloud must be specified")
	}
	if restoreFile != "" && restoreFromCloud != "" {
		return fmt.Errorf("cannot specify both --file and --from-cloud")
	}

	// If downloading from cloud, validate cloud flags
	if restoreFromCloud != "" {
		if restoreBucket == "" || restoreCompartment == "" {
			return fmt.Errorf("--bucket and --compartment are required when using --from-cloud")
		}
	}

	// Build PostgreSQL config
	pgConfig := backup.PostgresConfig{
		Host:     restoreDBHost,
		Port:     restoreDBPort,
		User:     restoreDBUser,
		Password: restoreDBPassword,
		Database: restoreDBName,
	}

	// Determine backup file path
	var backupFilePath string
	var cleanupFile bool

	if restoreFromCloud != "" {
		// Download from Oracle Cloud
		fmt.Printf("📥 Downloading backup from Oracle Cloud...\n")
		fmt.Printf("   Bucket: %s\n", restoreBucket)
		fmt.Printf("   Object: %s\n", restoreFromCloud)

		// Initialize Oracle Cloud client
		client, err := oracle.NewClient(restoreOCIConfig, restoreOCIProfile, restoreCompartment)
		if err != nil {
			return fmt.Errorf("failed to initialize Oracle Cloud client: %w", err)
		}

		// Create temporary directory
		tempDir, err := os.MkdirTemp("", "orchestrator-restore-*")
		if err != nil {
			return fmt.Errorf("failed to create temp directory: %w", err)
		}
		defer os.RemoveAll(tempDir)

		// Download file
		backupFilePath = filepath.Join(tempDir, filepath.Base(restoreFromCloud))
		if err := client.DownloadObject(restoreBucket, restoreFromCloud, backupFilePath); err != nil {
			return fmt.Errorf("failed to download backup: %w", err)
		}
		cleanupFile = true
		fmt.Printf("✅ Downloaded to: %s\n\n", backupFilePath)
	} else {
		// Use local file
		backupFilePath = restoreFile
		if _, err := os.Stat(backupFilePath); os.IsNotExist(err) {
			return fmt.Errorf("backup file not found: %s", backupFilePath)
		}
	}

	// Show restore plan
	fmt.Printf("🔄 Restore Plan:\n")
	fmt.Printf("   Backup file: %s\n", backupFilePath)
	fmt.Printf("   Target host: %s:%d\n", pgConfig.Host, pgConfig.Port)
	fmt.Printf("   Target database: %s\n", pgConfig.Database)
	if restoreTargetDB != "" {
		fmt.Printf("   Will restore as: %s\n", restoreTargetDB)
	}
	fmt.Printf("\n")

	// Confirmation prompt
	if !restoreSkipConfirm {
		fmt.Printf("⚠️  WARNING: This will overwrite the database '%s'!\n", pgConfig.Database)
		fmt.Printf("Are you sure you want to continue? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}

		response = strings.TrimSpace(strings.ToLower(response))
		if response != "yes" && response != "y" {
			fmt.Println("❌ Restore cancelled.")
			return nil
		}
		fmt.Println()
	}

	// Perform restore
	if err := backup.RestorePostgres(pgConfig, backupFilePath, restoreTargetDB); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	// Cleanup downloaded file if needed
	if cleanupFile {
		os.Remove(backupFilePath)
	}

	return nil
}
