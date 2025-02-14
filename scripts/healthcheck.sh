#!/bin/bash

# Config
SERVICE_NAME="image-cleanup"
HEALTH_CHECK_URL="http://localhost:8080/health"
MAX_RETRIES=3
RETRY_INTERVAL=5
LOG_FILE="/var/log/${SERVICE_NAME}/health.log"

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
    response=$(curl -s -w "\n%{http_code}" $HEALTH_CHECK_URL)
    http_code=$(echo "$response" | tail -n1)
    content=$(echo "$response" | head -n1)

    if [ "$http_code" = "200" ]; then
        if echo "$content" | grep -q '"status":"healthy"'; then
            return 0
        fi
    fi
    return 1
}

# Main health check logic
log "Starting health check for $SERVICE_NAME"

# Check if service is running
if ! check_service_status; then
    log "ERROR: $SERVICE_NAME service is not running"
    systemctl restart $SERVICE_NAME
    log "Service restarted"
    exit 1
fi

# Check health endpoint with retries
retry_count=0
while [ $retry_count -lt $MAX_RETRIES ]; do
    if check_health_endpoint; then
        log "Health check passed"
        exit 0
    fi
    
    retry_count=$((retry_count + 1))
    log "Health check failed (attempt $retry_count/$MAX_RETRIES)"
    
    if [ $retry_count -lt $MAX_RETRIES ]; then
        sleep $RETRY_INTERVAL
    fi
done

# If we reach here, health check failed after all retries
log "ERROR: Health check failed after $MAX_RETRIES attempts"
systemctl restart $SERVICE_NAME
log "Service restarted"
exit 1