# Backup and Recovery Guide

This guide covers backup and recovery procedures for DeCube.

## Table of Contents

1. [Backup Strategies](#backup-strategies)
2. [Backup Procedures](#backup-procedures)
3. [Recovery Procedures](#recovery-procedures)
4. [Disaster Recovery](#disaster-recovery)
5. [Best Practices](#best-practices)

## Backup Strategies

### Backup Types

#### Full Backup
- Complete system state
- All data and configuration
- Requires most storage
- Longest recovery time

#### Incremental Backup
- Only changed data since last backup
- Faster backup process
- Requires full backup + increments for recovery

#### Snapshot Backup
- Point-in-time state
- Fast creation and recovery
- Storage-efficient

### Backup Frequency

- **Production**: Daily full backups, hourly incremental
- **Development**: Weekly full backups
- **Critical Systems**: Continuous replication + daily backups

## Backup Procedures

### Data Backup

#### Manual Backup

```bash
# Backup data directory
tar -czf decube-data-$(date +%Y%m%d).tar.gz /var/lib/decube/

# Backup configuration
cp config/config.yaml config/config.yaml.backup-$(date +%Y%m%d)

# Backup certificates
tar -czf decube-certs-$(date +%Y%m%d).tar.gz /etc/decube/tls/
```

#### Automated Backup Script

```bash
#!/bin/bash
# scripts/backup.sh

BACKUP_DIR="/backup/decube"
DATE=$(date +%Y%m%d-%H%M%S)

# Create backup directory
mkdir -p $BACKUP_DIR

# Backup data
tar -czf $BACKUP_DIR/data-$DATE.tar.gz /var/lib/decube/

# Backup configuration
cp config/config.yaml $BACKUP_DIR/config-$DATE.yaml

# Backup snapshots
./bin/decub-snapshot export-all > $BACKUP_DIR/snapshots-$DATE.tar.gz

# Upload to remote storage (S3, etc.)
aws s3 cp $BACKUP_DIR s3://decube-backups/ --recursive

# Cleanup old backups (keep last 30 days)
find $BACKUP_DIR -type f -mtime +30 -delete
```

### Snapshot Backup

#### Export Snapshots

```bash
# Export single snapshot
./bin/decub-snapshot export snapshot-001 > snapshot-001.tar.gz

# Export all snapshots
./bin/decub-snapshot export-all > all-snapshots.tar.gz

# Export with metadata
./bin/decub-snapshot export snapshot-001 --include-metadata > snapshot-001-full.tar.gz
```

#### Catalog Backup

```bash
# Export catalog
curl http://localhost:8080/catalog/export > catalog-$(date +%Y%m%d).json

# Backup catalog database
# (depends on storage backend)
```

### Configuration Backup

```bash
# Backup all configuration files
tar -czf config-backup-$(date +%Y%m%d).tar.gz config/

# Backup environment variables
env | grep DECUBE > env-backup-$(date +%Y%m%d).txt
```

### Certificate Backup

```bash
# Backup TLS certificates
tar -czf certs-backup-$(date +%Y%m%d).tar.gz /etc/decube/tls/

# Backup keys (encrypted)
gpg --encrypt --recipient backup@example.com \
  /etc/decube/tls/key.pem > key.pem.gpg
```

## Recovery Procedures

### Full System Recovery

#### From Backup Archive

```bash
# Stop services
docker-compose down

# Restore data
tar -xzf decube-data-YYYYMMDD.tar.gz -C /

# Restore configuration
cp config/config.yaml.backup-YYYYMMDD config/config.yaml

# Restore certificates
tar -xzf decube-certs-YYYYMMDD.tar.gz -C /

# Start services
docker-compose up -d

# Verify recovery
./scripts/health-check.sh
```

#### From Snapshot

```bash
# Import snapshot
./bin/decub-snapshot import snapshot-001.tar.gz

# Verify import
curl http://localhost:8080/catalog/snapshots/snapshot-001
```

### Partial Recovery

#### Restore Specific Snapshot

```bash
# Import specific snapshot
./bin/decub-snapshot import snapshot-001.tar.gz

# Restore to specific location
./bin/decub-snapshot restore snapshot-001 /data/restore
```

#### Restore Catalog Entry

```bash
# Import catalog entry
curl -X POST http://localhost:8080/catalog/import \
  -H "Content-Type: application/json" \
  -d @catalog-entry.json
```

### Point-in-Time Recovery

```bash
# List available snapshots
curl http://localhost:8080/catalog/snapshots

# Restore to specific point in time
./bin/decub-snapshot restore snapshot-001 \
  --timestamp "2024-01-15T10:30:00Z" \
  /data/restore
```

## Disaster Recovery

### Recovery Scenarios

#### Complete Cluster Failure

1. **Assess Damage**
   ```bash
   # Check what's available
   ls -la /var/lib/decube/
   ```

2. **Restore from Backup**
   ```bash
   # Restore latest backup
   ./scripts/restore.sh --backup=latest
   ```

3. **Verify Integrity**
   ```bash
   # Check data integrity
   ./bin/decube-verify --data-dir=/var/lib/decube/
   ```

4. **Start Services**
   ```bash
   docker-compose up -d
   ```

#### Data Corruption

1. **Identify Corrupted Data**
   ```bash
   ./bin/decube-verify --check-integrity
   ```

2. **Restore from Backup**
   ```bash
   # Restore specific corrupted component
   ./scripts/restore.sh --component=catalog --backup=latest
   ```

3. **Verify Recovery**
   ```bash
   ./bin/decube-verify --data-dir=/var/lib/decube/
   ```

#### Network Partition Recovery

1. **Reconnect Clusters**
   ```bash
   # Update cluster configuration
   # Restart services
   docker-compose restart
   ```

2. **Sync State**
   ```bash
   # Trigger anti-entropy
   curl -X POST http://localhost:8000/gossip/sync
   ```

3. **Verify Consistency**
   ```bash
   # Check Merkle roots
   curl http://localhost:8080/catalog/verify
   ```

### Recovery Testing

#### Test Recovery Procedure

```bash
#!/bin/bash
# scripts/test-recovery.sh

# Create test backup
./scripts/backup.sh

# Simulate failure
docker-compose down
rm -rf /var/lib/decube/data/*

# Restore
./scripts/restore.sh --backup=latest

# Verify
./scripts/health-check.sh
./bin/decube-verify
```

#### Regular Testing

- Test backups monthly
- Test recovery procedures quarterly
- Document any issues
- Update procedures as needed

## Best Practices

### Backup Best Practices

1. **Automate Backups**
   - Use cron jobs or scheduled tasks
   - Monitor backup success
   - Alert on failures

2. **Store Offsite**
   - Use cloud storage (S3, GCS, Azure)
   - Multiple geographic locations
   - Encrypt sensitive data

3. **Test Regularly**
   - Verify backup integrity
   - Test recovery procedures
   - Document results

4. **Version Backups**
   - Keep multiple backup versions
   - Implement retention policies
   - Label backups clearly

### Recovery Best Practices

1. **Document Procedures**
   - Step-by-step recovery guides
   - Contact information
   - Escalation procedures

2. **Practice Recovery**
   - Regular recovery drills
   - Time recovery procedures
   - Identify bottlenecks

3. **Monitor Health**
   - Continuous monitoring
   - Alert on anomalies
   - Regular health checks

4. **Maintain Runbooks**
   - Keep procedures current
   - Include troubleshooting
   - Update based on experience

### Security Best Practices

1. **Encrypt Backups**
   ```bash
   # Encrypt backup
   gpg --encrypt --recipient backup@example.com \
     decube-backup.tar.gz
   ```

2. **Secure Storage**
   - Use encrypted storage
   - Restrict access
   - Audit access logs

3. **Protect Keys**
   - Encrypt private keys
   - Use key management systems
   - Rotate keys regularly

## Backup Retention Policy

### Recommended Retention

- **Daily Backups**: 30 days
- **Weekly Backups**: 12 weeks
- **Monthly Backups**: 12 months
- **Yearly Backups**: 7 years

### Cleanup Script

```bash
#!/bin/bash
# scripts/cleanup-backups.sh

BACKUP_DIR="/backup/decube"

# Remove backups older than retention period
find $BACKUP_DIR -name "daily-*" -mtime +30 -delete
find $BACKUP_DIR -name "weekly-*" -mtime +84 -delete
find $BACKUP_DIR -name "monthly-*" -mtime +365 -delete
```

## Monitoring Backups

### Backup Monitoring

```yaml
# Monitoring configuration
backup:
  monitoring:
    enabled: true
    check_interval: "1h"
    alert_on_failure: true
    metrics_endpoint: "http://prometheus:9090"
```

### Health Checks

```bash
# Check backup status
./scripts/check-backups.sh

# Verify backup integrity
./bin/decube-verify --backup=/backup/decube/latest.tar.gz
```

## Troubleshooting

### Backup Failures

- Check disk space
- Verify permissions
- Review logs
- Test network connectivity

### Recovery Issues

- Verify backup integrity
- Check compatibility
- Review error messages
- Test in staging first

## References

- [Deployment Guide](deployment.md)
- [Troubleshooting Guide](troubleshooting.md)
- [Security Guide](SECURITY.md)

---

*Last updated: January 2024*

