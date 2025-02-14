# Image Cleanup Service

Automated service for cleaning up unused container images with Telegram notifications.

## Features

- Automated cleanup of unused container images
- Telegram notifications with cleanup results and host info
- Health monitoring with auto-recovery
- Prometheus metrics endpoint
- Systemd service integration
- Parallel processing with worker pool
- Comprehensive logging with automatic rotation
- ICT+7 timezone support

## Prerequisites

- Go 1.21 or higher
- Linux with systemd
- Root access for service installation
- crictl installed and configured
- Telegram bot token and chat ID

## Installation

1. Clone the repository:

```bash
git clone <repository-url>
cd image-cleanup
```

2. Build and install:

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
make start       # Start the service
make stop        # Stop the service
make restart     # Restart the service
make status      # Check service status
make check       # Check service health
make enable      # Enable service autostart
make disable     # Disable service autostart
```

### Logs

```bash
make logs        # View service logs
make error-logs  # View error logs
make health-logs # View health check logs
```

### Configuration

```bash
make edit-config  # Edit configuration
make show-config  # Show current configuration
```

## Development

### Build and Test

```bash
make build      # Build binary
make test       # Run tests with coverage
make clean      # Clean build artifacts
```

### Directory Structure

```
.
├── cmd/                # Main application
├── config/            # Configuration handling
├── internal/
│   ├── domain/        # Business logic interfaces
│   │   ├── models/    # Domain models
│   │   ├── notification/ # Notification interface
│   │   └── repositories/ # Repository interfaces
│   ├── infrastructure/# External services implementation
│   │   ├── container/ # Container runtime implementation
│   │   ├── logger/    # Logging implementation
│   │   └── notification/ # Notification implementation
│   ├── interfaces/    # HTTP handlers
│   └── usecases/      # Business logic implementation
├── scripts/           # Installation and maintenance scripts
└── Makefile          # Build and management commands
```

## Monitoring

### Health Check

- Endpoint: `http://localhost:8080/health`
- Automatic check every 5 minutes
- Auto-restart on failure
- Health status logging

### Metrics

Endpoint: `http://localhost:8080/metrics`

Prometheus metrics include:

- Total images cleaned
- Cleanup duration
- Error counts
- Last run timestamp
- Number of images skipped
- Worker pool statistics

## Log Management

### Log Files

- Service logs: `/var/log/image-cleanup/service.log`
- Error logs: `/var/log/image-cleanup/error.log`
- Health check logs: `/var/log/image-cleanup/health.log`

### Log Rotation

- Automatic rotation based on file size
- Compression of old log files
- Age-based cleanup
- Backup retention management

### Log Configuration

- `LOG_LEVEL`: Set logging detail level (debug, info, warn, error)
- `LOG_MAX_SIZE`: Maximum size in MB before rotating (default: 100MB)
- `LOG_MAX_BACKUPS`: Number of old log files to keep (default: 5)
- `LOG_MAX_AGE`: Days to keep old log files (default: 30)
- `LOG_COMPRESS`: Whether to compress old log files (default: true)

### Viewing Logs

```bash
# View active logs
make logs

# View specific log files
tail -f /var/log/image-cleanup/service.log
tail -f /var/log/image-cleanup/error.log

# List rotated log files
ls -l /var/log/image-cleanup/
```

## Notification Format

Notifications sent to Telegram include:

- Host information (hostname and IPs)
- Start time (ICT+7)
- End time (ICT+7)
- Duration
- Total images processed
- Number of images removed
- Number of images skipped

Example:

```
Image cleanup completed on:
Host: server-name
IP(s): 192.168.1.100, 10.0.0.50

Started: 2025-02-14 14:30:00 ICT
Finished: 2025-02-14 14:31:05 ICT
Duration: 1m5s

Results:
- Total: 10
- Removed: 5
- Skipped: 5
```

## Uninstallation

To remove the service:

```bash
make uninstall
```

The uninstall process will:

1. Stop all services
2. Remove systemd service files
3. Remove binaries
4. Optionally remove configuration and log files

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

[Add your license information here]
