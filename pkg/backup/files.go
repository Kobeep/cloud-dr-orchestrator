package backup

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/schollz/progressbar/v3"
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

	// First pass: count total files to backup
	fmt.Println("📊 Scanning files...")
	totalFilesToBackup, err := fb.countFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to count files: %w", err)
	}
	fmt.Printf("Found %d files to backup\n\n", totalFilesToBackup)

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

	// Create progress bar
	bar := progressbar.NewOptions64(
		totalFilesToBackup,
		progressbar.OptionSetDescription("📦 Backing up files"),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)

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

				written, err := io.Copy(tarWriter, file)
				file.Close() // Close immediately after copying
				if err != nil {
					return fmt.Errorf("failed to copy file contents: %w", err)
				}

				totalFiles++
				totalSize += written
				bar.Add(1)
			}

			return nil
		})

		if err != nil {
			return nil, fmt.Errorf("failed to walk source %s: %w", source, err)
		}
	}

	bar.Finish()
	fmt.Println() // Add newline after progress bar

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

// countFiles counts total number of files to backup (for progress bar)
func (fb *FileBackup) countFiles() (int64, error) {
	var count int64
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

			// Count only files, not directories
			if !info.IsDir() {
				count++
			}
			return nil
		})
		if err != nil {
			return 0, err
		}
	}
	return count, nil
}

// shouldExclude checks if a path matches any exclude pattern
func (fb *FileBackup) shouldExclude(path string) bool {
	for _, pattern := range fb.ExcludePatterns {
		// Try matching full path first
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}
		// Also try matching just the base name for simple patterns like "*.log"
		matched, err = filepath.Match(pattern, filepath.Base(path))
		if err == nil && matched {
			return true
		}
	}
	return false
}
