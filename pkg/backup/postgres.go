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
