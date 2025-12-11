package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

type PostgresConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

type BackupResult struct {
	FilePath       string
	OriginalSize   int64
	CompressedSize int64
	Duration       time.Duration
}

// DumpPostgres creates a PostgreSQL dump and compresses it to .tar.gz
func DumpPostgres(config PostgresConfig, backupName string, outputDir string) (*BackupResult, error) {
	startTime := time.Now()

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate filenames
	timestamp := time.Now().Format("20060102-150405")
	dumpFileName := fmt.Sprintf("%s-%s.sql", backupName, timestamp)
	dumpFilePath := filepath.Join(outputDir, dumpFileName)
	tarGzFileName := fmt.Sprintf("%s-%s.tar.gz", backupName, timestamp)
	tarGzFilePath := filepath.Join(outputDir, tarGzFileName)

	// Step 1: Run pg_dump
	fmt.Printf("Dumping PostgreSQL database '%s'...\n", config.Database)
	if err := runPgDump(config, dumpFilePath); err != nil {
		return nil, fmt.Errorf("pg_dump failed: %w", err)
	}

	// Get original size
	fileInfo, err := os.Stat(dumpFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat dump file: %w", err)
	}
	originalSize := fileInfo.Size()
	fmt.Printf("Dump created: %s (%.2f MB)\n", dumpFilePath, float64(originalSize)/1024/1024)

	// Step 2: Compress to .tar.gz
	fmt.Printf("Compressing to %s...\n", tarGzFileName)
	if err := compressTarGz(dumpFilePath, tarGzFilePath); err != nil {
		os.Remove(dumpFilePath) // Cleanup
		return nil, fmt.Errorf("compression failed: %w", err)
	}

	// Get compressed size
	compressedInfo, err := os.Stat(tarGzFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat compressed file: %w", err)
	}
	compressedSize := compressedInfo.Size()

	// Remove original dump file
	os.Remove(dumpFilePath)

	duration := time.Since(startTime)
	compressionRatio := float64(compressedSize) / float64(originalSize) * 100

	fmt.Printf("✅ Backup completed successfully!\n")
	fmt.Printf("   Original size: %.2f MB\n", float64(originalSize)/1024/1024)
	fmt.Printf("   Compressed size: %.2f MB (%.1f%% of original)\n",
		float64(compressedSize)/1024/1024, compressionRatio)
	fmt.Printf("   Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Printf("   Output: %s\n", tarGzFilePath)

	return &BackupResult{
		FilePath:       tarGzFilePath,
		OriginalSize:   originalSize,
		CompressedSize: compressedSize,
		Duration:       duration,
	}, nil
}

// runPgDump executes pg_dump command
func runPgDump(config PostgresConfig, outputPath string) error {
	// Set PGPASSWORD environment variable
	env := os.Environ()
	if config.Password != "" {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", config.Password))
	}

	// Build pg_dump command
	args := []string{
		"-h", config.Host,
		"-p", fmt.Sprintf("%d", config.Port),
		"-U", config.User,
		"-d", config.Database,
		"-f", outputPath,
		"--verbose",
		"--format=plain",
	}

	cmd := exec.Command("pg_dump", args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// compressTarGz compresses a file to .tar.gz format
func compressTarGz(inputPath, outputPath string) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Get file info
	fileInfo, err := inputFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat input file: %w", err)
	}

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	// Create tar header
	header := &tar.Header{
		Name:    filepath.Base(inputPath),
		Size:    fileInfo.Size(),
		Mode:    int64(fileInfo.Mode()),
		ModTime: fileInfo.ModTime(),
	}

	// Write header
	if err := tarWriter.WriteHeader(header); err != nil {
		return fmt.Errorf("failed to write tar header: %w", err)
	}

	// Copy file content to tar
	if _, err := io.Copy(tarWriter, inputFile); err != nil {
		return fmt.Errorf("failed to write file to tar: %w", err)
	}

	return nil
}

// RestorePostgres restores a PostgreSQL database from a .tar.gz backup
func RestorePostgres(config PostgresConfig, backupFile string, targetDB string) error {
	fmt.Printf("Starting restore from backup: %s\n", backupFile)

	// Create temporary directory for extraction
	tempDir, err := os.MkdirTemp("", "pg-restore-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Step 1: Extract .tar.gz
	fmt.Printf("Extracting backup file...\n")
	sqlFile, err := extractTarGz(backupFile, tempDir)
	if err != nil {
		return fmt.Errorf("extraction failed: %w", err)
	}
	fmt.Printf("Extracted: %s\n", sqlFile)

	// Override target database if specified
	if targetDB != "" {
		config.Database = targetDB
	}

	// Step 2: Restore to PostgreSQL
	fmt.Printf("Restoring to database '%s'...\n", config.Database)
	if err := runPsqlRestore(config, sqlFile); err != nil {
		return fmt.Errorf("restore failed: %w", err)
	}

	fmt.Printf("✅ Restore completed successfully!\n")
	fmt.Printf("   Database: %s\n", config.Database)
	fmt.Printf("   From: %s\n", backupFile)

	return nil
}

// extractTarGz extracts a .tar.gz file and returns the path to the extracted SQL file
func extractTarGz(tarGzPath, destDir string) (string, error) {
	// Open the tar.gz file
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", fmt.Errorf("failed to open tar.gz file: %w", err)
	}
	defer file.Close()

	// Create gzip reader
	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return "", fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzipReader.Close()

	// Create tar reader
	tarReader := tar.NewReader(gzipReader)

	var extractedFile string

	// Extract files
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", fmt.Errorf("failed to read tar header: %w", err)
		}

		// Construct output path
		targetPath := filepath.Join(destDir, filepath.Base(header.Name))

		// Only extract regular files (skip directories)
		if header.Typeflag == tar.TypeReg {
			outFile, err := os.Create(targetPath)
			if err != nil {
				return "", fmt.Errorf("failed to create output file: %w", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return "", fmt.Errorf("failed to extract file: %w", err)
			}
			outFile.Close()

			extractedFile = targetPath
			fmt.Printf("  Extracted: %s (%.2f MB)\n", filepath.Base(targetPath), float64(header.Size)/1024/1024)
		}
	}

	if extractedFile == "" {
		return "", fmt.Errorf("no files found in archive")
	}

	return extractedFile, nil
}

// runPsqlRestore executes psql command to restore database
func runPsqlRestore(config PostgresConfig, sqlFilePath string) error {
	// Set PGPASSWORD environment variable
	env := os.Environ()
	if config.Password != "" {
		env = append(env, fmt.Sprintf("PGPASSWORD=%s", config.Password))
	}

	// Build psql command
	args := []string{
		"-h", config.Host,
		"-p", fmt.Sprintf("%d", config.Port),
		"-U", config.User,
		"-d", config.Database,
		"-f", sqlFilePath,
		"--echo-errors",
	}

	cmd := exec.Command("psql", args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
