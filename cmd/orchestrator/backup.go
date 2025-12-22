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
	backupType      string
	backupName      string
	backupSources   []string // For file backups
	excludePatterns []string // For file backups
	dbHost          string
	dbPort          int
	dbUser          string
	dbPassword      string
	dbName          string
	outputDir       string
	encryptBackup   bool
	encryptionKey   string
)

var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "Create a backup (database, files, or directories)",
	Long: `Create backups of various types:
  - postgres: PostgreSQL database backup
  - files: Backup specific files or directories
  - mysql: MySQL database backup (coming soon)

Examples:
  # PostgreSQL backup
  orchestrator backup --type postgres --name prod-db --db-name myapp

  # File backup
  orchestrator backup --type files --name configs --source /etc/nginx --source /etc/ssl

  # Directory backup with exclusions
  orchestrator backup --type files --name app-data --source /var/www --exclude "*.log" --exclude "tmp/*"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Start timing for metrics
		startTime := time.Now()

		// Resolve output directory
		absOutputDir, err := filepath.Abs(outputDir)
		if err != nil {
			return fmt.Errorf("invalid output directory: %w", err)
		}

		// Check for encryption key from environment if not provided
		if encryptBackup && encryptionKey == "" {
			encryptionKey = os.Getenv("BACKUP_ENCRYPTION_KEY")
		}

		fmt.Printf("Starting %s backup: %s\n", backupType, backupName)
		fmt.Printf("Output directory: %s\n", absOutputDir)
		fmt.Println()

		var result *backup.Result

		// Create backup based on type
		switch backupType {
		case "postgres":
			result, err = performPostgresBackup(absOutputDir)
		case "files", "directory":
			result, err = performFileBackup(absOutputDir)
		default:
			return fmt.Errorf("unsupported backup type: %s (supported: postgres, files)", backupType)
		}

		if err != nil {
			// Record failure metrics
			metrics.BackupFailure.WithLabelValues("backup_failed").Inc()
			metrics.RecordBackupError(err)
			return fmt.Errorf("backup failed: %w", err)
		}

		// Record success metrics
		duration := time.Since(startTime).Seconds()
		metrics.BackupDuration.Observe(duration)
		metrics.BackupSuccess.Inc()
		metrics.RecordBackupSuccess()

		// Get file size for metrics
		fileInfo, err := os.Stat(result.Path)
		if err == nil {
			metrics.BackupSize.Observe(float64(fileInfo.Size()))
		}

		finalPath := result.Path

		// Encrypt backup if requested
		if encryptBackup {
			if encryptionKey == "" {
				metrics.BackupFailure.WithLabelValues("missing_encryption_key").Inc()
				return fmt.Errorf("encryption key required when --encrypt is enabled")
			}

			fmt.Printf("🔐 Encrypting backup...\n")
			encryptedPath, err := encryption.EncryptFile(result.Path, encryptionKey)
			if err != nil {
				metrics.BackupFailure.WithLabelValues("encryption_failed").Inc()
				return fmt.Errorf("encryption failed: %w", err)
			}

			// Remove unencrypted file
			if err := os.Remove(result.Path); err != nil {
				fmt.Printf("⚠️  Warning: failed to remove unencrypted file: %v\n", err)
			}

			finalPath = encryptedPath
			fmt.Printf("✅ Backup encrypted\n")
		}

		fmt.Printf("\n📦 Backup file: %s\n", finalPath)
		fmt.Printf("📊 Size: %.2f MB", float64(result.Size)/(1024*1024))
		if result.CompressionPct > 0 {
			fmt.Printf(" (%.1f%% compression)", result.CompressionPct)
		}
		fmt.Printf("\n⏱️  Duration: %.2fs\n", duration)

		return nil
	},
}

func performPostgresBackup(outputDir string) (*backup.Result, error) {
	if dbName == "" {
		return nil, fmt.Errorf("--db-name is required for postgres backup")
	}

	config := backup.PostgresConfig{
		Host:     dbHost,
		Port:     dbPort,
		User:     dbUser,
		Password: dbPassword,
		Database: dbName,
	}

	legacyResult, err := backup.DumpPostgres(config, backupName, outputDir)
	if err != nil {
		return nil, err
	}

	// Convert legacy result to new format
	return &backup.Result{
		Type:         backup.TypePostgreSQL,
		Path:         legacyResult.FilePath,
		Size:         legacyResult.CompressedSize,
		OriginalSize: legacyResult.OriginalSize,
		Duration:     legacyResult.Duration,
		DatabaseName: dbName,
		Timestamp:    time.Now(),
	}, nil
}

func performFileBackup(outputDir string) (*backup.Result, error) {
	if len(backupSources) == 0 {
		return nil, fmt.Errorf("--source is required for files backup (can be specified multiple times)")
	}

	fileBackup := &backup.FileBackup{
		Name:            backupName,
		Sources:         backupSources,
		ExcludePatterns: excludePatterns,
	}

	if err := fileBackup.Validate(); err != nil {
		return nil, fmt.Errorf("invalid file backup configuration: %w", err)
	}

	// Generate output filename
	timestamp := time.Now().Format("20060102-150405")
	outputPath := filepath.Join(outputDir, fmt.Sprintf("%s-%s.tar.gz", backupName, timestamp))

	// Ensure output directory exists
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	return fileBackup.Backup(outputPath)
}

func init() {
	rootCmd.AddCommand(backupCmd)

	backupCmd.Flags().StringVar(&backupType, "type", "postgres", "Backup type: postgres, files, mysql")
	backupCmd.Flags().StringVar(&backupName, "name", "", "Backup name (required)")
	backupCmd.MarkFlagRequired("name")

	// PostgreSQL flags
	backupCmd.Flags().StringVar(&dbHost, "db-host", "localhost", "PostgreSQL host")
	backupCmd.Flags().IntVar(&dbPort, "db-port", 5432, "PostgreSQL port")
	backupCmd.Flags().StringVar(&dbUser, "db-user", "postgres", "PostgreSQL user")
	backupCmd.Flags().StringVar(&dbPassword, "db-password", "", "PostgreSQL password")
	backupCmd.Flags().StringVar(&dbName, "db-name", "", "PostgreSQL database name (required for postgres type)")

	// File backup flags
	backupCmd.Flags().StringSliceVar(&backupSources, "source", []string{}, "Source files/directories to backup (can be specified multiple times)")
	backupCmd.Flags().StringSliceVar(&excludePatterns, "exclude", []string{}, "Patterns to exclude (e.g., *.log, tmp/*)")

	backupCmd.Flags().StringVar(&outputDir, "output", "./backups", "Output directory for backups")

	// Encryption flags
	backupCmd.Flags().BoolVar(&encryptBackup, "encrypt", false, "Encrypt backup file")
	backupCmd.Flags().StringVar(&encryptionKey, "encryption-key", "", "Encryption key (or use BACKUP_ENCRYPTION_KEY env var)")
}
