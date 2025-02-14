#!/bin/bash

# Exit on any error
set -e

# Config
SERVICE_NAME="image-cleanup"
SERVICE_DESC="Image Cleanup Service"
BINARY_PATH="/usr/local/bin/${SERVICE_NAME}"
CONFIG_DIR="/etc/${SERVICE_NAME}"
LOG_DIR="/var/log/${SERVICE_NAME}"
USER="root"
GROUP="root"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Check if we're in the build directory
if [ ! -f "${SERVICE_NAME}" ] || [ ! -f "healthcheck.sh" ]; then
    log "Error: Required files not found. Please run build.sh first"
    exit 1
fi

# Check OS
if [ "$(uname)" != "Linux" ]; then
    log "Error: This script only supports Linux"
    exit 1
fi

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log "Please run as root"
    exit 1
fi

# Check systemd
if ! pidof systemd >/dev/null; then
    log "Error: systemd is not running"
    exit 1
fi

# Check crictl
if ! command -v crictl >/dev/null 2>&1; then
    log "Error: crictl is not installed"
    exit 1
fi

# Create directories
log "Creating directories..."
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"

# Install binary and scripts
log "Installing files..."
cp "${SERVICE_NAME}" "$BINARY_PATH"
cp "healthcheck.sh" "${BINARY_PATH}-healthcheck"
chmod +x "$BINARY_PATH"
chmod +x "${BINARY_PATH}-healthcheck"

# Create config file if it doesn't exist
if [ ! -f "$CONFIG_DIR/.env" ]; then
    log "Creating default config file..."
    cat > "$CONFIG_DIR/.env" << EOF
# Service configuration
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
CLEANUP_SCHEDULE="0 0 * * *"
HTTP_PORT=8080

# Logger configuration
LOG_LEVEL=info
LOG_DIR=/var/log/image-cleanup
LOG_MAX_SIZE=100       # Maximum size of each log file in megabytes
LOG_MAX_BACKUPS=5      # Number of old log files to keep
LOG_MAX_AGE=30         # Days to keep old log files
LOG_COMPRESS=true      # Compress old log files
EOF

    log "Created default configuration file at $CONFIG_DIR/.env"
fi

# Create systemd service file
log "Creating systemd service..."
cat > "/etc/systemd/system/${SERVICE_NAME}.service" << EOF
[Unit]
Description=${SERVICE_DESC}
After=network.target
StartLimitIntervalSec=0

[Service]
Type=simple
User=${USER}
Group=${GROUP}
EnvironmentFile=${CONFIG_DIR}/.env
ExecStart=${BINARY_PATH}
Restart=always
RestartSec=10
WorkingDirectory=/usr/local/bin
StandardOutput=append:${LOG_DIR}/service.log
StandardError=append:${LOG_DIR}/error.log
LimitNOFILE=65535

[Install]
WantedBy=multi-user.target
EOF

# Create health check service
log "Creating health check service..."
cat > "/etc/systemd/system/${SERVICE_NAME}-health.service" << EOF
[Unit]
Description=${SERVICE_DESC} Health Check
After=network.target

[Service]
Type=oneshot
ExecStart=${BINARY_PATH}-healthcheck
StandardOutput=append:${LOG_DIR}/health.log
StandardError=append:${LOG_DIR}/health-error.log

[Install]
WantedBy=multi-user.target
EOF

# Create health check timer
log "Creating health check timer..."
cat > "/etc/systemd/system/${SERVICE_NAME}-health.timer" << EOF
[Unit]
Description=Run ${SERVICE_DESC} Health Check periodically

[Timer]
OnBootSec=1min
OnUnitActiveSec=5min

[Install]
WantedBy=timers.target
EOF

# Set permissions
log "Setting permissions..."
chown -R ${USER}:${GROUP} "$CONFIG_DIR"
chown -R ${USER}:${GROUP} "$LOG_DIR"
chmod 600 "$CONFIG_DIR/env"
chmod 644 "/etc/systemd/system/${SERVICE_NAME}.service"
chmod 644 "/etc/systemd/system/${SERVICE_NAME}-health.service"
chmod 644 "/etc/systemd/system/${SERVICE_NAME}-health.timer"

# Verify systemd files
log "Verifying systemd files..."
if ! systemd-analyze verify "/etc/systemd/system/${SERVICE_NAME}.service"; then
    log "Error: Service file verification failed"
    exit 1
fi

# Reload systemd
log "Reloading systemd..."
systemctl daemon-reload

# Start and enable services
log "Starting and enabling services..."
systemctl enable "$SERVICE_NAME" || { log "Error enabling service"; exit 1; }
systemctl start "$SERVICE_NAME" || { log "Error starting service"; exit 1; }
systemctl enable "${SERVICE_NAME}-health.timer" || { log "Error enabling health timer"; exit 1; }
systemctl start "${SERVICE_NAME}-health.timer" || { log "Error starting health timer"; exit 1; }

# Verify service is running
log "Verifying service status..."
if ! systemctl is-active --quiet "$SERVICE_NAME"; then
    log "Error: Service failed to start"
    journalctl -u "$SERVICE_NAME" --no-pager -n 50
    exit 1
fi

# Display version information
if [ -f "version.txt" ]; then
    log "Installing version:"
    cat "version.txt"
fi

# Display status and final instructions
log "Installation completed successfully!"
log "Service status:"
systemctl status "$SERVICE_NAME" --no-pager
systemctl status "${SERVICE_NAME}-health.timer" --no-pager

log "Important next steps:"
log "1. Update configuration at: ${CONFIG_DIR}/env"
log "2. Verify logs at: ${LOG_DIR}"
log "3. Check service health: curl http://localhost:8080/health"
log "4. View logs: journalctl -u ${SERVICE_NAME} -f"
