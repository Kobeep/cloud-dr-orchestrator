package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// BackupDuration tracks how long backup operations take (in seconds)
	BackupDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "orchestrator_backup_duration_seconds",
		Help:    "Duration of backup operations in seconds",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800}, // 1s to 30min
	})

	// BackupSize tracks the size of backups (in bytes)
	BackupSize = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "orchestrator_backup_size_bytes",
		Help:    "Size of backup files in bytes",
		Buckets: prometheus.ExponentialBuckets(1024, 2, 20), // 1KB to ~1GB
	})

	// BackupSuccess counts successful backup operations
	BackupSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orchestrator_backup_success_total",
		Help: "Total number of successful backup operations",
	})

	// BackupFailure counts failed backup operations
	BackupFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orchestrator_backup_failure_total",
		Help: "Total number of failed backup operations",
	}, []string{"reason"})

	// UploadDuration tracks upload operation duration (in seconds)
	UploadDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "orchestrator_upload_duration_seconds",
		Help:    "Duration of upload operations in seconds",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
	})

	// UploadSuccess counts successful upload operations
	UploadSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orchestrator_upload_success_total",
		Help: "Total number of successful upload operations",
	})

	// UploadFailure counts failed upload operations
	UploadFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orchestrator_upload_failure_total",
		Help: "Total number of failed upload operations",
	}, []string{"reason"})

	// DownloadDuration tracks download operation duration (in seconds)
	DownloadDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "orchestrator_download_duration_seconds",
		Help:    "Duration of download operations in seconds",
		Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30, 60, 120, 300},
	})

	// DownloadSuccess counts successful download operations
	DownloadSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orchestrator_download_success_total",
		Help: "Total number of successful download operations",
	})

	// DownloadFailure counts failed download operations
	DownloadFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orchestrator_download_failure_total",
		Help: "Total number of failed download operations",
	}, []string{"reason"})

	// RestoreDuration tracks restore operation duration (in seconds)
	RestoreDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "orchestrator_restore_duration_seconds",
		Help:    "Duration of restore operations in seconds",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300, 600, 1800, 3600},
	})

	// RestoreSuccess counts successful restore operations
	RestoreSuccess = promauto.NewCounter(prometheus.CounterOpts{
		Name: "orchestrator_restore_success_total",
		Help: "Total number of successful restore operations",
	})

	// RestoreFailure counts failed restore operations
	RestoreFailure = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "orchestrator_restore_failure_total",
		Help: "Total number of failed restore operations",
	}, []string{"reason"})
)

// HealthStatus stores the last operation status for health checks
type HealthStatus struct {
	mu              sync.RWMutex
	LastBackupTime  time.Time
	LastBackupError string
	BackupCount     int64
	IsHealthy       bool
}

var globalHealth = &HealthStatus{
	IsHealthy: true,
}

// GetHealth returns the current health status
func GetHealth() *HealthStatus {
	globalHealth.mu.RLock()
	defer globalHealth.mu.RUnlock()

	return &HealthStatus{
		LastBackupTime:  globalHealth.LastBackupTime,
		LastBackupError: globalHealth.LastBackupError,
		BackupCount:     globalHealth.BackupCount,
		IsHealthy:       globalHealth.IsHealthy,
	}
}

// RecordBackupSuccess updates health status after successful backup
func RecordBackupSuccess() {
	globalHealth.mu.Lock()
	defer globalHealth.mu.Unlock()

	globalHealth.LastBackupTime = time.Now()
	globalHealth.LastBackupError = ""
	globalHealth.BackupCount++
	globalHealth.IsHealthy = true
}

// RecordBackupError updates health status after backup failure
func RecordBackupError(err error) {
	globalHealth.mu.Lock()
	defer globalHealth.mu.Unlock()

	globalHealth.LastBackupTime = time.Now()
	globalHealth.LastBackupError = err.Error()
	globalHealth.IsHealthy = false
}

// ResetHealth resets health status (useful for testing)
func ResetHealth() {
	globalHealth.mu.Lock()
	defer globalHealth.mu.Unlock()

	globalHealth.LastBackupTime = time.Time{}
	globalHealth.LastBackupError = ""
	globalHealth.BackupCount = 0
	globalHealth.IsHealthy = true
}
