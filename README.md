<div align="center">

# 🌩️ Cloud DR Orchestrator

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Oracle Cloud](https://img.shields.io/badge/Oracle%20Cloud-Free%20Tier-F80000?style=flat&logo=oracle)](https://www.oracle.com/cloud/free/)
[![Terraform](https://img.shields.io/badge/Terraform-1.0+-7B42BC?style=flat&logo=terraform)](https://www.terraform.io/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

**🚀 Production-Ready Multi-Purpose Disaster Recovery with Oracle Cloud Free Tier**

*Automated backups • Databases & Files • Encryption • Monitoring • Zero-cost infrastructure*

[Features](#-features) • [Quick Start](#-quick-start) • [Documentation](#-documentation) • [Contributing](#-contributing)

---

<img src="https://raw.githubusercontent.com/Kobeep/cloud-dr-orchestrator/main/docs/images/architecture.svg" alt="Architecture" width="800">

<sub>Built with ❤️ using Go, Terraform, Oracle Cloud, and Prometheus</sub>

</div>

---

## 📖 Overview

**Cloud DR Orchestrator** is a production-grade disaster recovery solution for **databases and file systems**, leveraging Oracle Cloud's generous **Free Tier** (20GB storage, 50k API calls/month). Perfect for backing up:

- 🗄️ **PostgreSQL databases** - Full dumps with compression
- 📁 **Application configs** - Nginx, Apache, app settings
- 🔧 **System files** - SSL certificates, SSH keys, scripts
- 📂 **User data** - Documents, logs, any files/directories
- 🔜 **MySQL databases** (coming soon)

### Why Cloud DR Orchestrator?

- 💰 **$0/month** - Runs entirely on Oracle Cloud Free Tier
- 🎯 **Multi-purpose** - Databases, files, configs - all in one tool
- 🔐 **AES-256-GCM** encryption for data at rest
- 📊 **Prometheus metrics** + Grafana dashboards included
- 🤖 **Automated scheduling** with Cronify integration
- 🏗️ **Infrastructure as Code** - Full Terraform setup
- 🎯 **Production tested** - Used in real-world applications

---

## ✨ Features

<table>
<tr>
<td width="50%">

### 🗄️ Database Backup
- PostgreSQL automated dumps
- MySQL support (coming soon)
- Custom naming schemes
- Date-based organization
- Multiple database support

</td>
<td width="50%">

### 🔐 Security & Encryption
- AES-256-GCM encryption
- PBKDF2 key derivation
- 100k iterations
- Per-file salt & nonce
- Environment variable keys

</td>
</tr>
<tr>
<td width="50%">

### ☁️ Cloud Integration
- Oracle Cloud Object Storage
- 20GB Free Tier storage
- 50k API calls/month
- Automatic folder structure
- Upload/download/list operations

</td>
<td width="50%">

### 📊 Monitoring & Observability
- Prometheus metrics
- Grafana dashboards
- Alloy integration
- Backup/restore duration
- Success/failure tracking

</td>
</tr>
<tr>
<td width="50%">

### 📁 File & Directory Backup
- Generic file backup
- Directory recursion
- Exclude patterns (*.log, tmp/*)
- Config file backups
- SSL certificates, SSH keys
- Application data

</td>
<td width="50%">

### 🤖 Automation
- Cronify integration
- YAML-based schedules
- Daily/weekly/monthly
- Conflict detection
- Dry-run mode

</td>
</tr>
<tr>
<td width="50%">

### 🏗️ Infrastructure as Code
- Full Terraform setup
- Oracle Cloud provider
- Object Storage bucket
- IAM policies
- One-command deploy

</td>
<td width="50%">

<!-- Pusta komórka dla zachowania struktury 2-kolumnowej -->

</td>
</tr>
</table>

---

---

## 🚀 Quick Start

### 📋 Prerequisites

Before you begin, ensure you have:

- ✅ Go 1.21 or higher
- ✅ PostgreSQL with `pg_dump` and `pg_restore`
- ✅ [Oracle Cloud Free Tier account](https://www.oracle.com/cloud/free/)
- ✅ OCI credentials in `~/.oci/config`
- ✅ (Optional) [Cronify](https://github.com/Kobeep/Cronify) for automated schedules

### 📦 Installation

**Option 1: Build from source**

```bash
git clone https://github.com/Kobeep/cloud-dr-orchestrator.git
cd cloud-dr-orchestrator
go build -o orchestrator ./cmd/orchestrator
sudo mv orchestrator /usr/local/bin/
```

**Option 2: Go install**

```bash
go install github.com/Kobeep/cloud-dr-orchestrator/cmd/orchestrator@latest
```

**Option 3: Download binary** (coming soon)

```bash
# Linux
wget https://github.com/Kobeep/cloud-dr-orchestrator/releases/latest/download/orchestrator-linux-amd64
chmod +x orchestrator-linux-amd64
sudo mv orchestrator-linux-amd64 /usr/local/bin/orchestrator
```

### 🏗️ Infrastructure Setup

Deploy Oracle Cloud infrastructure with Terraform:

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

This creates:
- Object Storage bucket (20GB Free Tier)
- IAM policies
- Compartment structure
- API keys

### 🎯 Basic Usage

**PostgreSQL Database Backup:**

```bash
# Create database backup
orchestrator backup \
  --type postgres \
  --name prod-db \
  --db-name myapp \
  --db-host localhost \
  --db-user postgres \
  --db-password secret

# With encryption
orchestrator backup \
  --type postgres \
  --name prod-db \
  --db-name myapp \
  --encrypt
```

**File & Directory Backup:**

```bash
# Backup nginx configs
orchestrator backup \
  --type files \
  --name nginx-configs \
  --source /etc/nginx \
  --source /etc/ssl/nginx

# Backup with exclusions
orchestrator backup \
  --type files \
  --name app-data \
  --source /var/www/myapp \
  --exclude "*.log" \
  --exclude "tmp/*" \
  --exclude "cache/*"

# Backup multiple directories
orchestrator backup \
  --type files \
  --name system-configs \
  --source /etc/systemd \
  --source /etc/cron.d \
  --source ~/.ssh \
  --encrypt
```

**Upload to Oracle Cloud:**

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

## Encryption

Encrypt your backups before uploading to Oracle Cloud for maximum security! 🔐

### Generate Encryption Key

```bash
# Generate a secure 256-bit key
orchestrator keygen

# Output:
# 🔑 Generated 256-bit encryption key:
# s0m3R4nd0mB4s364Enc0d3dK3y==
```

### Store Key Securely

```bash
# Save to environment variable
export BACKUP_ENCRYPTION_KEY="your-key-here"

# Or save to config file
echo "BACKUP_ENCRYPTION_KEY=your-key-here" > ~/.backup-encryption.env
export $(cat ~/.backup-encryption.env | xargs)
```

### Create Encrypted Backup

```bash
# Backup with encryption
orchestrator backup \
  --name prod-db \
  --db-name myapp \
  --db-host localhost \
  --db-user postgres \
  --db-password secret \
  --encrypt

# Output: backup-20251209-104235.tar.gz.encrypted
```

### Restore Encrypted Backup

```bash
# Automatic decryption (detects .encrypted extension)
orchestrator restore \
  --file backup-20251209-104235.tar.gz.encrypted \
  --db-name myapp \
  --db-host localhost \
  --db-user postgres \
  --db-password secret

# Manual decryption (if needed)
orchestrator restore \
  --file backup.tar.gz \
  --decrypt \
  --decryption-key "$BACKUP_ENCRYPTION_KEY" \
  --db-name myapp ...
```

### Encryption Details

- **Algorithm:** AES-256-GCM (industry standard)
- **Key Derivation:** PBKDF2 with 100,000 iterations
- **Security:** Each file has unique salt and nonce
- **Authentication:** GCM mode provides built-in integrity check
- **File Format:** `.tar.gz.encrypted`

⚠️ **IMPORTANT:**

- **Never lose your encryption key!** Lost key = lost backups
- **Store keys securely** (use secret managers in production)
- **Never commit keys** to version control
- **Backup your keys** in a secure location
- **Rotate keys** periodically

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

## Automated Backup Schedules

Schedule automated backups using **Cronify** integration! 🕒

### Install Cronify

Install Cronify (YAML-driven cron job manager):

```bash
git clone https://github.com/Kobeep/Cronify.git
cd Cronify
sudo ./install.sh
```

### Generate Schedule Template

```bash
# Create example backup-schedule.yaml
orchestrator schedule init

# Or specify custom output file
orchestrator schedule init --output my-schedule.yaml
```

### Example Schedule YAML

```yaml
jobs:
  - name: daily-backup
    schedule: "0 0 * * *"  # Every day at midnight
    command: /usr/local/bin/orchestrator backup --name prod-db --db-name myapp --encrypt
    env:
      BACKUP_ENCRYPTION_KEY: "your-key-here"
      PATH: "/usr/local/bin:/usr/bin:/bin"

  - name: weekly-backup
    schedule: "0 3 * * 0"  # Every Sunday at 3 AM
    command: /usr/local/bin/orchestrator backup --name prod-db-weekly --encrypt

  - name: monthly-backup
    schedule: "0 2 1 * *"  # 1st of month at 2 AM
    command: /usr/local/bin/orchestrator backup --name prod-db-monthly --encrypt
```

### Validate Schedule

```bash
# Check cron expressions and commands
orchestrator schedule validate --file backup-schedule.yaml

# Simulate next 5 run times
orchestrator schedule validate --file backup-schedule.yaml --simulate
```

### Deploy to Crontab

```bash
# Preview without deploying
orchestrator schedule deploy --file backup-schedule.yaml --dry-run

# Deploy to crontab
orchestrator schedule deploy --file backup-schedule.yaml
```

### View Active Schedules

```bash
# View current crontab
crontab -l
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

- ✅ **Issue #10**: Backup encryption
  - AES-256-GCM encryption
  - PBKDF2 key derivation (100k iterations)
  - Key generation command
  - Environment variable support

- ✅ **Issue #13**: Cronify integration
  - YAML-based schedule management
  - Automated cron deployment
  - Schedule validation and simulation

---

## 🤝 Contributing

We welcome contributions! Here's how you can help:

<details>
<summary>💡 Ways to Contribute</summary>

- 🐛 **Report bugs** - Open an issue with reproduction steps
- ✨ **Suggest features** - Share your ideas for improvements
- 📝 **Improve docs** - Help make documentation clearer
- 🔧 **Submit PRs** - Fix bugs or implement features
- ⭐ **Star the repo** - Show your support!

</details>

<details>
<summary>🔧 Development Setup</summary>

```bash
# Clone repository
git clone https://github.com/Kobeep/cloud-dr-orchestrator.git
cd cloud-dr-orchestrator

# Install dependencies
go mod download

# Run tests
go test ./...

# Build
go build -o orchestrator ./cmd/orchestrator

# Run linter
golangci-lint run
```

</details>

### 📜 Code of Conduct

Be respectful, inclusive, and collaborative. See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) for details.

---

## 📄 License

This project is licensed under the **Apache License 2.0** - see the [LICENSE](LICENSE) file for details.

```
Copyright 2024-2025 Jakub Pospieszny

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
```

---

## 🙏 Acknowledgments

<table>
<tr>
<td align="center" width="25%">
<img src="https://www.oracle.com/a/ocom/img/rh02-free-tier.png" width="100"><br>
<b>Oracle Cloud</b><br>
<sub>Free Tier Infrastructure</sub>
</td>
<td align="center" width="25%">
<img src="https://www.terraform.io/assets/images/og-image-8b3e4f7d.png" width="100"><br>
<b>Terraform</b><br>
<sub>Infrastructure as Code</sub>
</td>
<td align="center" width="25%">
<img src="https://prometheus.io/assets/prometheus_logo_grey.svg" width="100"><br>
<b>Prometheus</b><br>
<sub>Metrics & Monitoring</sub>
</td>
<td align="center" width="25%">
<img src="https://grafana.com/static/img/menu/grafana2.svg" width="100"><br>
<b>Grafana</b><br>
<sub>Dashboards & Alerts</sub>
</td>
</tr>
</table>

### Special Thanks

- **[Cronify](https://github.com/Kobeep/Cronify)** - YAML-based cron management
- **PostgreSQL Community** - Excellent database and tools
- **Go Community** - Amazing language and ecosystem

---

<div align="center">

**⭐ Star this repo if you find it useful! ⭐**

Made with ❤️ by [Jakub Pospieszny](https://github.com/Kobeep)

[Report Bug](https://github.com/Kobeep/cloud-dr-orchestrator/issues) • [Request Feature](https://github.com/Kobeep/cloud-dr-orchestrator/issues) • [Documentation](https://github.com/Kobeep/cloud-dr-orchestrator/wiki)

</div>
