package backup

import "time"

// BackupType defines the type of backup to perform
type BackupType string

const (
	TypePostgreSQL BackupType = "postgres"
	TypeMySQL      BackupType = "mysql"
	TypeFiles      BackupType = "files"
	TypeDirectory  BackupType = "directory"
)

// Backuper is the common interface for all backup types
type Backuper interface {
	Backup(outputPath string) (*Result, error)
	Validate() error
}

// Result contains information about a completed backup
type Result struct {
	Type           BackupType
	Filename       string
	Path           string
	Size           int64
	OriginalSize   int64
	Duration       time.Duration
	FilesIncluded  int64
	DatabaseName   string // For database backups
	Timestamp      time.Time
	CompressionPct float64
}

// CalculateCompressionPct calculates compression percentage
func (r *Result) CalculateCompressionPct() float64 {
	if r.OriginalSize == 0 {
		return 0
	}
	return (1.0 - float64(r.Size)/float64(r.OriginalSize)) * 100
}
