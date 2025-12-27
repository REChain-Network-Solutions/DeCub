#!/bin/bash
set -e

# DeCube Backup Verification Script

BACKUP_DIR="${BACKUP_DIR:-/backup/decube}"
BACKUP_REF="${1:-latest}"

echo "DeCube Backup Verification"
echo "========================="
echo "Backup: $BACKUP_REF"
echo ""

# Find backup file
if [ "$BACKUP_REF" = "latest" ]; then
    BACKUP_FILE=$(ls -t "$BACKUP_DIR"/data-*.tar.gz 2>/dev/null | head -1)
elif [ -f "$BACKUP_REF" ]; then
    BACKUP_FILE="$BACKUP_REF"
elif [ -f "$BACKUP_DIR/data-$BACKUP_REF.tar.gz" ]; then
    BACKUP_FILE="$BACKUP_DIR/data-$BACKUP_REF.tar.gz"
else
    echo "Error: Backup file not found: $BACKUP_REF"
    exit 1
fi

if [ ! -f "$BACKUP_FILE" ]; then
    echo "Error: Backup file not found: $BACKUP_FILE"
    exit 1
fi

echo "Found backup: $BACKUP_FILE"
echo ""

# Check file integrity
echo "Checking file integrity..."
if gzip -t "$BACKUP_FILE" 2>/dev/null; then
    echo "✓ Backup file is valid"
else
    echo "✗ Backup file is corrupted"
    exit 1
fi

# Check checksum if available
CHECKSUM_FILE="$BACKUP_DIR/checksums-$(basename $BACKUP_FILE | cut -d- -f2- | sed 's/.tar.gz//').txt"
if [ -f "$CHECKSUM_FILE" ]; then
    echo "Verifying checksum..."
    if sha256sum -c "$CHECKSUM_FILE" > /dev/null 2>&1; then
        echo "✓ Checksum verified"
    else
        echo "⚠ Checksum verification failed"
    fi
fi

# Check backup contents
echo "Checking backup contents..."
TEMP_DIR=$(mktemp -d)
tar -tzf "$BACKUP_FILE" > "$TEMP_DIR/contents.txt" 2>/dev/null || {
    echo "✗ Cannot read backup contents"
    rm -rf "$TEMP_DIR"
    exit 1
}

FILE_COUNT=$(wc -l < "$TEMP_DIR/contents.txt")
echo "✓ Backup contains $FILE_COUNT files"

# Check for critical files
CRITICAL_FILES=(
    "var/lib/decube/raft"
    "var/lib/decube/catalog"
)

MISSING=0
for file in "${CRITICAL_FILES[@]}"; do
    if grep -q "$file" "$TEMP_DIR/contents.txt"; then
        echo "✓ Found: $file"
    else
        echo "⚠ Missing: $file"
        MISSING=$((MISSING + 1))
    fi
done

rm -rf "$TEMP_DIR"

if [ $MISSING -gt 0 ]; then
    echo ""
    echo "⚠ Warning: Some critical files are missing"
    exit 1
fi

echo ""
echo "✓ Backup verification complete!"
echo "Backup is valid and ready for restore"

