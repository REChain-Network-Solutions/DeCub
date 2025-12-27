#!/bin/bash
set -e

echo "Setting up DeCube development environment..."

# Check prerequisites
echo "Checking prerequisites..."
command -v go >/dev/null 2>&1 || { echo "Go is required but not installed. Aborting." >&2; exit 1; }
command -v docker >/dev/null 2>&1 || { echo "Docker is required but not installed. Aborting." >&2; exit 1; }
command -v docker-compose >/dev/null 2>&1 || { echo "Docker Compose is required but not installed. Aborting." >&2; exit 1; }

# Check Go version
GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
REQUIRED_VERSION="1.19"
if [ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]; then
    echo "Go version $GO_VERSION is installed, but $REQUIRED_VERSION or higher is required."
    exit 1
fi

echo "✓ Go version check passed"

# Install development tools
echo "Installing development tools..."
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest || true
go install github.com/goreleaser/goreleaser@latest || true

echo "✓ Development tools installed"

# Download dependencies
echo "Downloading dependencies..."
for dir in decub-control-plane decub-gcl/go decub-gossip decub-cas decub-catalog decub-snapshot decub-object-storage rechain decube; do
    if [ -f "$dir/go.mod" ]; then
        echo "  Downloading dependencies for $dir..."
        cd "$dir" && go mod download && cd - > /dev/null
    fi
done

echo "✓ Dependencies downloaded"

# Create necessary directories
echo "Creating directories..."
mkdir -p bin
mkdir -p logs
mkdir -p data

echo "✓ Directories created"

# Build all components
echo "Building components..."
make build || echo "⚠ Some components failed to build (this may be expected)"

echo ""
echo "Development environment setup complete!"
echo ""
echo "Next steps:"
echo "  1. Review configuration files in config/"
echo "  2. Start services: docker-compose up -d"
echo "  3. Run tests: make test"
echo "  4. Check documentation: docs/"

