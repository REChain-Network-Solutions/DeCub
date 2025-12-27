#!/bin/bash
set -e

# DeCube Load Testing Script
# Tests API endpoints under load

ENDPOINT="${ENDPOINT:-http://localhost:8080}"
CONCURRENT="${CONCURRENT:-100}"
REQUESTS="${REQUESTS:-10000}"
DURATION="${DURATION:-60s}"

echo "DeCube Load Test"
echo "================"
echo "Endpoint: $ENDPOINT"
echo "Concurrent: $CONCURRENT"
echo "Requests: $REQUESTS"
echo "Duration: $DURATION"
echo ""

# Check if wrk is installed
if ! command -v wrk >/dev/null 2>&1; then
    echo "Error: wrk is not installed"
    echo "Install with: brew install wrk (macOS) or apt-get install wrk (Linux)"
    exit 1
fi

# Health check
echo "Checking service health..."
if ! curl -sf "$ENDPOINT/health" > /dev/null; then
    echo "Error: Service is not healthy"
    exit 1
fi
echo "✓ Service is healthy"
echo ""

# Test endpoints
ENDPOINTS=(
    "/health"
    "/catalog/snapshots"
    "/gossip/status"
)

for endpoint in "${ENDPOINTS[@]}"; do
    echo "Testing $endpoint..."
    wrk -t4 -c$CONCURRENT -d$DURATION "$ENDPOINT$endpoint" | tee "load-test-$(basename $endpoint).txt"
    echo ""
done

echo "Load test complete!"
echo "Results saved to load-test-*.txt"

