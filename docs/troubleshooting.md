# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with DeCube deployments.

## Quick Diagnosis

### Health Check Commands

```bash
# Check all services
decubectl status

# Check specific service
curl http://localhost:8080/api/v1/status

# Check cluster health
decubectl cluster health

# Run diagnostics
decubectl diagnose --all
```

### Log Collection

```bash
# Collect all logs
decubectl logs collect --output logs.tar.gz

# View service logs
docker-compose logs <service-name>

# Kubernetes logs
kubectl logs -l app=decube --tail=100 -n decube
```

## Service-Specific Issues

### Catalog Service Issues

#### Symptoms
- Snapshot queries return empty results
- CRDT operations fail
- High memory usage

#### Diagnosis
```bash
# Check catalog status
curl http://localhost:8080/api/v1/status

# Verify database connectivity
decubectl catalog db-check

# Check CRDT state
decubectl crdt status
```

#### Solutions

**Database Corruption**
```bash
# Stop service
docker-compose stop catalog

# Backup corrupted database
cp ./catalog.db ./catalog.db.backup

# Reset database
rm ./catalog.db
docker-compose up -d catalog

# Restore from backup if needed
# (Manual process required)
```

**Memory Issues**
```yaml
# Increase memory limits in docker-compose.yml
catalog:
  deploy:
    resources:
      limits:
        memory: 2G
      reservations:
        memory: 1G
```

**Network Partition**
```bash
# Check gossip connectivity
decubectl gossip peers

# Force resync
decubectl gossip sync --force
```

### Gossip Service Issues

#### Symptoms
- Peers not connecting
- Synchronization delays
- Merkle root mismatches

#### Diagnosis
```bash
# Check peer connections
decubectl gossip status

# Verify network connectivity
telnet <peer-ip> 8082

# Check firewall rules
sudo ufw status
```

#### Solutions

**Peer Discovery Issues**
```yaml
# Update initial peers in config
gossip:
  initial_peers:
    - "/ip4/192.168.1.100/tcp/8082/p2p/QmPeer1..."
    - "/ip4/192.168.1.101/tcp/8082/p2p/QmPeer2..."
```

**Port Binding Issues**
```bash
# Check port availability
netstat -tulpn | grep 8082

# Change port in configuration
gossip:
  listen_addr: "/ip4/0.0.0.0/tcp/8083"
```

**NAT Traversal**
```bash
# Enable UPnP (if available)
gossip:
  nat_traversal: true

# Configure port forwarding
# Router: Forward external port 8082 to internal 8082
```

### GCL (Global Consensus Layer) Issues

#### Symptoms
- Transactions not confirming
- High latency
- Consensus failures

#### Diagnosis
```bash
# Check consensus status
curl http://localhost:8081/api/v1/status

# View validator status
decubectl gcl validators

# Check transaction pool
decubectl gcl tx-pool
```

#### Solutions

**Validator Connectivity**
```yaml
# Update validator endpoints
gcl:
  validators:
    - "http://validator-1:8081"
    - "http://validator-2:8081"
    - "http://validator-3:8081"
```

**Consensus Timeout**
```yaml
# Increase timeouts for high-latency networks
gcl:
  consensus_timeout: "60s"
  block_time: "30s"
```

**Quorum Issues**
```bash
# Check validator count (need 2f+1 for f faulty nodes)
decubectl gcl validators | wc -l

# Add more validators if needed
decubectl gcl validator add <endpoint>
```

### Storage Service Issues

#### Symptoms
- Upload failures
- Chunk corruption
- Slow performance

#### Diagnosis
```bash
# Check storage status
curl http://localhost:9000/minio/health/ready

# Verify disk space
df -h /data

# Check chunk integrity
decubectl storage verify <snapshot-id>
```

#### Solutions

**Disk Space Issues**
```bash
# Add more storage
docker volume create --driver local \
  --opt type=tmpfs \
  --opt device=tmpfs \
  --opt o=size=10g \
  decube_storage

# Or expand existing volume
# (Depends on storage driver)
```

**Chunk Corruption**
```bash
# Re-upload corrupted chunks
decubectl storage repair <snapshot-id>

# Enable integrity checking
storage:
  integrity_check: true
  checksum_algorithm: "sha256"
```

**Performance Issues**
```yaml
# Optimize MinIO configuration
minio:
  environment:
    MINIO_REGION: "us-east-1"
    MINIO_BROWSER: "off"
  volumes:
    - ./data:/data
  command: server /data --console-address ":9001"
```

### Control Plane Issues

#### Symptoms
- API requests failing
- Orchestration not working
- Resource allocation issues

#### Diagnosis
```bash
# Check control plane status
curl http://localhost:8083/api/v1/status

# Verify etcd connectivity
decubectl etcd status

# Check resource usage
decubectl resources usage
```

#### Solutions

**etcd Issues**
```bash
# Check etcd cluster health
docker exec etcd etcdctl cluster-health

# Add etcd member
docker exec etcd etcdctl member add <name> <peer-url>

# Remove failed member
docker exec etcd etcdctl member remove <member-id>
```

**API Rate Limiting**
```yaml
# Adjust rate limits
control_plane:
  rate_limit:
    requests_per_second: 100
    burst: 200
```

**Resource Exhaustion**
```bash
# Check system resources
top
free -h
df -h

# Scale control plane
docker-compose up -d --scale control-plane=2
```

## Network Issues

### Connectivity Problems

#### DNS Resolution
```bash
# Check DNS
nslookup api.decube.local

# Update /etc/hosts if needed
echo "127.0.0.1 api.decube.local" >> /etc/hosts
```

#### Firewall Rules
```bash
# Check iptables
sudo iptables -L

# Allow DeCube ports
sudo ufw allow 8080:8085/tcp
sudo ufw allow 9000/tcp
sudo ufw allow 2379:2380/tcp
```

#### TLS Issues
```bash
# Check certificate validity
openssl x509 -in cert.pem -text -noout

# Test TLS connection
openssl s_client -connect localhost:8443 -servername api.decube.local
```

### Performance Issues

#### High Latency
```bash
# Measure network latency
ping <target-host>

# Check MTU
ip link show | grep mtu

# Adjust TCP settings
sysctl -w net.core.rmem_max=16777216
sysctl -w net.core.wmem_max=16777216
```

#### Packet Loss
```bash
# Check packet loss
ping -c 100 <target-host> | grep loss

# Diagnose network issues
mtr <target-host>

# Check interface errors
ip -s link show <interface>
```

## Database Issues

### etcd Problems

#### Cluster Unavailable
```bash
# Check cluster status
etcdctl cluster-health

# View member list
etcdctl member list

# Add learner (for recovery)
etcdctl member add <name> --peer-urls=<peer-url> --learner
```

#### Data Corruption
```bash
# Stop etcd
systemctl stop etcd

# Backup data directory
cp -r /var/lib/etcd /var/lib/etcd.backup

# Remove corrupted data
rm -rf /var/lib/etcd/member

# Restart etcd (will start fresh)
systemctl start etcd
```

#### Performance Issues
```bash
# Check etcd metrics
curl http://localhost:2379/metrics | grep etcd

# Optimize etcd configuration
etcd:
  quota-backend-bytes: 4294967296  # 4GB
  snapshot-count: 10000
  heartbeat-interval: 100
  election-timeout: 1000
```

### LevelDB Issues

#### Corruption
```bash
# Stop catalog service
docker-compose stop catalog

# Repair database
# (LevelDB has built-in repair)
# For custom databases, implement repair logic

# Restart service
docker-compose up -d catalog
```

#### Performance
```yaml
# Optimize LevelDB settings
catalog:
  environment:
    LEVELDB_CACHE_SIZE: "128MB"
    LEVELDB_WRITE_BUFFER_SIZE: "64MB"
  volumes:
    - ./catalog.db:/app/data
```

## Security Issues

### Authentication Failures

#### Token Issues
```bash
# Validate token
decubectl auth validate <token>

# Check token expiration
decubectl auth inspect <token>

# Renew token
decubectl auth renew
```

#### Certificate Issues
```bash
# Check certificate chain
openssl verify -CAfile ca.pem cert.pem

# Check certificate expiry
openssl x509 -in cert.pem -text | grep "Not After"

# Renew certificate
certbot renew --cert-name decube.local
```

### Authorization Issues

#### Permission Denied
```bash
# Check user roles
decubectl auth roles <user>

# Verify policy
decubectl auth policy check <resource> <action>

# Update permissions
decubectl auth grant <user> <role>
```

## Monitoring Issues

### Metrics Not Available

#### Prometheus Configuration
```yaml
# Check scrape configuration
curl http://localhost:9090/config

# Verify targets
curl http://localhost:9090/targets

# Check service annotations
kubectl describe service <service-name>
```

#### Grafana Dashboards
```bash
# Import dashboard
curl -X POST http://localhost:3000/api/dashboards/import \
  -H "Content-Type: application/json" \
  -d @decube-dashboard.json
```

### Alert Issues

#### False Positives
```yaml
# Adjust alert thresholds
alerts:
  high_cpu:
    threshold: 85
    duration: 5m
  low_memory:
    threshold: 10
    duration: 2m
```

#### Missing Alerts
```bash
# Check alert rules
curl http://localhost:9090/api/v1/rules

# Test alert
decubectl alert test <alert-name>
```

## Upgrade Issues

### Version Compatibility

#### Breaking Changes
```bash
# Check compatibility matrix
decubectl version compatibility <current> <target>

# Plan upgrade path
decubectl upgrade plan <target-version>

# Backup before upgrade
decubectl backup create pre-upgrade
```

#### Rollback Procedures
```bash
# Rollback deployment
kubectl rollout undo deployment/<deployment-name>

# Restore from backup
decubectl backup restore pre-upgrade

# Verify rollback
decubectl status
```

## Performance Tuning

### General Optimization

#### Resource Allocation
```yaml
# Adjust resource limits
services:
  catalog:
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 4G
        reservations:
          cpus: '1.0'
          memory: 2G
```

#### Database Tuning
```bash
# Optimize queries
decubectl db analyze

# Add indexes
decubectl db index create <table> <column>

# Vacuum database
decubectl db vacuum
```

### Advanced Troubleshooting

#### Debug Mode
```bash
# Enable debug logging
export DECUB_LOG_LEVEL=debug

# Start with debug flags
decubectl start --debug

# Collect debug information
decubectl debug collect --output debug.tar.gz
```

#### Profiling
```bash
# CPU profiling
go tool pprof http://localhost:8080/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:8080/debug/pprof/heap

# Goroutine profiling
curl http://localhost:8080/debug/pprof/goroutine > goroutines.txt
```

#### Core Dumps
```bash
# Enable core dumps
echo "core.%e.%p.%t" > /proc/sys/kernel/core_pattern
ulimit -c unlimited

# Analyze core dump
gdb <binary> <core-file>
```

## Getting Help

### Support Resources

- **Documentation**: https://docs.decube.io
- **Community Forum**: https://forum.decube.io
- **GitHub Issues**: https://github.com/REChain-Network-Solutions/DeCub/issues
- **Slack Channel**: #troubleshooting

### Escalation Process

1. **Self-Service**: Check documentation and known issues
2. **Community Support**: Post in forum or Slack
3. **Bug Report**: Create GitHub issue with full details
4. **Enterprise Support**: Contact support@decube.io for SLA-based assistance

### Information to Provide

When seeking help, include:

- DeCube version and commit hash
- Operating system and version
- Hardware specifications
- Configuration files (redacted)
- Log files from all affected services
- Steps to reproduce the issue
- Expected vs actual behavior
- Any recent changes or deployments

### Emergency Contacts

- **Security Issues**: security@decube.io (24/7)
- **Production Down**: emergency@decube.io (24/7)
- **General Support**: support@decube.io (business hours)

Remember to redact sensitive information like passwords, private keys, and personal data when sharing logs or configuration files.
