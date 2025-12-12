#!/bin/bash
# Installation script for Cloud DR Orchestrator automated backups
# Supports systemd (Linux) and cron (Linux/macOS)

set -e

INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/cloud-dr-orchestrator"
WORK_DIR="/opt/cloud-dr-orchestrator"
SYSTEMD_DIR="/etc/systemd/system"

echo "🚀 Cloud DR Orchestrator - Automated Backup Setup"
echo "=================================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Please run as root (use sudo)"
    exit 1
fi

# Check if orchestrator binary exists
if ! command -v orchestrator &> /dev/null; then
    echo "❌ orchestrator binary not found in PATH"
    echo "   Please install it first: go install github.com/Kobeep/cloud-dr-orchestrator/cmd/orchestrator@latest"
    exit 1
fi

echo "✅ Found orchestrator at: $(which orchestrator)"
echo ""

# Create working directory
echo "📁 Creating working directory: $WORK_DIR"
mkdir -p "$WORK_DIR"
chmod 755 "$WORK_DIR"

# Create config directory
echo "📁 Creating config directory: $CONFIG_DIR"
mkdir -p "$CONFIG_DIR"
chmod 700 "$CONFIG_DIR"

# Copy example config
echo "📝 Installing configuration template"
cp configs/backup.env.example "$CONFIG_DIR/backup.env"
chmod 600 "$CONFIG_DIR/backup.env"

echo ""
echo "⚠️  IMPORTANT: Edit $CONFIG_DIR/backup.env with your settings!"
echo ""

# Detect init system
if command -v systemctl &> /dev/null && [ -d "/etc/systemd/system" ]; then
    echo "🔧 Detected systemd - installing service and timer"

    # Install systemd files
    cp systemd/orchestrator-backup.service "$SYSTEMD_DIR/"
    cp systemd/orchestrator-backup.timer "$SYSTEMD_DIR/"

    # Reload systemd
    systemctl daemon-reload

    # Enable timer (don't start yet - user needs to configure first)
    systemctl enable orchestrator-backup.timer

    echo "✅ Systemd service installed!"
    echo ""
    echo "Next steps:"
    echo "  1. Edit config: nano $CONFIG_DIR/backup.env"
    echo "  2. Start timer: systemctl start orchestrator-backup.timer"
    echo "  3. Check status: systemctl status orchestrator-backup.timer"
    echo "  4. View logs: journalctl -u orchestrator-backup.service"

else
    echo "🔧 systemd not found - installing cron job"

    # Create cron script wrapper
    cat > "$INSTALL_DIR/orchestrator-backup-cron.sh" << 'EOF'
#!/bin/bash
source /etc/cloud-dr-orchestrator/backup.env

# Run backup
/usr/local/bin/orchestrator backup \
  --name "$BACKUP_NAME" \
  --db-name "$DB_NAME" \
  --db-host "$DB_HOST" \
  --db-port "$DB_PORT" \
  --db-user "$DB_USER" \
  --db-password "$DB_PASSWORD" \
  2>&1 | logger -t orchestrator-backup

# Upload to cloud
BACKUP_FILE=$(ls -t /opt/cloud-dr-orchestrator/*.tar.gz | head -1)
/usr/local/bin/orchestrator upload \
  --file "$BACKUP_FILE" \
  --bucket "$OCI_BUCKET" \
  --compartment "$OCI_COMPARTMENT" \
  2>&1 | logger -t orchestrator-backup
EOF

    chmod +x "$INSTALL_DIR/orchestrator-backup-cron.sh"

    echo "✅ Cron wrapper script installed!"
    echo ""
    echo "Next steps:"
    echo "  1. Edit config: nano $CONFIG_DIR/backup.env"
    echo "  2. Add to crontab: crontab -e"
    echo ""
    echo "Example crontab entry (daily at 2 AM):"
    echo "  0 2 * * * $INSTALL_DIR/orchestrator-backup-cron.sh"
fi

echo ""
echo "✨ Setup complete! Oracle Cloud Free Tier compatible."
