#!/bin/bash
set -e

# DeCube Restore Script
# Restores from backup

if [ $# -lt 1 ]; then
    echo "Usage: $0 <backup-file|backup-date> [options]"
    echo ""
    echo "Options:"
    echo "  --component=<component>  Restore specific component"
    echo "  --data-dir=<dir>         Data directory (default: /var/lib/decube)"
    echo "  --dry-run                Show what would be restored without restoring"
    echo ""
    echo "Examples:"
    echo "  $0 data-20240115-120000.tar.gz"
    echo "  $0 20240115-120000"
    echo "  $0 latest"
    exit 1
fi

BACKUP_DIR="${BACKUP_DIR:-/backup/decube}"
BACKUP_REF="$1"
DATA_DIR="${DATA_DIR:-/var/lib/decube}"
DRY_RUN=false
COMPONENT=""

# Parse arguments
for arg in "$@"; do
    case $arg in
        --component=*)
            COMPONENT="${arg#*=}"
            shift
            ;;
        --data-dir=*)
            DATA_DIR="${arg#*=}"
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
    esac
done

echo "DeCube Restore Script"
echo "===================="
echo "Backup reference: $BACKUP_REF"
echo "Data directory: $DATA_DIR"
echo "Dry run: $DRY_RUN"
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

# Verify backup integrity
if [ -f "${BACKUP_FILE%.tar.gz}.txt" ] || [ -f "$BACKUP_DIR/checksums-$(basename $BACKUP_FILE | cut -d- -f2- | sed 's/.tar.gz//').txt" ]; then
    echo "Verifying backup integrity..."
    # Check if checksum file exists and verify
    CHECKSUM_FILE="$BACKUP_DIR/checksums-$(basename $BACKUP_FILE | cut -d- -f2- | sed 's/.tar.gz//').txt"
    if [ -f "$CHECKSUM_FILE" ]; then
        if sha256sum -c "$CHECKSUM_FILE" > /dev/null 2>&1; then
            echo "✓ Backup integrity verified"
        else
            echo "⚠ Backup integrity check failed, continuing anyway"
        fi
    fi
fi

if [ "$DRY_RUN" = true ]; then
    echo ""
    echo "DRY RUN - Would restore:"
    echo "  Backup: $BACKUP_FILE"
    echo "  Target: $DATA_DIR"
    echo "  Component: ${COMPONENT:-all}"
    echo ""
    echo "To actually restore, run without --dry-run"
    exit 0
fi

# Confirm restore
read -p "This will overwrite existing data. Continue? (yes/no): " confirm
if [ "$confirm" != "yes" ]; then
    echo "Restore cancelled"
    exit 0
fi

# Stop services (if running)
if command -v docker-compose >/dev/null 2>&1 && [ -f "docker-compose.yml" ]; then
    echo "Stopping services..."
    docker-compose down || true
fi

# Create data directory if it doesn't exist
mkdir -p "$DATA_DIR"

# Restore data
echo "Restoring data from $BACKUP_FILE..."
tar -xzf "$BACKUP_FILE" -C /

# Restore configuration if available
CONFIG_BACKUP="$BACKUP_DIR/config-$(basename $BACKUP_FILE | cut -d- -f2- | sed 's/.tar.gz//').yaml"
if [ -f "$CONFIG_BACKUP" ]; then
    echo "Restoring configuration..."
    cp "$CONFIG_BACKUP" config/config.yaml
    echo "✓ Configuration restored"
fi

# Restore certificates if available
CERTS_BACKUP="$BACKUP_DIR/certs-$(basename $BACKUP_FILE | cut -d- -f2- | sed 's/.tar.gz//').tar.gz"
if [ -f "$CERTS_BACKUP" ]; then
    echo "Restoring certificates..."
    tar -xzf "$CERTS_BACKUP" -C /
    echo "✓ Certificates restored"
fi

echo ""
echo "Restore complete!"
echo ""
echo "Next steps:"
echo "  1. Review restored data"
echo "  2. Start services: docker-compose up -d"
echo "  3. Verify: ./scripts/health-check.sh"

