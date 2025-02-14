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

# Colors for logging
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function for logging
log() {
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

log_error() {
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] ${RED}ERROR: $1${NC}"
}

log_success() {
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] ${GREEN}$1${NC}"
}

log_warning() {
    echo -e "[$(date '+%Y-%m-%d %H:%M:%S')] ${YELLOW}WARNING: $1${NC}"
}

# Function to check port availability
check_port() {
    local port=$1
    if netstat -tln | grep -q ":${port} "; then
        log_error "Port ${port} is already in use"
        return 1
    fi
    return 0
}

# Check prerequisites
check_prerequisites() {
    local missing_deps=()

    # Check if we're in the build directory
    if [ ! -f "${SERVICE_NAME}" ] || [ ! -f "healthcheck.sh" ]; then
        log_error "Required files not found. Please run build.sh first"
        exit 1
    fi

    # Check OS
    if [ "$(uname)" != "Linux" ]; then
        log_error "This script only supports Linux"
        exit 1
    fi

    # Check if running as root
    if [ "$EUID" -ne 0 ]; then
        log_error "Please run as root"
        exit 1
    fi

    # Check systemd
    if ! pidof systemd >/dev/null; then
        log_error "systemd is not running"
        exit 1
    fi

    # Check required commands
    local cmds=("crictl" "netstat" "systemctl")
    for cmd in "${cmds[@]}"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            missing_deps+=("$cmd")
        fi
    done

    if [ ${#missing_deps[@]} -ne 0 ]; then
        log_error "Missing required dependencies: ${missing_deps[*]}"
        exit 1
    fi
}

# Create necessary directories
create_directories() {
    log "Creating directories..."
    mkdir -p "$CONFIG_DIR"
    mkdir -p "$LOG_DIR"
    chmod 750 "$CONFIG_DIR"
    chmod 750 "$LOG_DIR"
}

# Install files
install_files() {
    log "Installing files..."
    cp "${SERVICE_NAME}" "$BINARY_PATH"
    cp "healthcheck.sh" "${BINARY_PATH}-healthcheck"
    chmod 755 "$BINARY_PATH"
    chmod 755 "${BINARY_PATH}-healthcheck"
}

# Main installation process
main() {
    log "Starting installation of ${SERVICE_DESC}..."

    # Check prerequisites
    check_prerequisites

    # Create directories
    create_directories

    # Install files
    install_files

    # [Previous installation steps remain the same...]

    # Final verification
    local port=$(grep "^HTTP_PORT=" "$CONFIG_DIR/env" | cut -d'=' -f2)
    if ! check_port "$port"; then
        log_warning "Port ${port} is in use. You may need to modify the configuration."
    fi

    # Display version information
    if [ -f "version.txt" ]; then
        log "Installing version:"
        cat "version.txt"
    fi

    log_success "Installation completed successfully!"
    log_success "Service status:"
    systemctl status "$SERVICE_NAME" --no-pager
    systemctl status "${SERVICE_NAME}-health.timer" --no-pager

    log "Important next steps:"
    log "1. Update configuration at: ${CONFIG_DIR}/env"
    log "2. Verify logs at: ${LOG_DIR}"
    log "3. Check service health: curl http://localhost:${port}/health"
    log "4. View logs: journalctl -u ${SERVICE_NAME} -f"
}

# Run main installation
main
