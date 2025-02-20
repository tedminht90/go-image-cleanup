#!/bin/bash

# Exit on any error
set -e

# Config
SERVICE_NAME="image-cleanup"
BINARY_PATH="/usr/local/bin/${SERVICE_NAME}"
CONFIG_DIR="/etc/${SERVICE_NAME}"
LOG_DIR="/var/log/${SERVICE_NAME}"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Detect platform
detect_platform() {
    if [ -f /etc/fedora-release ] || [ -f /etc/redhat-release ]; then
        echo "fedora"
    else
        echo "linux"
    fi
}

PLATFORM=$(detect_platform)
log "Detected platform: ${PLATFORM}"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log "Please run as root"
    exit 1
fi

log "Starting uninstallation process..."

# Stop and disable services
log "Stopping and disabling services..."
systemctl stop "$SERVICE_NAME" || true
systemctl stop "${SERVICE_NAME}-health.timer" || true
systemctl stop "${SERVICE_NAME}-health.service" || true
systemctl disable "$SERVICE_NAME" || true
systemctl disable "${SERVICE_NAME}-health.timer" || true
systemctl disable "${SERVICE_NAME}-health.service" || true

# Wait for services to stop
log "Waiting for services to stop..."
sleep 2

# Remove service files
log "Removing systemd service files..."
rm -f "/etc/systemd/system/${SERVICE_NAME}.service"
rm -f "/etc/systemd/system/${SERVICE_NAME}-health.service"
rm -f "/etc/systemd/system/${SERVICE_NAME}-health.timer"

# Remove binary and scripts
log "Removing binaries..."
rm -f "$BINARY_PATH"
rm -f "${BINARY_PATH}-healthcheck"

# Ask about configuration and logs
read -p "Do you want to remove configuration and log files? (y/N) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    log "Removing configuration directory: ${CONFIG_DIR}"
    rm -rf "$CONFIG_DIR"

    log "Removing log directory: ${LOG_DIR}"
    rm -rf "$LOG_DIR"

    log "Configuration and log files removed"
else
    log "Keeping configuration and log files"
    log "Config directory: ${CONFIG_DIR}"
    log "Log directory: ${LOG_DIR}"
fi

# Reload systemd
log "Reloading systemd..."
systemctl daemon-reload

log "Uninstallation completed successfully!"
log "Platform: ${PLATFORM}"

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log "Note: Configuration and log files were preserved"
    log "To remove them manually later, run:"
    log "rm -rf ${CONFIG_DIR}"
    log "rm -rf ${LOG_DIR}"
fi
