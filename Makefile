.PHONY: all build clean test install uninstall start stop restart status check logs help db-check db-backup db-restore

# Build parameters
BUILD_DIR = build
PLATFORM := $(shell if [ -f /etc/fedora-release ] || [ -f /etc/redhat-release ]; then echo "fedora"; else echo "linux"; fi)

# Service name and paths
SERVICE_NAME = image-cleanup
CONFIG_DIR = /etc/$(SERVICE_NAME)
LOG_DIR = /var/log/$(SERVICE_NAME)
DATA_DIR = /var/lib/$(SERVICE_NAME)
DB_PATH = $(DATA_DIR)/cleanup.db
DB_BACKUP_DIR = $(DATA_DIR)/backups

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

# Verify build directory exists
verify-build:
	@if [ ! -d "$(BUILD_DIR)/$(PLATFORM)" ]; then \
		echo "$(COLOR_YELLOW)Build directory not found. Running build first...$(COLOR_RESET)"; \
		make build; \
	fi

# Install service (requires root)
install: verify-build
	$(call log,"Installing service for platform: $(PLATFORM)...")
	@if [ ! -f "$(BUILD_DIR)/$(PLATFORM)/$(SERVICE_NAME)" ]; then \
		echo "$(COLOR_YELLOW)Binary not found in build directory. Please run 'make build' first.$(COLOR_RESET)"; \
		exit 1; \
	fi
	@cd $(BUILD_DIR)/$(PLATFORM) && sudo ./scripts/install.sh

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

# Database commands
db-check:
	$(call log,"Checking database...")
	@if [ ! -f "$(DB_PATH)" ]; then \
		echo "$(COLOR_YELLOW)Database not found at $(DB_PATH)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)Database information:$(COLOR_RESET)"
	@sudo sqlite3 $(DB_PATH) "PRAGMA integrity_check;"
	@echo "$(COLOR_BOLD)Database schema:$(COLOR_RESET)"
	@sudo sqlite3 $(DB_PATH) ".schema"
	@echo "$(COLOR_BOLD)Recent cleanup records:$(COLOR_RESET)"
	@sudo sqlite3 $(DB_PATH) "SELECT id, host_info, datetime(start_time), datetime(end_time), duration_ms/1000.0 as seconds, total_count, removed, skipped FROM cleanup_results ORDER BY start_time DESC LIMIT 5;"

db-stats:
	$(call log,"Database statistics...")
	@if [ ! -f "$(DB_PATH)" ]; then \
		echo "$(COLOR_YELLOW)Database not found at $(DB_PATH)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_BOLD)Total records:$(COLOR_RESET)"
	@sudo sqlite3 $(DB_PATH) "SELECT COUNT(*) FROM cleanup_results;"
	@echo "$(COLOR_BOLD)Records by month:$(COLOR_RESET)"
	@sudo sqlite3 $(DB_PATH) "SELECT strftime('%Y-%m', start_time) as month, COUNT(*) as count FROM cleanup_results GROUP BY month ORDER BY month DESC;"
	@echo "$(COLOR_BOLD)Images removed summary:$(COLOR_RESET)"
	@sudo sqlite3 $(DB_PATH) "SELECT SUM(removed) as total_removed, SUM(skipped) as total_skipped, SUM(total_count) as total_images FROM cleanup_results;"
	@echo "$(COLOR_BOLD)Average duration:$(COLOR_RESET)"
	@sudo sqlite3 $(DB_PATH) "SELECT AVG(duration_ms)/1000.0 as avg_seconds FROM cleanup_results;"

db-backup:
	$(call log,"Backing up database...")
	@sudo mkdir -p $(DB_BACKUP_DIR)
	@BACKUP_FILE="$(DB_BACKUP_DIR)/cleanup-$(shell date '+%Y%m%d-%H%M%S').db"; \
	sudo sqlite3 $(DB_PATH) ".backup '$${BACKUP_FILE}'"; \
	echo "$(COLOR_GREEN)Database backed up to $${BACKUP_FILE}$(COLOR_RESET)"

db-restore:
	$(call log,"Restoring database...")
	@if [ -z "$(BACKUP)" ]; then \
		echo "$(COLOR_YELLOW)Please specify backup file: make db-restore BACKUP=/path/to/backup.db$(COLOR_RESET)"; \
		exit 1; \
	fi
	@if [ ! -f "$(BACKUP)" ]; then \
		echo "$(COLOR_YELLOW)Backup file not found: $(BACKUP)$(COLOR_RESET)"; \
		exit 1; \
	fi
	@echo "$(COLOR_YELLOW)WARNING: This will overwrite the current database. Continue? [y/N]$(COLOR_RESET)"; \
	read CONFIRM; \
	if [ "$${CONFIRM}" != "y" ] && [ "$${CONFIRM}" != "Y" ]; then \
		echo "$(COLOR_YELLOW)Restore cancelled.$(COLOR_RESET)"; \
		exit 1; \
	fi; \
	sudo systemctl stop $(SERVICE_NAME); \
	sudo sqlite3 $(DB_PATH) ".restore '$(BACKUP)'"; \
	sudo systemctl start $(SERVICE_NAME); \
	echo "$(COLOR_GREEN)Database restored from $(BACKUP)$(COLOR_RESET)"

# API commands
api-check:
	$(call log,"Checking API endpoints...")
	@echo "$(COLOR_BOLD)Health endpoint:$(COLOR_RESET)"
	@curl -s http://localhost:8080/health | jq . || echo "Failed to get health status"
	@echo "$(COLOR_BOLD)Cleanup status endpoint:$(COLOR_RESET)"
	@curl -s http://localhost:8080/api/v1/cleanup | jq . || echo "Failed to get cleanup status"
	@echo "$(COLOR_BOLD)Metrics endpoint:$(COLOR_RESET)"
	@curl -s http://localhost:8080/metrics | head -n 10 || echo "Failed to get metrics"

# Monitor commands
monitor: status check api-check
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
	@echo "Database Status:"
	@if [ -f "$(DB_PATH)" ]; then \
		echo "Database exists and size: $$(sudo du -h $(DB_PATH) | cut -f1)"; \
		sudo sqlite3 $(DB_PATH) "PRAGMA integrity_check;" || echo "Database integrity check failed"; \
	else \
		echo "Database not found at $(DB_PATH)"; \
	fi
	@echo "Recent Logs:"
	@journalctl -u $(SERVICE_NAME) --no-pager -n 10
	@echo "API Status:"
	@curl -s http://localhost:8080/api/v1/cleanup || echo "Cleanup API failed"
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
	@echo "$(COLOR_BOLD)Database:$(COLOR_RESET)"
	@echo "  make db-check      - Check database integrity and schema"
	@echo "  make db-stats      - Show database statistics"
	@echo "  make db-backup     - Backup database"
	@echo "  make db-restore    - Restore database (specify BACKUP=/path/to/file.db)"
	@echo ""
	@echo "$(COLOR_BOLD)API:$(COLOR_RESET)"
	@echo "  make api-check     - Check all API endpoints"
	@echo ""
	@echo "$(COLOR_BOLD)Configuration:$(COLOR_RESET)"
	@echo "  make edit-config   - Edit configuration"
	@echo "  make show-config   - Show current configuration"
	@echo "  make version       - Show build version"
	@echo ""
	@echo "$(COLOR_BOLD)Troubleshooting:$(COLOR_RESET)"
	@echo "  make troubleshoot  - Run troubleshooting checks"
