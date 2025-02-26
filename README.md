# Image Cleanup Service

Automated service for cleaning up unused container images with Telegram notifications.

## Features

- Automated cleanup of unused container images
- Telegram notifications with cleanup results and host info (ICT+7 timezone)
- Health monitoring with auto-recovery
- Prometheus metrics endpoint
- Systemd service integration
- Parallel processing with worker pool
- Comprehensive logging with automatic rotation
- SQLite database for storing and querying cleanup history
- REST API for accessing cleanup results
- Static binary build for Linux

## Prerequisites

- Go 1.21 or higher
- Linux with systemd
- Root access for service installation
- crictl installed and configured
- Telegram bot token and chat ID
- SQLite3 (installed automatically by the installer)

## Building

### Development Build

```bash
# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### Build Artifacts

After a successful build, the following files will be created in the `build/` directory:

- `image-cleanup`: Main service binary
- `healthcheck.sh`: Health check script
- `install.sh`: Installation script
- `version.txt`: Build version information

## Installation

1. Build the application:

```bash
make build
```

2. Install the service:

```bash
make install
```

3. Configure the service:

```bash
make edit-config
```

Required configuration in `/etc/image-cleanup/env`:

```env
# Service configuration
TELEGRAM_BOT_TOKEN=your_bot_token   # Your Telegram bot token
TELEGRAM_CHAT_ID=your_chat_id       # Target Telegram chat ID
CLEANUP_SCHEDULE="0 0 * * *"        # Cron schedule for cleanup job
HTTP_PORT=8080                      # Port for HTTP server

# Database configuration
SQLITE_DB_PATH=/var/lib/image-cleanup/cleanup.db  # SQLite database path

# Logger configuration
LOG_LEVEL=info                      # Log level (debug, info, warn, error)
LOG_DIR=/var/log/image-cleanup      # Directory for log files
LOG_MAX_SIZE=100                    # Maximum size of each log file in MB
LOG_MAX_BACKUPS=5                   # Number of old log files to keep
LOG_MAX_AGE=30                      # Days to keep old log files
LOG_COMPRESS=true                   # Compress old log files
```

4. Start and enable the service:

```bash
make enable
make start
```

## Service Management

### Basic Commands

```bash
make start         # Start the service
make stop          # Stop the service
make restart       # Restart the service
make status        # Check service status
make check         # Check service health
make enable        # Enable service autostart
make disable       # Disable service autostart
```

### Logs

```bash
make logs          # View service logs
make error-logs    # View error logs
make health-logs   # View health check logs
```

### Configuration

```bash
make edit-config   # Edit configuration
make show-config   # Show current configuration
make version       # Show build version
```

## API Endpoints

### Health Check

- Endpoint: `http://localhost:8080/health`
- Method: GET
- Response: Health status information
- Automatic check every 5 minutes
- Auto-restart on failure

### Cleanup Status

- Endpoint: `http://localhost:8080/api/v1/cleanup`
- Method: GET
- Response: Latest cleanup results including:
  - Host information
  - Start and end time
  - Duration
  - Total image count
  - Removed image count
  - Skipped image count

### Metrics

- Endpoint: `http://localhost:8080/metrics`
- Method: GET
- Response: Prometheus metrics including:
  - Total images cleaned
  - Cleanup duration
  - Error counts
  - Last run timestamp
  - Worker pool statistics

## Database Management

The service uses SQLite to store cleanup results. The database is located at:

```
/var/lib/image-cleanup/cleanup.db
```

To manually query the database:

```bash
sqlite3 /var/lib/image-cleanup/cleanup.db
```

Common SQLite commands:

```sql
-- View all cleanup results
SELECT * FROM cleanup_results ORDER BY start_time DESC;

-- View most recent cleanup
SELECT * FROM cleanup_results ORDER BY start_time DESC LIMIT 1;

-- Count cleanup runs by month
SELECT strftime('%Y-%m', start_time) as month, COUNT(*) as count
FROM cleanup_results
GROUP BY month
ORDER BY month DESC;

-- Get total images removed
SELECT SUM(removed) FROM cleanup_results;
```

## Log Management

### Log Files

All logs are automatically rotated based on size and age:

```bash
/var/log/image-cleanup/
├── service.log        # Main service logs
├── error.log         # Error logs
└── health.log        # Health check logs
```

### Log Configuration

- Size-based rotation: Files are rotated when they reach MAX_SIZE
- Age-based cleanup: Files older than MAX_AGE days are removed
- Compression: Old log files are automatically compressed
- Backup management: Keeps up to MAX_BACKUPS old files

## Uninstallation

Remove the service:

```bash
make uninstall
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## Development

### Project Structure

```
.
├── cmd/                        # Main application
├── config/                     # Configuration handling
├── internal/                   # Private application code
│   ├── domain/                 # Business logic interfaces
│   │   ├── models/             # Domain models
│   │   ├── notification/       # Notification interface
│   │   ├── repositories/       # Repository interfaces
│   │   └── metrics/            # Metrics interfaces
│   ├── infrastructure/         # External services implementation
│   │   ├── container/          # Container runtime implementation
│   │   ├── logger/             # Logging implementation
│   │   ├── metrics/            # Metrics collection (Prometheus)
│   │   ├── notification/       # Notification implementation
│   │   └── repositories/       # Repository implementations (SQLite)
│   ├── interfaces/             # Interface adapters
│   │   └── http/               # HTTP layer
│   │       ├── handlers/       # HTTP handlers
│   │       ├── middleware/     # HTTP middleware
│   │       └── router/         # Router setup
│   └── usecases/               # Application business rules
│       └── cleanup/            # Image cleanup implementation
├── pkg/                        # Public shared code
│   ├── helper/                 # Helper functions
│   └── constants/              # Global constants
├── scripts/                    # Build and installation scripts
├── build/                      # Build artifacts
├── Makefile                    # Build and management commands
├── go.mod                      # Go module definition
├── go.sum                      # Go module checksums
└── README.md                   # Project documentation
```

### Running Tests

```bash
# Run all tests
make test

# Run specific tests
go test ./internal/usecases/cleanup/...
```

## Troubleshooting

1. View service status:

```bash
make status
```

2. Check service logs:

```bash
make logs
```

3. Verify health endpoint:

```bash
make check
```

4. Check cleanup results API:

```bash
curl http://localhost:8080/api/v1/cleanup | jq .
```

5. Examine the database directly:

```bash
sqlite3 /var/lib/image-cleanup/cleanup.db "SELECT * FROM cleanup_results ORDER BY start_time DESC LIMIT 1;"
```

6. Common issues:

- Permission denied: Make sure to run installation as root
- Service won't start: Check error logs with `make error-logs`
- Health check fails: Verify port 8080 is available
- Database errors: Check if SQLite is installed properly
- API not accessible: Verify firewall settings allow access to port 8080

## License

[Add your license information here]
