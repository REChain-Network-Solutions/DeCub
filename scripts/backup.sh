#!/bin/bash
set -e

# DeCube Backup Script
# Backs up data, configuration, and snapshots

BACKUP_DIR="${BACKUP_DIR:-/backup/decube}"
DATE=$(date +%Y%m%d-%H%M%S)
RETENTION_DAYS="${RETENTION_DAYS:-30}"

echo "Starting DeCube backup..."
echo "Backup directory: $BACKUP_DIR"
echo "Date: $DATE"

# Create backup directory
mkdir -p "$BACKUP_DIR"

# Backup data directory
if [ -d "/var/lib/decube" ]; then
    echo "Backing up data directory..."
    tar -czf "$BACKUP_DIR/data-$DATE.tar.gz" /var/lib/decube/
    echo "✓ Data backup complete"
else
    echo "⚠ Data directory not found, skipping"
fi

# Backup configuration
if [ -f "config/config.yaml" ]; then
    echo "Backing up configuration..."
    cp config/config.yaml "$BACKUP_DIR/config-$DATE.yaml"
    echo "✓ Configuration backup complete"
else
    echo "⚠ Configuration file not found, skipping"
fi

# Backup snapshots (if service is running)
if command -v curl >/dev/null 2>&1; then
    echo "Exporting snapshots..."
    if curl -sf http://localhost:8080/health > /dev/null; then
        curl -s http://localhost:8080/catalog/snapshots > "$BACKUP_DIR/snapshots-$DATE.json" || true
        echo "✓ Snapshots exported"
    else
        echo "⚠ Service not available, skipping snapshot export"
    fi
fi

# Backup certificates (if they exist)
if [ -d "/etc/decube/tls" ]; then
    echo "Backing up certificates..."
    tar -czf "$BACKUP_DIR/certs-$DATE.tar.gz" /etc/decube/tls/ 2>/dev/null || true
    echo "✓ Certificates backed up"
fi

# Create checksums
echo "Creating checksums..."
cd "$BACKUP_DIR"
sha256sum *-$DATE* > "checksums-$DATE.txt" 2>/dev/null || true

# Cleanup old backups
echo "Cleaning up backups older than $RETENTION_DAYS days..."
find "$BACKUP_DIR" -type f -mtime +$RETENTION_DAYS -delete
echo "✓ Cleanup complete"

echo ""
echo "Backup complete!"
echo "Backup location: $BACKUP_DIR"
echo "Files created:"
ls -lh "$BACKUP_DIR"/*-$DATE* 2>/dev/null || echo "No files created"

