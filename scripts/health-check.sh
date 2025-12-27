#!/bin/bash
set -e

echo "DeCube Health Check"
echo "=================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
CATALOG_ENDPOINT="${CATALOG_ENDPOINT:-http://localhost:8080}"
GOSSIP_ENDPOINT="${GOSSIP_ENDPOINT:-http://localhost:8000}"
CAS_ENDPOINT="${CAS_ENDPOINT:-http://localhost:9000}"

check_service() {
    local name=$1
    local endpoint=$2
    
    echo -n "Checking $name... "
    if curl -sf "$endpoint/health" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

check_docker() {
    echo -n "Checking Docker... "
    if command -v docker >/dev/null 2>&1 && docker ps >/dev/null 2>&1; then
        echo -e "${GREEN}✓ OK${NC}"
        
        echo "  Running containers:"
        docker ps --format "table {{.Names}}\t{{.Status}}" | grep -E "decube|rechain" || echo "    None"
        return 0
    else
        echo -e "${RED}✗ FAILED${NC}"
        return 1
    fi
}

check_ports() {
    echo "Checking ports..."
    for port in 8080 8000 7000 9000; do
        if netstat -an 2>/dev/null | grep -q ":$port " || ss -an 2>/dev/null | grep -q ":$port "; then
            echo -e "  Port $port: ${GREEN}✓ In use${NC}"
        else
            echo -e "  Port $port: ${YELLOW}⚠ Not in use${NC}"
        fi
    done
}

# Run checks
echo ""
check_docker
echo ""
check_service "Catalog" "$CATALOG_ENDPOINT"
check_service "Gossip" "$GOSSIP_ENDPOINT"
check_service "CAS" "$CAS_ENDPOINT"
echo ""
check_ports
echo ""

# Summary
echo "Health check complete!"

