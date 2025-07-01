#!/bin/sh

# Health check script for Torq service
# This script checks both liveness and readiness endpoints

set -e

# Accept port as first argument, fallback to env or default
PORT=${1:-${PORT:-8080}}
HOST=${HOST:-localhost}
TIMEOUT=${TIMEOUT:-3}

# Function to check endpoint
check_endpoint() {
    local endpoint=$1
    local description=$2

    echo "Checking $description endpoint..."

    if wget --no-verbose --tries=1 --timeout=$TIMEOUT --spider "http://$HOST:$PORT$endpoint" 2>/dev/null; then
        echo "✓ $description endpoint is healthy"
        return 0
    else
        echo "✗ $description endpoint is unhealthy"
        return 1
    fi
}

# Check liveness endpoint
if ! check_endpoint "/health/live" "liveness"; then
    exit 1
fi

# Check readiness endpoint
if ! check_endpoint "/health/ready" "readiness"; then
    exit 1
fi

echo "All health checks passed!"
exit 0