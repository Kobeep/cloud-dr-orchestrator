package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Kobeep/cloud-dr-orchestrator/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

var metricsCmd = &cobra.Command{
	Use:   "metrics",
	Short: "Start metrics server for Prometheus/Grafana monitoring",
	Long: `Start an HTTP server that exposes metrics for Prometheus scraping.
The server provides two endpoints:
  - /metrics: Prometheus metrics endpoint
  - /health: Health check endpoint (JSON)

Metrics include:
  - Backup operation duration, size, success/failure counts
  - Upload/download operation statistics
  - Restore operation statistics

Example:
  # Start metrics server on default port 9090
  orchestrator metrics

  # Start on custom port and address
  orchestrator metrics --port 8080 --address 0.0.0.0
`,
	RunE: runMetrics,
}

var (
	metricsPort    int
	metricsAddress string
)

func init() {
	rootCmd.AddCommand(metricsCmd)

	metricsCmd.Flags().IntVar(&metricsPort, "port", 9090, "Port to listen on")
	metricsCmd.Flags().StringVar(&metricsAddress, "address", "0.0.0.0", "Address to bind to")
}

func runMetrics(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", metricsAddress, metricsPort)

	// Create HTTP server
	mux := http.NewServeMux()

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	// Health check endpoint
	mux.HandleFunc("/health", handleHealth)

	// Root endpoint with info
	mux.HandleFunc("/", handleRoot)

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("🚀 Starting metrics server...\n")
	fmt.Printf("   Address: %s\n", addr)
	fmt.Printf("   Metrics endpoint: http://%s/metrics\n", addr)
	fmt.Printf("   Health endpoint:  http://%s/health\n\n", addr)
	fmt.Printf("📊 Ready for Prometheus scraping!\n")

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

// handleHealth returns health status as JSON
func handleHealth(w http.ResponseWriter, r *http.Request) {
	health := metrics.GetHealth()

	status := "healthy"
	httpStatus := http.StatusOK

	// Consider unhealthy if last backup failed or no backup in 25 hours
	if !health.IsHealthy {
		status = "unhealthy"
		httpStatus = http.StatusServiceUnavailable
	} else if !health.LastBackupTime.IsZero() {
		timeSinceBackup := time.Since(health.LastBackupTime)
		if timeSinceBackup > 25*time.Hour {
			status = "degraded"
			httpStatus = http.StatusOK // Still 200, but flagged
		}
	}

	response := map[string]interface{}{
		"status":            status,
		"last_backup_time":  health.LastBackupTime.Format(time.RFC3339),
		"last_backup_error": health.LastBackupError,
		"backup_count":      health.BackupCount,
		"timestamp":         time.Now().Format(time.RFC3339),
	}

	// If never backed up
	if health.LastBackupTime.IsZero() {
		response["last_backup_time"] = nil
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	json.NewEncoder(w).Encode(response)
}

// handleRoot returns basic server info
func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Cloud DR Orchestrator - Metrics Server</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            max-width: 800px;
            margin: 40px auto;
            padding: 20px;
            line-height: 1.6;
        }
        h1 { color: #333; }
        .endpoint {
            background: #f5f5f5;
            padding: 15px;
            margin: 10px 0;
            border-radius: 5px;
        }
        .endpoint a {
            color: #0066cc;
            text-decoration: none;
            font-weight: bold;
        }
        .endpoint a:hover { text-decoration: underline; }
        code {
            background: #eee;
            padding: 2px 6px;
            border-radius: 3px;
            font-family: monospace;
        }
    </style>
</head>
<body>
    <h1>🎯 Cloud DR Orchestrator - Metrics Server</h1>
    <p>This server exposes metrics for Prometheus/Grafana monitoring.</p>

    <div class="endpoint">
        <h3>📊 <a href="/metrics">Prometheus Metrics</a></h3>
        <p>Metrics endpoint for Prometheus scraping. Configure your <code>prometheus.yml</code>:</p>
        <pre>scrape_configs:
  - job_name: 'orchestrator'
    static_configs:
      - targets: ['localhost:9090']</pre>
    </div>

    <div class="endpoint">
        <h3>💚 <a href="/health">Health Check</a></h3>
        <p>JSON health status including last backup time and error information.</p>
    </div>

    <hr>
    <p><em>Cloud DR Orchestrator - Multi-cloud Disaster Recovery</em></p>
</body>
</html>`)
}
