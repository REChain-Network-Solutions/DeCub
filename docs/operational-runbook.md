# Operational Runbook

This runbook provides step-by-step procedures for common operational tasks.

## Table of Contents

1. [Daily Operations](#daily-operations)
2. [Weekly Operations](#weekly-operations)
3. [Monthly Operations](#monthly-operations)
4. [Incident Response](#incident-response)
5. [Maintenance Windows](#maintenance-windows)

## Daily Operations

### Health Checks

```bash
# Run health check script
./scripts/health-check.sh

# Check service status
docker-compose ps

# Check logs for errors
docker-compose logs --tail=100 | grep -i error
```

### Monitoring Review

1. Check dashboard metrics
2. Review alert status
3. Check error rates
4. Review resource usage

### Backup Verification

```bash
# Verify latest backup
ls -lh /backup/decube/

# Check backup integrity
./scripts/verify-backup.sh latest
```

## Weekly Operations

### Performance Review

1. Review performance metrics
2. Check for performance degradation
3. Review slow queries/operations
4. Analyze resource trends

### Log Review

```bash
# Review error logs
grep -i error /var/log/decube/decube.log | tail -100

# Review access logs
grep -i "unauthorized\|forbidden" /var/log/decube/decube.log
```

### Capacity Planning

1. Review storage usage
2. Check disk space
3. Review network usage
4. Plan for growth

## Monthly Operations

### Security Review

1. Review security logs
2. Check for suspicious activity
3. Review access patterns
4. Update security policies

### Backup Testing

```bash
# Test restore procedure
./scripts/test-restore.sh latest
```

### Documentation Update

1. Update runbooks
2. Document new procedures
3. Update contact information
4. Review and update playbooks

## Incident Response

### Service Down

1. **Identify Issue**
   ```bash
   ./scripts/health-check.sh
   docker-compose ps
   docker-compose logs
   ```

2. **Check Dependencies**
   - Storage services
   - Network connectivity
   - Resource availability

3. **Restart Services**
   ```bash
   docker-compose restart
   ```

4. **Verify Recovery**
   ```bash
   ./scripts/health-check.sh
   ```

### Data Corruption

1. **Identify Corrupted Data**
   ```bash
   ./bin/decube-verify --check-integrity
   ```

2. **Stop Services**
   ```bash
   docker-compose down
   ```

3. **Restore from Backup**
   ```bash
   ./scripts/restore.sh latest
   ```

4. **Verify and Restart**
   ```bash
   ./bin/decube-verify
   docker-compose up -d
   ```

### Performance Degradation

1. **Identify Bottleneck**
   - Check CPU usage
   - Check memory usage
   - Check disk I/O
   - Check network

2. **Review Metrics**
   - Request rates
   - Response times
   - Error rates

3. **Scale if Needed**
   ```bash
   docker-compose scale decube=5
   ```

### Security Incident

1. **Isolate Affected Systems**
   ```bash
   docker-compose stop
   ```

2. **Preserve Evidence**
   - Save logs
   - Take snapshots
   - Document timeline

3. **Notify Security Team**
   - Contact security@decube.io
   - Provide incident details

4. **Remediate**
   - Apply patches
   - Rotate credentials
   - Review access

## Maintenance Windows

### Planned Maintenance

1. **Notify Users**
   - Send maintenance notice
   - Schedule maintenance window

2. **Prepare**
   - Create backup
   - Review change plan
   - Prepare rollback

3. **Execute**
   - Apply changes
   - Verify functionality
   - Monitor closely

4. **Complete**
   - Verify services
   - Notify completion
   - Document changes

### Upgrade Procedure

1. **Pre-Upgrade**
   ```bash
   # Backup
   ./scripts/backup.sh
   
   # Review release notes
   cat CHANGELOG.md
   ```

2. **Upgrade**
   ```bash
   # Pull new images
   docker-compose pull
   
   # Stop services
   docker-compose down
   
   # Start with new version
   docker-compose up -d
   ```

3. **Post-Upgrade**
   ```bash
   # Verify
   ./scripts/health-check.sh
   
   # Check functionality
   curl http://localhost:8080/health
   ```

## Common Tasks

### Add New Node

1. Install DeCube
2. Configure cluster membership
3. Join to cluster
4. Verify connectivity

### Remove Node

1. Drain node
2. Remove from cluster
3. Update configuration
4. Verify cluster health

### Scale Services

```bash
# Scale up
docker-compose scale decube=5

# Scale down
docker-compose scale decube=3
```

### Update Configuration

1. Edit configuration
2. Validate configuration
3. Reload services
4. Verify changes

## Escalation Procedures

### Level 1: On-Call Engineer
- Initial response
- Basic troubleshooting
- Service restart

### Level 2: Senior Engineer
- Complex issues
- Performance problems
- Configuration changes

### Level 3: Architect/Lead
- Critical incidents
- Architecture decisions
- Major changes

## Contact Information

- **On-Call**: oncall@decube.io
- **Security**: security@decube.io
- **Support**: support@decube.io

## References

- [Troubleshooting Guide](troubleshooting.md)
- [Backup and Recovery](backup-recovery.md)
- [Monitoring Guide](monitoring.md)

---

*Last updated: January 2024*

