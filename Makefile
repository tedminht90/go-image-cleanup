.PHONY: all build clean test install uninstall start stop restart status check logs help

# Build parameters
BUILD_DIR = build
PLATFORM := $(shell if [ -f /etc/fedora-release ] || [ -f /etc/redhat-release ]; then echo "fedora"; else echo "linux"; fi)

# Service name and paths
SERVICE_NAME = image-cleanup
CONFIG_DIR = /etc/$(SERVICE_NAME)
LOG_DIR = /var/log/$(SERVICE_NAME)

# Colors for output
COLOR_RESET = \033[0m
COLOR_BOLD = \033[1m
COLOR_GREEN = \033[32m
COLOR_YELLOW = \033[33m
COLOR_BLUE = \033[34m

# Logging function
define log
	@echo "$(COLOR_BLUE)[$(shell date '+%Y-%m-%d %H:%M:%S')]$(COLOR_RESET) $(1)"
endef

# Default target
all: clean build test

# Build the application
build:
	$(call log,"Building application for platform: $(PLATFORM)...")
	@./scripts/build.sh
	$(call log,"Build completed")

# Run tests with coverage
test:
	$(call log,"Running tests...")
	@go test -v -race -cover ./...
	$(call log,"Tests completed")

# Clean build artifacts
clean:
	$(call log,"Cleaning build artifacts...")
	@rm -rf $(BUILD_DIR)
	@go clean
	$(call log,"Clean completed")

# Install service (requires root)
install:
	$(call log,"Installing service for platform: $(PLATFORM)...")
	@cd $(BUILD_DIR)/$(PLATFORM)/scripts && sudo ./install.sh

# Uninstall service (requires root)
uninstall:
	$(call log,"Uninstalling service...")
	@sudo ./scripts/uninstall.sh

# Service management commands
start:
	$(call log,"Starting service...")
	@sudo systemctl start $(SERVICE_NAME)
	@sudo systemctl start $(SERVICE_NAME)-health.timer

stop:
	$(call log,"Stopping service...")
	@sudo systemctl stop $(SERVICE_NAME)
	@sudo systemctl stop $(SERVICE_NAME)-health.timer

restart:
	$(call log,"Restarting service...")
	@sudo systemctl restart $(SERVICE_NAME)
	@sudo systemctl restart $(SERVICE_NAME)-health.timer

status:
	$(call log,"Service status:")
	@sudo systemctl status $(SERVICE_NAME)
	@echo ""
	$(call log,"Health timer status:")
	@sudo systemctl status $(SERVICE_NAME)-health.timer

check:
	$(call log,"Checking service health...")
	@curl -s http://localhost:8080/health | jq . || echo "Failed to get health status"

enable:
	$(call log,"Enabling services...")
	@sudo systemctl enable $(SERVICE_NAME)
	@sudo systemctl enable $(SERVICE_NAME)-health.timer

disable:
	$(call log,"Disabling services...")
	@sudo systemctl disable $(SERVICE_NAME)
	@sudo systemctl disable $(SERVICE_NAME)-health.timer

# Log viewing commands
logs:
	$(call log,"Viewing service logs...")
	@sudo journalctl -u $(SERVICE_NAME) -f

error-logs:
	$(call log,"Viewing error logs...")
	@sudo tail -f $(LOG_DIR)/error.log

health-logs:
	$(call log,"Viewing health check logs...")
	@sudo tail -f $(LOG_DIR)/health.log

service-logs:
	$(call log,"Viewing all service related logs...")
	@sudo journalctl -u $(SERVICE_NAME) -u $(SERVICE_NAME)-health.timer -f

# Configuration commands
edit-config:
	$(call log,"Editing configuration...")
	@sudo vim $(CONFIG_DIR)/.env

show-config:
	$(call log,"Current configuration:")
	@sudo cat $(CONFIG_DIR)/.env

version:
	$(call log,"Version information:")
	@if [ -f "$(BUILD_DIR)/$(PLATFORM)/version.txt" ]; then \
		cat $(BUILD_DIR)/$(PLATFORM)/version.txt; \
	else \
		echo "$(COLOR_YELLOW)Version information not available. Build the application first.$(COLOR_RESET)"; \
	fi

# Monitor commands
monitor: status check
	$(call log,"Service monitoring:")
	@systemctl list-timers | grep $(SERVICE_NAME)

# Troubleshooting commands
troubleshoot:
	$(call log,"Running troubleshooting checks...")
	@echo "$(COLOR_BOLD)System Information:$(COLOR_RESET)"
	@echo "Platform: $(PLATFORM)"
	@echo "Service Status:"
	@systemctl is-active $(SERVICE_NAME) || true
	@echo "Timer Status:"
	@systemctl is-active $(SERVICE_NAME)-health.timer || true
	@echo "Recent Logs:"
	@journalctl -u $(SERVICE_NAME) --no-pager -n 10
	@echo "Health Check Status:"
	@curl -s http://localhost:8080/health || echo "Health check failed"

# Help
help:
	@echo "$(COLOR_BOLD)Available commands:$(COLOR_RESET)"
	@echo ""
	@echo "$(COLOR_BOLD)Build commands:$(COLOR_RESET)"
	@echo "  make build         - Build the application (Platform: $(PLATFORM))"
	@echo "  make test          - Run tests"
	@echo "  make clean         - Clean build artifacts"
	@echo ""
	@echo "$(COLOR_BOLD)Installation:$(COLOR_RESET)"
	@echo "  make install       - Install service"
	@echo "  make uninstall     - Uninstall service"
	@echo ""
	@echo "$(COLOR_BOLD)Service control:$(COLOR_RESET)"
	@echo "  make start         - Start service and timer"
	@echo "  make stop          - Stop service and timer"
	@echo "  make restart       - Restart service and timer"
	@echo "  make status        - Show service status"
	@echo "  make check         - Check service health"
	@echo "  make enable        - Enable service autostart"
	@echo "  make disable       - Disable service autostart"
	@echo "  make monitor       - Monitor service status"
	@echo ""
	@echo "$(COLOR_BOLD)Logs:$(COLOR_RESET)"
	@echo "  make logs          - View service logs"
	@echo "  make error-logs    - View error logs"
	@echo "  make health-logs   - View health check logs"
	@echo "  make service-logs  - View all service related logs"
	@echo ""
	@echo "$(COLOR_BOLD)Configuration:$(COLOR_RESET)"
	@echo "  make edit-config   - Edit configuration"
	@echo "  make show-config   - Show current configuration"
	@echo "  make version       - Show build version"
	@echo ""
	@echo "$(COLOR_BOLD)Troubleshooting:$(COLOR_RESET)"
	@echo "  make troubleshoot  - Run troubleshooting checks"
