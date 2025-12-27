# Migration Guide

This guide helps you migrate between different versions of DeCube or from other systems.

## Table of Contents

1. [Upgrading DeCube](#upgrading-decube)
2. [Migrating from Other Systems](#migrating-from-other-systems)
3. [Data Migration](#data-migration)
4. [Configuration Migration](#configuration-migration)
5. [Rollback Procedures](#rollback-procedures)

## Upgrading DeCube

### Version 0.1.x to 0.2.x

#### Breaking Changes
- Configuration file format updated
- API endpoint changes
- Storage format changes

#### Migration Steps

1. **Backup Current Installation**
   ```bash
   # Backup data
   tar -czf decube-backup-$(date +%Y%m%d).tar.gz /var/lib/decube/
   
   # Backup configuration
   cp config/config.yaml config/config.yaml.backup
   ```

2. **Review Release Notes**
   - Check [CHANGELOG.md](../CHANGELOG.md) for changes
   - Review [RELEASE_NOTES.md](RELEASE_NOTES.md) for version-specific notes

3. **Update Configuration**
   ```bash
   # Use migration script
   ./scripts/migrate-config.sh config/config.yaml
   
   # Or manually update based on config.example.yaml
   ```

4. **Upgrade Components**
   ```bash
   # Stop services
   docker-compose down
   
   # Pull new images
   docker-compose pull
   
   # Start with new version
   docker-compose up -d
   ```

5. **Verify Migration**
   ```bash
   # Check service health
   ./scripts/health-check.sh
   
   # Verify data integrity
   curl http://localhost:8080/catalog/health
   ```

### Version 0.0.x to 0.1.x

#### Major Changes
- New component structure
- Updated API endpoints
- Enhanced security features

#### Migration Steps

1. Export existing data
2. Update configuration files
3. Migrate data format
4. Import into new version
5. Verify functionality

## Migrating from Other Systems

### From etcd

#### Export Data
```bash
# Export etcd data
etcdctl snapshot save snapshot.db
```

#### Import to DeCube
```bash
# Use migration tool
./bin/decube-migrate --from=etcd --snapshot=snapshot.db
```

### From Consul

#### Export Data
```bash
# Export Consul KV store
consul kv export > consul-data.json
```

#### Import to DeCube
```bash
# Import Consul data
./bin/decube-migrate --from=consul --data=consul-data.json
```

### From Kubernetes etcd

#### Steps
1. Create etcd snapshot
2. Convert to DeCube format
3. Import into DeCube cluster
4. Verify data integrity

## Data Migration

### Snapshot Migration

#### Export Snapshots
```bash
# List all snapshots
curl http://old-cluster:8080/catalog/snapshots

# Export snapshot
./bin/decub-snapshot export snapshot-id > snapshot.tar.gz
```

#### Import Snapshots
```bash
# Import snapshot
./bin/decub-snapshot import snapshot.tar.gz

# Verify import
curl http://new-cluster:8080/catalog/snapshots/snapshot-id
```

### Catalog Migration

#### Export Catalog
```bash
# Export all catalog entries
curl http://old-cluster:8080/catalog/export > catalog.json
```

#### Import Catalog
```bash
# Import catalog
curl -X POST http://new-cluster:8080/catalog/import \
  -H "Content-Type: application/json" \
  -d @catalog.json
```

### Storage Migration

#### CAS Migration
```bash
# Sync CAS storage
./bin/decube-cas-sync \
  --source=http://old-cluster:9000 \
  --target=http://new-cluster:9000 \
  --bucket=decube-cas
```

## Configuration Migration

### Automated Migration

```bash
# Run migration script
./scripts/migrate-config.sh old-config.yaml new-config.yaml
```

### Manual Migration

1. **Compare Configurations**
   ```bash
   diff config/config.yaml config/config.example.yaml
   ```

2. **Update Section by Section**
   - Cluster configuration
   - Storage settings
   - Network settings
   - Security settings

3. **Validate Configuration**
   ```bash
   ./scripts/validate-config.sh
   ```

## Rollback Procedures

### Quick Rollback

```bash
# Stop current version
docker-compose down

# Restore previous version
git checkout v0.1.0
docker-compose up -d
```

### Data Rollback

```bash
# Restore from backup
tar -xzf decube-backup-YYYYMMDD.tar.gz -C /

# Restore configuration
cp config/config.yaml.backup config/config.yaml

# Restart services
docker-compose restart
```

### Configuration Rollback

```bash
# Restore configuration
cp config/config.yaml.backup config/config.yaml

# Reload configuration
curl -X POST http://localhost:8080/admin/reload
```

## Migration Best Practices

### Before Migration

1. **Backup Everything**
   - Data directories
   - Configuration files
   - Database snapshots
   - Certificates and keys

2. **Test in Staging**
   - Set up staging environment
   - Test migration process
   - Verify data integrity
   - Performance testing

3. **Plan Downtime**
   - Schedule maintenance window
   - Notify users
   - Prepare rollback plan

### During Migration

1. **Monitor Progress**
   - Watch migration logs
   - Monitor resource usage
   - Check for errors

2. **Verify Steps**
   - Validate each step
   - Test functionality
   - Check data integrity

### After Migration

1. **Verification**
   - Health checks
   - Data integrity checks
   - Performance validation
   - User acceptance testing

2. **Monitoring**
   - Watch for errors
   - Monitor performance
   - Check logs regularly

3. **Documentation**
   - Document any issues
   - Update runbooks
   - Share lessons learned

## Troubleshooting Migration Issues

### Common Issues

#### Data Loss
- **Symptom**: Missing data after migration
- **Solution**: Restore from backup, verify export/import process

#### Configuration Errors
- **Symptom**: Services won't start
- **Solution**: Validate configuration, check logs

#### Performance Degradation
- **Symptom**: Slower operations after migration
- **Solution**: Review resource allocation, check network

#### Compatibility Issues
- **Symptom**: Features not working
- **Solution**: Check version compatibility, review release notes

## Migration Tools

### Available Tools

- `decube-migrate`: General migration tool
- `decub-snapshot`: Snapshot import/export
- `decube-cas-sync`: CAS storage synchronization
- `migrate-config.sh`: Configuration migration script

### Getting Help

- Check [Troubleshooting Guide](troubleshooting.md)
- Review [FAQ](faq.md)
- Open an issue on GitHub
- Contact support

## Version Compatibility Matrix

| From Version | To Version | Migration Path | Breaking Changes |
|-------------|------------|----------------|------------------|
| 0.0.x | 0.1.x | Direct | Yes |
| 0.1.x | 0.2.x | Direct | Yes |
| 0.2.x | 0.3.x | Direct | Minor |

---

*Last updated: January 2024*

