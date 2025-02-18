#!/bin/bash

# Config
SERVICE_NAME="image-cleanup"
HEALTH_CHECK_URL="http://localhost:8080/health"
MAX_RETRIES=3
RETRY_INTERVAL=5
LOG_FILE="/var/log/${SERVICE_NAME}/health.log"
CURL_TIMEOUT=10  # Added timeout for curl

# Function for logging
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Create log directory if it doesn't exist
mkdir -p "/var/log/${SERVICE_NAME}"

# Function to do HTTP health check
do_health_check() {
    log "Making HTTP request to: $HEALTH_CHECK_URL"
    response=$(curl -s -m $CURL_TIMEOUT -w "\n%{http_code}" $HEALTH_CHECK_URL)
    curl_status=$?

    if [ $curl_status -ne 0 ]; then
        log "Curl failed with status: $curl_status"
        return 1
    fi

    http_code=$(echo "$response" | tail -n1)
    content=$(echo "$response" | head -n1)

    log "HTTP response code: $http_code"
    log "Response content: $content"

    if [ "$http_code" = "200" ]; then
        if echo "$content" | grep -q '"status":"ok"'; then
            # Extract and log useful information
            uptime=$(echo "$content" | grep -o '"uptime":"[^"]*"' | cut -d'"' -f4)
            goroutines=$(echo "$content" | grep -o '"goroutines":[0-9]*' | cut -d':' -f2)
            log "Health check passed - Uptime: $uptime, Goroutines: $goroutines"
            return 0
        else
            log "Response status is not 'ok'"
            return 1
        fi
    else
        log "Unexpected HTTP status code: $http_code"
        return 1
    fi
}

# Function to check if service is running
check_service_status() {
    if systemctl is-active --quiet $SERVICE_NAME; then
        log "Service is running"
        return 0
    else
        log "Service is not running"
        return 1
    fi
}

# Main health check logic
log "Starting health check for $SERVICE_NAME"

# Check if service is running
if ! check_service_status; then
    log "ERROR: $SERVICE_NAME service is not running"
    log "Attempting to restart service..."
    systemctl restart $SERVICE_NAME
    log "Service restart initiated"

    # Wait for service to start
    sleep 5

    if check_service_status; then
        log "Service successfully restarted"
    else
        log "ERROR: Service failed to restart"
        exit 1
    fi
fi

# Do health checks with retries
retry_count=0
while [ $retry_count -lt $MAX_RETRIES ]; do
    do_health_check
    if [ $? -eq 0 ]; then
        log "Health check passed successfully"
        exit 0
    fi

    retry_count=$((retry_count + 1))
    log "Health check failed (attempt $retry_count/$MAX_RETRIES)"

    if [ $retry_count -lt $MAX_RETRIES ]; then
        log "Waiting $RETRY_INTERVAL seconds before next attempt..."
        sleep $RETRY_INTERVAL
    fi
done

# If we reach here, health check failed after all retries
log "ERROR: Health check failed after $MAX_RETRIES attempts"
log "Attempting to restart service..."
systemctl restart $SERVICE_NAME
log "Service restart initiated"

# Final check after restart
sleep 5
if do_health_check; then
    log "Service recovered after restart"
    exit 0
else
    log "ERROR: Service remains unhealthy after restart"
    exit 1
fi
