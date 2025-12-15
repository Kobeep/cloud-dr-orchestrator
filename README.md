# Cloud DR Orchestrator

Multi-cloud Disaster Recovery Orchestrator with Oracle Cloud Free Tier

## Overview

Cloud DR Orchestrator is a tool for automated PostgreSQL database backups with Oracle Cloud Object Storage integration. Built with Go, it provides a simple CLI for creating compressed backups and managing them in the cloud.

## Features

✅ **PostgreSQL Backup** - Automated database dumps with compression
✅ **Oracle Cloud Integration** - Upload, download, and list backups in OCI Object Storage
✅ **Compression** - tar.gz compression to save space
✅ **Organization** - Automatic date-based folder structure (backups/YYYY/MM/)
✅ **Free Tier** - Uses Oracle Cloud Free Tier (20GB storage, 50k API calls/month)

## Quick Start

### Prerequisites

- Go 1.21+
- PostgreSQL client tools (`pg_dump`)
- Oracle Cloud account with Object Storage bucket
- OCI credentials configured in `~/.oci/config`

### Installation

```bash
# Build the binary
go build -o bin/orchestrator ./cmd/orchestrator

# Or install directly
go install github.com/Kobeep/cloud-dr-orchestrator/cmd/orchestrator@latest
```

### Usage

**1. Create a local backup:**
```bash
orchestrator backup \
  --name my-backup \
  --db-name mydb \
  --db-host localhost \
  --db-user postgres \
  --db-password secret
```

**2. Upload backup to Oracle Cloud:**
```bash
orchestrator upload \
  --file backup-20251209-092658.tar.gz \
  --bucket cloud-dr-orchestrator-dr-backups \
  --compartment ocid1.compartment.oc1..xxx
```

**3. List backups in cloud:**
```bash
orchestrator list \
  --bucket cloud-dr-orchestrator-dr-backups \
  --compartment ocid1.compartment.oc1..xxx \
  --year 2025 --month 12
```

**4. Download backup from cloud:**
```bash
orchestrator download \
  --object backups/2025/12/backup-20251209-092658.tar.gz \
  --output ./restored-backup.tar.gz \
  --bucket cloud-dr-orchestrator-dr-backups \
  --compartment ocid1.compartment.oc1..xxx
```

**5. Restore from local backup:**
```bash
orchestrator restore \
  --file backup-20251209-092658.tar.gz \
  --db-name mydb \
  --db-host localhost \
  --db-user postgres \
  --db-password secret
```

**6. Restore directly from cloud:**
```bash
orchestrator restore \
  --from-cloud backups/2025/12/backup-20251209-092658.tar.gz \
  --bucket cloud-dr-orchestrator-dr-backups \
  --compartment ocid1.compartment.oc1..xxx \
  --db-name mydb \
  --db-host localhost \
  --db-user postgres \
  --db-password secret
```

## Configuration

### Oracle Cloud Credentials

The tool reads credentials from `~/.oci/config`:

```ini
[DEFAULT]
user=ocid1.user.oc1..xxx
fingerprint=aa:bb:cc:dd:ee:ff
tenancy=ocid1.tenancy.oc1..xxx
region=eu-frankfurt-1
key_file=/home/user/.oci/oci_api_key.pem
```

You can override the config file path and profile:
```bash
orchestrator upload --oci-config /path/to/config --oci-profile MYPROFILE ...
```

## Complete Workflow Example

Here's a complete disaster recovery workflow:

```bash
# 1. Create a PostgreSQL backup (compressed)
orchestrator backup \
  --name production-db \
  --db-name myapp \
  --db-host localhost \
  --db-user postgres \
  --db-password secretpass
# Output: backup-20251209-104235.tar.gz

# 2. Upload to Oracle Cloud (automatically organized in backups/2025/12/)
orchestrator upload \
  --file backup-20251209-104235.tar.gz \
  --bucket cloud-dr-orchestrator-dr-backups \
  --compartment ocid1.tenancy.oc1..xxx
# ✓ Upload successful in 90ms

# 3. List all backups from December 2025
orchestrator list \
  --bucket cloud-dr-orchestrator-dr-backups \
  --compartment ocid1.tenancy.oc1..xxx \
  --year 2025 --month 12
# Found 1 backup(s)

# 4. Download when disaster strikes
orchestrator download \
  --object backups/2025/12/backup-20251209-104235.tar.gz \
  --output ./restore.tar.gz \
  --bucket cloud-dr-orchestrator-dr-backups \
  --compartment ocid1.tenancy.oc1..xxx
# ✓ Download successful in 81ms

# 5. Extract and restore
tar xzf restore.tar.gz
psql myapp < backup-20251209-104235.sql
```

## Automated Scheduling

Set up automated backups with systemd (Linux) or cron:

### systemd (Recommended for Linux)

```bash
# Run the setup script
sudo ./scripts/setup-automation.sh

# Edit configuration
sudo nano /etc/cloud-dr-orchestrator/backup.env

# Start the timer
sudo systemctl start orchestrator-backup.timer
sudo systemctl enable orchestrator-backup.timer

# Check status
systemctl status orchestrator-backup.timer
journalctl -u orchestrator-backup.service
```

The timer runs daily at 2 AM by default. Edit `/etc/systemd/system/orchestrator-backup.timer` to customize.

### cron (Linux/macOS)

```bash
# Run setup script (creates wrapper)
sudo ./scripts/setup-automation.sh

# Edit configuration
sudo nano /etc/cloud-dr-orchestrator/backup.env

# Add to crontab
crontab -e

# Daily at 2 AM
0 2 * * * /usr/local/bin/orchestrator-backup-cron.sh

# Every 6 hours
0 */6 * * * /usr/local/bin/orchestrator-backup-cron.sh

# Weekly on Sunday at 3 AM
0 3 * * 0 /usr/local/bin/orchestrator-backup-cron.sh
```

**Free Tier Note:** With 20GB storage and automated daily backups, keep ~30 days of backups before rotation needed.

## Monitoring and Observability

Monitor your backups with **Grafana Alloy** and **Grafana Cloud** (free tier)! 📊

### Quick Start

```bash
# Start metrics server
orchestrator metrics --port 9090

# Available endpoints:
# - http://localhost:9090/metrics  (Prometheus metrics)
# - http://localhost:9090/health   (Health check JSON)
```

### Metrics Available

The orchestrator exposes comprehensive metrics:

**Backup Metrics:**
- `orchestrator_backup_duration_seconds` - Backup operation time (histogram)
- `orchestrator_backup_size_bytes` - Backup file size (histogram)
- `orchestrator_backup_success_total` - Successful backup counter
- `orchestrator_backup_failure_total` - Failed backup counter (with reason labels)

**Cloud Operations:**
- `orchestrator_upload_duration_seconds` - Upload time to Oracle Cloud
- `orchestrator_upload_success_total` / `_failure_total` - Upload counters
- `orchestrator_download_duration_seconds` - Download time from Oracle Cloud
- `orchestrator_download_success_total` / `_failure_total` - Download counters

**Restore Operations:**
- `orchestrator_restore_duration_seconds` - Restore operation time
- `orchestrator_restore_success_total` / `_failure_total` - Restore counters

### Grafana Cloud Setup (FREE)

1. **Sign up:** https://grafana.com/auth/sign-up/create-user
   - Free tier: 10k metrics, 50GB logs, 14 days retention
   - Our app uses ~50 metrics series (well within limit!)

2. **Install Grafana Alloy:**
   ```bash
   # Linux (Debian/Ubuntu)
   wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor | sudo tee /usr/share/keyrings/grafana.gpg
   echo "deb [signed-by=/usr/share/keyrings/grafana.gpg] https://apt.grafana.com stable main" | sudo tee /etc/apt/sources.list.d/grafana.list
   sudo apt update && sudo apt install alloy

   # macOS
   brew install grafana/grafana/alloy
   ```

3. **Configure:**
   ```bash
   # Copy config templates
   cp configs/grafana-cloud.env.example configs/grafana-cloud.env
   cp configs/alloy-config.alloy /etc/alloy/config.alloy

   # Edit with your Grafana Cloud credentials
   nano configs/grafana-cloud.env
   ```

4. **Start Alloy:**
   ```bash
   export $(cat configs/grafana-cloud.env | xargs)
   sudo alloy run configs/alloy-config.alloy
   ```

5. **Import Dashboard:**
   - Go to Grafana Cloud → Dashboards → Import
   - Upload `configs/grafana-dashboard.json`
   - Done! 🎉

### Dashboard Panels

Your dashboard includes:
- ✅ Overall health status indicator
- 📊 Backup success rate (24h)
- ⏱️ Backup duration trends (p95, median)
- 📦 Backup size distribution over time
- ☁️ Upload/download performance
- 🔴 Recent failures table with error reasons

### Alerting

Automatic alerts for:
- **BackupFailed** - Immediate alert on backup failure
- **BackupNotRunRecently** - No backup in 25 hours
- **BackupTakingTooLong** - Duration > 30 minutes
- **UploadFailed** - Cloud upload errors
- **OrchestratorDown** - Service unavailable

Configure notifications via:
- Email
- Slack
- PagerDuty
- Webhook

📖 **Full setup guide:** [docs/MONITORING.md](docs/MONITORING.md)

## Development Status

� **Active Development** - Core features implemented and tested!

### Completed Features
- ✅ **Issue #2**: Terraform infrastructure for Oracle Cloud Object Storage
  - Bucket with versioning enabled
  - Free Tier configuration (20GB storage)

- ✅ **Issue #3**: PostgreSQL backup with compression
  - `pg_dump` integration
  - tar.gz compression (40-50% size reduction)
  - Backup metadata and statistics

- ✅ **Issue #4**: Oracle Cloud Object Storage integration
  - Upload backups to OCI (90ms average)
  - Download backups from OCI (81ms average)
  - List backups with filtering by date
  - Automatic folder organization (backups/YYYY/MM/)
  - Tested with real OCI bucket ✓

- ✅ **Issue #7**: Restore functionality
  - Restore from local .tar.gz backups
  - Download from cloud and restore
  - Confirmation prompt before restore
  - Support for target database override

- ✅ **Issue #8**: Scheduling and automation
  - systemd service and timer
  - Cron wrapper script
  - Automated installation script
  - Works on Linux and macOS

- ✅ **Issue #9**: Monitoring and observability
  - Prometheus metrics endpoint
  - Grafana Alloy configuration
  - Grafana Cloud integration (free tier)
  - Pre-built dashboard with 11 panels
  - Comprehensive alerting rules
  - Health check endpoint

## Contributing

This is a personal learning project, but suggestions and feedback are welcome! Open an issue or submit a PR.

## License

MIT License - see LICENSE file for details
