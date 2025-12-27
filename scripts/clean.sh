#!/bin/bash
set -e

echo "Cleaning DeCube build artifacts..."

# Remove build artifacts
echo "Removing build artifacts..."
rm -rf bin/
rm -rf dist/
rm -rf logs/
rm -rf data/

# Clean Go cache
echo "Cleaning Go cache..."
go clean -cache -modcache -testcache || true

# Clean Docker
echo "Cleaning Docker..."
docker-compose down -v 2>/dev/null || true
docker system prune -f || true

# Remove test artifacts
echo "Removing test artifacts..."
find . -name "*.test" -type f -delete
find . -name "coverage.out" -type f -delete
find . -name "*.coverprofile" -type f -delete

echo "✓ Cleanup complete"

