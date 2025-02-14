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

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log "Please run as root"
    exit 1
fi

# Create directories
log "Creating directories..."
mkdir -p "$CONFIG_DIR"
mkdir -p "$LOG_DIR"

# Build the application
log "Building application..."
go build -o "$SERVICE_NAME" ./cmd/main.go

# Install binary and scripts
log "Installing files..."
mv "$SERVICE_NAME" "$BINARY_PATH"
cp "scripts/healthcheck.sh" "${BINARY_PATH}-healthcheck"
chmod +x "$BINARY_PATH"
chmod +x "${BINARY_PATH}-healthcheck"

# Create config file if it doesn't exist
if [ ! -f "$CONFIG_DIR/env" ]; then
    log "Creating default config file..."
    cat > "$CONFIG_DIR/env" << EOF
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
CLEANUP_SCHEDULE="0 0 * * *"
LOG_LEVEL=info
HTTP_PORT=8080
EOF
fi

# Create systemd service file
log "Creating systemd service..."
cat > "/etc/systemd/system/${SERVICE_NAME}.service" << EOF
[Unit]
Description=${SERVICE_DESC}
After=network.target

[Service]
Type=simple
User=${USER}
Group=${GROUP}
EnvironmentFile=${CONFIG_DIR}/env
ExecStart=${BINARY_PATH}
Restart=always
RestartSec=10
WorkingDirectory=/usr/local/bin
StandardOutput=append:${LOG_DIR}/service.log
StandardError=append:${LOG_DIR}/error.log

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

# Reload systemd
log "Reloading systemd..."
systemctl daemon-reload

# Start and enable services
log "Starting and enabling services..."
systemctl enable "$SERVICE_NAME"
systemctl start "$SERVICE_NAME"
systemctl enable "${SERVICE_NAME}-health.timer"
systemctl start "${SERVICE_NAME}-health.timer"

# Display status
log "Installation completed. Service status:"
systemctl status "$SERVICE_NAME"
systemctl status "${SERVICE_NAME}-health.timer"

log "Please update the configuration file at ${CONFIG_DIR}/env with your settings"
log "You can view logs at ${LOG_DIR}"