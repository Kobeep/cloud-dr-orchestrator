package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Kobeep/cloud-dr-orchestrator/pkg/backup"
	"github.com/Kobeep/cloud-dr-orchestrator/pkg/encryption"
	"github.com/Kobeep/cloud-dr-orchestrator/pkg/metrics"
	"github.com/spf13/cobra"
)

var (
	backupSource  string
	backupName    string
	dbHost        string
	dbPort        int
	dbUser        string
	dbPassword    string
	dbName        string
	outputDir     string
	encryptBackup bool
	encryptionKey string
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup of PostgreSQL database",
	Long:  `Dump PostgreSQL database, compress it to .tar.gz and optionally upload to Oracle Cloud.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if backupSource != "postgres" {
			return fmt.Errorf("only 'postgres' source is currently supported")
		}

		// Start timing for metrics
		startTime := time.Now()

		// Prepare PostgreSQL config
		config := backup.PostgresConfig{
			Host:     dbHost,
			Port:     dbPort,
			User:     dbUser,
			Password: dbPassword,
			Database: dbName,
		}

		// Resolve output directory
		absOutputDir, err := filepath.Abs(outputDir)
		if err != nil {
			return fmt.Errorf("invalid output directory: %w", err)
		}

		fmt.Printf("Starting backup: %s\n", backupName)
		fmt.Printf("Output directory: %s\n", absOutputDir)
		fmt.Println()

		// Check for encryption key from environment if not provided
		if encryptBackup && encryptionKey == "" {
			encryptionKey = os.Getenv("BACKUP_ENCRYPTION_KEY")
		}

		// Create backup
		result, err := backup.DumpPostgres(config, backupName, absOutputDir)
		if err != nil {
			// Record failure metrics
			metrics.BackupFailure.WithLabelValues("dump_failed").Inc()
			metrics.RecordBackupError(err)
			return fmt.Errorf("backup failed: %w", err)
		}

		// Record success metrics
		duration := time.Since(startTime).Seconds()
		metrics.BackupDuration.Observe(duration)
		metrics.BackupSuccess.Inc()
		metrics.RecordBackupSuccess()

		// Get file size for metrics
		fileInfo, err := os.Stat(result.FilePath)
		if err == nil {
			metrics.BackupSize.Observe(float64(fileInfo.Size()))
		}

		finalPath := result.FilePath

		// Encrypt backup if requested
		if encryptBackup {
			if encryptionKey == "" {
				metrics.BackupFailure.WithLabelValues("missing_encryption_key").Inc()
				return fmt.Errorf("encryption key required when --encrypt is enabled")
			}

			fmt.Printf("🔐 Encrypting backup...\n")
			encryptedPath, err := encryption.EncryptFile(result.FilePath, encryptionKey)
			if err != nil {
				metrics.BackupFailure.WithLabelValues("encryption_failed").Inc()
				return fmt.Errorf("encryption failed: %w", err)
			}

			// Remove unencrypted file
			if err := os.Remove(result.FilePath); err != nil {
				fmt.Printf("⚠️  Warning: failed to remove unencrypted file: %v\n", err)
			}

			finalPath = encryptedPath
			fmt.Printf("✅ Backup encrypted\n")
		}

		fmt.Printf("\n📦 Backup file: %s\n", finalPath)
		fmt.Printf("⏱️  Duration: %.2fs\n", duration)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVar(&backupSource, "source", "postgres", "Backup source type (postgres|filesystem)")
	backupCmd.Flags().StringVar(&backupName, "name", "", "Backup name (required)")
	backupCmd.MarkFlagRequired("name")

	backupCmd.Flags().StringVar(&dbHost, "db-host", "localhost", "PostgreSQL host")
	backupCmd.Flags().IntVar(&dbPort, "db-port", 5432, "PostgreSQL port")
	backupCmd.Flags().StringVar(&dbUser, "db-user", "postgres", "PostgreSQL user")
	backupCmd.Flags().StringVar(&dbPassword, "db-password", "", "PostgreSQL password")
	backupCmd.Flags().StringVar(&dbName, "db-name", "", "PostgreSQL database name (required)")
	backupCmd.MarkFlagRequired("db-name")

	backupCmd.Flags().StringVar(&outputDir, "output", "./backups", "Output directory for backups")

	// Encryption flags
	backupCmd.Flags().BoolVar(&encryptBackup, "encrypt", false, "Encrypt backup file")
	backupCmd.Flags().StringVar(&encryptionKey, "encryption-key", "", "Encryption key (or use BACKUP_ENCRYPTION_KEY env var)")
}
