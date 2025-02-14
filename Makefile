.PHONY: all build test clean install uninstall logs status check help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
BINARY_NAME=image-cleanup
SERVICE_NAME=image-cleanup

# Build flags
LDFLAGS=-ldflags "-X main.Version=$$(git describe --tags --always) -X main.BuildTime=$$(date -u '+%Y-%m-%d_%H:%M:%S')"

# Environment variables
ENV_FILE=/etc/$(SERVICE_NAME)/env

# Default target
all: clean build test

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) -v ./cmd/main.go

# Run tests with coverage
test:
	$(GOTEST) -v -race -cover ./...

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f coverage.out

# Install service
install: build
	@echo "Installing $(SERVICE_NAME) service..."
	sudo ./scripts/install.sh

# Uninstall service
uninstall:
	@echo "Uninstalling $(SERVICE_NAME) service..."
	sudo ./scripts/uninstall.sh

# Service management
start:
	sudo systemctl start $(SERVICE_NAME)

stop:
	sudo systemctl stop $(SERVICE_NAME)

restart:
	sudo systemctl restart $(SERVICE_NAME)

enable:
	sudo systemctl enable $(SERVICE_NAME)
	sudo systemctl enable $(SERVICE_NAME)-health.timer

disable:
	sudo systemctl disable $(SERVICE_NAME)
	sudo systemctl disable $(SERVICE_NAME)-health.timer

# Log viewing
logs:
	sudo journalctl -u $(SERVICE_NAME) -f

error-logs:
	sudo tail -f /var/log/$(SERVICE_NAME)/error.log

health-logs:
	sudo tail -f /var/log/$(SERVICE_NAME)/health.log

# Status checks
status:
	@echo "Main service status:"
	sudo systemctl status $(SERVICE_NAME)
	@echo "\nHealth check timer status:"
	sudo systemctl status $(SERVICE_NAME)-health.timer

check: status
	@echo "\nChecking health endpoint..."
	curl -s http://localhost:8080/health

# Configuration
edit-config:
	sudo vim $(ENV_FILE)

show-config:
	@if [ -f $(ENV_FILE) ]; then \
		echo "Current configuration:"; \
		cat $(ENV_FILE); \
	else \
		echo "Configuration file not found at $(ENV_FILE)"; \
	fi

# Help
help:
	@echo "Available targets:"
	@echo "  make build       - Build the binary"
	@echo "  make test        - Run tests"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make install     - Install service"
	@echo "  make uninstall   - Uninstall service"
	@echo "  make start       - Start service"
	@echo "  make stop        - Stop service"
	@echo "  make restart     - Restart service"
	@echo "  make enable      - Enable service autostart"
	@echo "  make disable     - Disable service autostart"
	@echo "  make status      - Show service status"
	@echo "  make check       - Check service health"
	@echo "  make logs        - Show service logs"
	@echo "  make error-logs  - Show error logs"
	@echo "  make health-logs - Show health check logs"
	@echo "  make edit-config - Edit configuration file"
	@echo "  make show-config - Show current configuration"