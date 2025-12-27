#!/bin/bash
set -e

echo "Validating DeCube configuration files..."

# Check YAML syntax
command -v yamllint >/dev/null 2>&1 || { echo "yamllint not found. Install with: pip install yamllint"; exit 1; }

# Validate all YAML files
echo "Validating YAML files..."
find . -name "*.yaml" -o -name "*.yml" | while read -r file; do
    if [[ "$file" != *"vendor"* ]] && [[ "$file" != *"node_modules"* ]]; then
        echo "  Checking $file..."
        yamllint "$file" || echo "    ⚠ Issues found in $file"
    fi
done

echo "✓ Configuration validation complete"

