#!/bin/bash

# Exit on any error
set -e

# Service configuration
SERVICE_NAME="image-cleanup"
OUTPUT_DIR="./build"
PLATFORMS=("linux" "fedora")  # Build for both Ubuntu/Debian and Fedora CoreOS

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1"
}

# Check build requirements
check_requirements() {
    log "Checking build requirements..."

    # Check if go is installed
    if ! command -v go >/dev/null 2>&1; then
        log "Error: Go is not installed"
        exit 1
    fi

    # Check Go version (requires 1.21 or higher)
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if [[ "${GO_VERSION}" < "1.21" ]]; then
        log "Error: Go version must be 1.21 or higher (current: ${GO_VERSION})"
        exit 1
    fi

    log "Build requirements met: Go ${GO_VERSION}"
}

# Setup build environment
setup_build_env() {
    log "Setting up build environment..."

    # Clean and create build directories
    for platform in "${PLATFORMS[@]}"; do
        log "Setting up ${platform} build directory..."
        mkdir -p "${OUTPUT_DIR}/${platform}"
        rm -f "${OUTPUT_DIR}/${platform}/${SERVICE_NAME}"
    done
}

# Get build information
get_build_info() {
    log "Getting build information..."

    # Get commit hash
    COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

    # Get build time in ICT timezone
    BUILD_TIME=$(TZ=Asia/Bangkok date '+%Y-%m-%d_%H:%M:%S_ICT')

    log "Build info - Version: ${COMMIT_HASH}, Time: ${BUILD_TIME}"
    return 0
}

# Build binary for a specific platform
build_binary() {
    local platform=$1
    log "Building binary for ${platform}..."

    # Set output path
    local output="${OUTPUT_DIR}/${platform}/${SERVICE_NAME}"

    # Set build tags based on platform
    local tags="containers_image_openpgp"  # Both platforms use the same tags as they both use crictl

    # Build the binary
    export CGO_ENABLED=0  # Disable CGO for static binary
    export GOOS=linux    # Build for Linux

    log "Running go build for ${platform}..."
    go build \
        -o "$output" \
        -tags "$tags" \
        -ldflags "-X main.Version=$COMMIT_HASH -X main.BuildTime=$BUILD_TIME" \
        ./cmd/main.go

    # Verify binary
    if [ ! -f "$output" ]; then
        log "Error: Binary not created for ${platform}"
        return 1
    fi

    chmod +x "$output"
    local size=$(ls -lh "$output" | awk '{print $5}')
    log "${platform} binary built successfully (size: ${size})"
}

# Copy additional files
copy_additional_files() {
    local platform=$1
    log "Copying additional files for ${platform}..."

    # Create scripts directory
    mkdir -p "${OUTPUT_DIR}/${platform}/scripts"

    # Copy scripts
    cp scripts/healthcheck.sh "${OUTPUT_DIR}/${platform}/scripts/"
    cp scripts/install.sh "${OUTPUT_DIR}/${platform}/scripts/"
    cp scripts/uninstall.sh "${OUTPUT_DIR}/${platform}/scripts/"

    # Set permissions
    chmod +x "${OUTPUT_DIR}/${platform}/scripts/"*.sh

    # Create version file
    cat > "${OUTPUT_DIR}/${platform}/version.txt" << EOF
Platform: ${platform}
Version: ${COMMIT_HASH}
Build Time: ${BUILD_TIME}
Build Tags: containers_image_openpgp
EOF
}

# Main build process
main() {
    log "Starting build process for ${SERVICE_NAME}..."

    check_requirements
    setup_build_env
    get_build_info

    # Build for each platform
    for platform in "${PLATFORMS[@]}"; do
        log "Processing ${platform} platform..."
        build_binary "$platform"
        copy_additional_files "$platform"
        log "${platform} build completed"
        echo "----------------------------------------"
    done

    # Print completion message
    log "Build completed successfully!"
    log "Build outputs are in: ${OUTPUT_DIR}"

    # Show version information
    for platform in "${PLATFORMS[@]}"; do
        log "Version information for ${platform}:"
        cat "${OUTPUT_DIR}/${platform}/version.txt"
        echo "----------------------------------------"
    done
}

# Run the build
main
