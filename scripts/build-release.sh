#!/bin/bash
set -e

VERSION=${1:-"dev"}
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo "Building DeCube Release"
echo "====================="
echo "Version: $VERSION"
echo "Build Date: $BUILD_DATE"
echo "Git Commit: $GIT_COMMIT"
echo ""

# Create build directory
mkdir -p dist
rm -rf dist/*

# Build flags
LDFLAGS="-X main.Version=$VERSION -X main.BuildDate=$BUILD_DATE -X main.GitCommit=$GIT_COMMIT"

# Build components
components=(
    "decub-control-plane:control-plane"
    "decub-gcl/go:gcl"
    "decub-gossip:gossip"
    "decub-cas:cas"
    "decub-catalog:catalog"
    "decub-snapshot:snapshot"
    "decub-object-storage:object-storage"
)

for component in "${components[@]}"; do
    IFS=':' read -r dir name <<< "$component"
    if [ -d "$dir" ]; then
        echo "Building $name..."
        cd "$dir"
        GOOS=linux GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "../dist/${name}-linux-amd64" ./...
        GOOS=darwin GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "../dist/${name}-darwin-amd64" ./...
        GOOS=darwin GOARCH=arm64 go build -ldflags "$LDFLAGS" -o "../dist/${name}-darwin-arm64" ./...
        GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS" -o "../dist/${name}-windows-amd64.exe" ./...
        cd - > /dev/null
        echo "  ✓ Built $name"
    fi
done

# Create checksums
echo ""
echo "Creating checksums..."
cd dist
sha256sum * > checksums.txt
cd ..

echo ""
echo "Build complete! Artifacts in dist/"

