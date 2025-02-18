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

# Function to check if service is running
check_service_status() {
    if systemctl is-active --quiet $SERVICE_NAME; then
        return 0
    else
        return 1
    fi
}

# Function to check health endpoint
check_health_endpoint() {
    # Add timeout to curl and more complete response checking
    response=$(curl -s -m $CURL_TIMEOUT -w "\n%{http_code}" $HEALTH_CHECK_URL)
    if [ $? -ne 0 ]; then
        log "curl failed to execute"
        return 1
    }

    http_code=$(echo "$response" | tail -n1)
    content=$(echo "$response" | head -n1)

    if [ "$http_code" = "200" ]; then
        if echo "$content" | grep -q '"status":"ok"'; then
            # Extract and log useful information
            uptime=$(echo "$content" | grep -o '"uptime":"[^"]*"' | cut -d'"' -f4)
            goroutines=$(echo "$content" | grep -o '"goroutines":[0-9]*' | cut -d':' -f2)
            log "Health check details - Uptime: $uptime, Goroutines: $goroutines"
            return 0
        fi
    fi

    # Log failure details
    log "Failed health check - HTTP Code: $http_code, Response: $content"
    return 1
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

# Check health endpoint with retries
retry_count=0
while [ $retry_count -lt $MAX_RETRIES ]; do
    if check_health_endpoint; then
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
if check_health_endpoint; then
    log "Service recovered after restart"
    exit 0
else
    log "ERROR: Service remains unhealthy after restart"
    exit 1
fi
