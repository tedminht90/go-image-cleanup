.PHONY: all build clean test install uninstall start stop restart status check logs help

# Build parameters
BUILD_DIR=build

# Service name
SERVICE_NAME=image-cleanup
CONFIG_DIR=/etc/$(SERVICE_NAME)

# Default target
all: clean build test

# Build the application
build:
	@echo "Building application..."
	@./scripts/build.sh

# Run tests with coverage
test:
	@echo "Running tests..."
	go test -v -race -cover ./...
	@echo "Tests completed"

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@go clean
	@echo "Clean completed"

# Install service (requires root)
install:
	@echo "Installing service..."
	@cd $(BUILD_DIR) && sudo ./install.sh

# Uninstall service (requires root)
uninstall:
	@echo "Uninstalling service..."
	@sudo ./scripts/uninstall.sh

# Service management commands
start:
	@sudo systemctl start $(SERVICE_NAME)

stop:
	@sudo systemctl stop $(SERVICE_NAME)

restart:
	@sudo systemctl restart $(SERVICE_NAME)

status:
	@sudo systemctl status $(SERVICE_NAME)
	@sudo systemctl status $(SERVICE_NAME)-health.timer

check:
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health
	@echo ""

enable:
	@sudo systemctl enable $(SERVICE_NAME)
	@sudo systemctl enable $(SERVICE_NAME)-health.timer

disable:
	@sudo systemctl disable $(SERVICE_NAME)
	@sudo systemctl disable $(SERVICE_NAME)-health.timer

# Log viewing commands
logs:
	@sudo journalctl -u $(SERVICE_NAME) -f

error-logs:
	@sudo tail -f /var/log/$(SERVICE_NAME)/error.log

health-logs:
	@sudo tail -f /var/log/$(SERVICE_NAME)/health.log

# Configuration commands
edit-config:
	@sudo vim $(CONFIG_DIR)/.env

show-config:
	@sudo cat $(CONFIG_DIR)/.env

version:
	@if [ -f "$(BUILD_DIR)/version.txt" ]; then \
		cat $(BUILD_DIR)/version.txt; \
	else \
		echo "Version information not available. Build the application first."; \
	fi

# Help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build commands:"
	@echo "  make build         - Build the application"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "Installation:"
	@echo "  make install       - Install service"
	@echo "  make uninstall     - Uninstall service"
	@echo ""
	@echo "Service control:"
	@echo "  make start         - Start service"
	@echo "  make stop          - Stop service"
	@echo "  make restart       - Restart service"
	@echo "  make status        - Show service status"
	@echo "  make check         - Check service health"
	@echo "  make enable        - Enable service autostart"
	@echo "  make disable       - Disable service autostart"
	@echo ""
	@echo "Logs:"
	@echo "  make logs          - View service logs"
	@echo "  make error-logs    - View error logs"
	@echo "  make health-logs   - View health check logs"
	@echo ""
	@echo "Configuration:"
	@echo "  make edit-config   - Edit configuration"
	@echo "  make show-config   - Show current configuration"
	@echo "  make version       - Show build version"
