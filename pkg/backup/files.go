package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// FileBackup handles generic file and directory backups
type FileBackup struct {
	Name            string
	Sources         []string // List of files/directories to backup
	ExcludePatterns []string // Patterns to exclude (e.g., "*.log", "tmp/*")
}

// Validate checks if the configuration is valid
func (fb *FileBackup) Validate() error {
	if len(fb.Sources) == 0 {
		return fmt.Errorf("no sources specified for backup")
	}

	for _, source := range fb.Sources {
		if _, err := os.Stat(source); os.IsNotExist(err) {
			return fmt.Errorf("source does not exist: %s", source)
		}
	}

	return nil
}

// Backup creates a tar.gz archive of specified files/directories
func (fb *FileBackup) Backup(outputPath string) (*Result, error) {
	startTime := time.Now()

	// Create output file
	outFile, err := os.Create(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create gzip writer
	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	// Create tar writer
	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	var totalFiles int64
	var totalSize int64

	// Add each source to archive
	for _, source := range fb.Sources {
		err := filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// Check if should be excluded
			if fb.shouldExclude(path) {
				if info.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			// Create tar header
			header, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return fmt.Errorf("failed to create tar header: %w", err)
			}

			// Use relative path in archive
			header.Name = path

			// Write header
			if err := tarWriter.WriteHeader(header); err != nil {
				return fmt.Errorf("failed to write tar header: %w", err)
			}

			// If it's a file, copy contents
			if !info.IsDir() {
				file, err := os.Open(path)
				if err != nil {
					return fmt.Errorf("failed to open file: %w", err)
				}
				defer file.Close()

				written, err := io.Copy(tarWriter, file)
				if err != nil {
					return fmt.Errorf("failed to copy file contents: %w", err)
				}

				totalFiles++
				totalSize += written
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk source %s: %w", source, err)
		}
	}

	duration := time.Since(startTime)

	// Get output file size
	fileInfo, err := os.Stat(outputPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat output file: %w", err)
	}

	compressionPct := 0.0
	if totalSize > 0 {
		compressionPct = (1.0 - float64(fileInfo.Size())/float64(totalSize)) * 100
	}

	return &Result{
		Type:           TypeFiles,
		Filename:       filepath.Base(outputPath),
		Path:           outputPath,
		Size:           fileInfo.Size(),
		OriginalSize:   totalSize,
		Duration:       duration,
		FilesIncluded:  totalFiles,
		Timestamp:      startTime,
		CompressionPct: compressionPct,
	}, nil
}

// shouldExclude checks if a path matches any exclude pattern
func (fb *FileBackup) shouldExclude(path string) bool {
	for _, pattern := range fb.ExcludePatterns {
		matched, err := filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}
