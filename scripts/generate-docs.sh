#!/bin/bash
set -e

echo "Generating DeCube documentation..."

# Generate API documentation from Go code
echo "Generating API documentation..."
command -v godoc >/dev/null 2>&1 || { echo "godoc not found. Install with: go install golang.org/x/tools/cmd/godoc@latest"; exit 1; }

# Generate docs for each component
for dir in decub-control-plane decub-gcl/go decub-gossip decub-cas decub-catalog decub-snapshot decub-object-storage rechain decube; do
    if [ -d "$dir" ]; then
        echo "  Generating docs for $dir..."
        mkdir -p "docs/api/$dir"
        godoc -all "$dir" > "docs/api/$dir/index.html" 2>/dev/null || echo "    ⚠ Could not generate docs for $dir"
    fi
done

# Generate protocol buffer documentation
if command -v protoc-gen-doc >/dev/null 2>&1; then
    echo "Generating protocol buffer documentation..."
    find . -name "*.proto" | while read -r file; do
        echo "  Processing $file..."
        protoc --doc_out=docs/api/proto --doc_opt=markdown,"$(basename "$file" .proto).md" "$file" || true
    done
fi

echo "✓ Documentation generation complete"

