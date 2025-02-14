#!/bin/bash

# Exit on any error
set -e

# Config
SERVICE_NAME="image-cleanup"
OUTPUT_DIR="./build"

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Check if go is installed
if ! command -v go >/dev/null 2>&1; then
    log "Error: Go is not installed"
    exit 1
fi

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
if [[ "${GO_VERSION}" < "1.21" ]]; then
    log "Error: Go version must be 1.21 or higher (current: ${GO_VERSION})"
    exit 1
fi

# Create build directory
log "Creating build directory..."
mkdir -p "$OUTPUT_DIR"

# Clean previous build
log "Cleaning previous build..."
rm -f "${OUTPUT_DIR}/${SERVICE_NAME}"

# Build the application
log "Building application..."
export CGO_ENABLED=0  # Disable CGO for static binary
export GOOS=linux    # Ensure we're building for Linux
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

go build -o "${OUTPUT_DIR}/${SERVICE_NAME}" \
    -ldflags "-X main.Version=$COMMIT_HASH -X main.BuildTime=$BUILD_TIME" \
    ./cmd/main.go

if [ $? -ne 0 ]; then
    log "Error: Build failed"
    exit 1
fi

# Copy required scripts
log "Copying scripts..."
cp scripts/healthcheck.sh "${OUTPUT_DIR}/"
cp scripts/install.sh "${OUTPUT_DIR}/"
chmod +x "${OUTPUT_DIR}/healthcheck.sh"
chmod +x "${OUTPUT_DIR}/install.sh"

# Create version file
echo "${COMMIT_HASH} - ${BUILD_TIME}" > "${OUTPUT_DIR}/version.txt"

# Verify binary
log "Verifying binary..."
if ! "${OUTPUT_DIR}/${SERVICE_NAME}" --version >/dev/null 2>&1; then
    log "Error: Binary verification failed"
    rm -f "${OUTPUT_DIR}/${SERVICE_NAME}"
    exit 1
fi

log "Build completed successfully!"
log "Build outputs are in: $OUTPUT_DIR"
log "Version information:"
cat "${OUTPUT_DIR}/version.txt"