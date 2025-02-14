# Image Cleanup Service

Automated service for cleaning up unused container images with Telegram notifications.

## Features

- Automated cleanup of unused container images
- Telegram notifications with cleanup results
- Health monitoring with auto-recovery
- Prometheus metrics endpoint
- Systemd service integration
- Parallel processing with worker pool
- Comprehensive logging

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
TELEGRAM_BOT_TOKEN=your_bot_token
TELEGRAM_CHAT_ID=your_chat_id
CLEANUP_SCHEDULE="0 0 * * *"  # Cron schedule
LOG_LEVEL=info               # debug, info, warn, error
HTTP_PORT=8080
```

4. Start and enable the service:

```bash
make enable
make start
```

## Service Management

### Basic Commands

```bash
make start      # Start the service
make stop       # Stop the service
make restart    # Restart the service
make status     # Check service status
make check      # Check service health
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
make test       # Run tests
make clean      # Clean build artifacts
```

### Directory Structure

```
.
├── cmd/                # Main application
├── config/            # Configuration handling
├── internal/
│   ├── domain/        # Business logic interfaces
│   ├── infrastructure/# External services implementation
│   ├── interfaces/    # HTTP handlers
│   └── usecases/      # Business logic implementation
├── scripts/           # Installation scripts
└── Makefile          # Build and management commands
```

## Monitoring

### Health Check

- Endpoint: `http://localhost:8080/health`
- Automatic check every 5 minutes
- Auto-restart on failure

### Metrics

- Endpoint: `http://localhost:8080/metrics`
- Prometheus compatible
- Metrics include:
  - Total images cleaned
  - Cleanup duration
  - Error counts
  - Last run timestamp

## Logs

Logs are stored in the following locations:

- Service logs: `/var/log/image-cleanup/service.log`
- Error logs: `/var/log/image-cleanup/error.log`
- Health check logs: `/var/log/image-cleanup/health.log`

View logs using systemd:

```bash
make logs
```

## Uninstallation

To remove the service:

```bash
make uninstall
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

[Add your license information here]
