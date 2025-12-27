# Scaling Guide

This guide covers scaling DeCube deployments.

## Table of Contents

1. [Horizontal Scaling](#horizontal-scaling)
2. [Vertical Scaling](#vertical-scaling)
3. [Storage Scaling](#storage-scaling)
4. [Network Scaling](#network-scaling)
5. [Performance Optimization](#performance-optimization)

## Horizontal Scaling

### Adding Nodes

#### Add Node to Cluster

1. **Prepare Node**
   ```bash
   # Install DeCube
   ./scripts/setup-dev.sh
   ```

2. **Configure Node**
   ```yaml
   cluster:
     id: "cluster-001"
     nodes:
       - "node-1:8080"
       - "node-2:8080"
       - "node-3:8080"  # New node
   ```

3. **Join Cluster**
   ```bash
   # Node automatically joins via gossip protocol
   # Or manually:
   curl -X POST http://node-1:8080/cluster/join \
     -d '{"node_id": "node-3", "address": "node-3:8080"}'
   ```

4. **Verify**
   ```bash
   curl http://node-1:8080/cluster/nodes
   ```

### Scaling Services

#### Docker Compose

```bash
# Scale catalog service
docker-compose scale catalog=5

# Scale gossip service
docker-compose scale gossip=3
```

#### Kubernetes

```bash
# Scale deployment
kubectl scale deployment decube --replicas=10

# Auto-scaling
kubectl autoscale deployment decube --min=3 --max=10 --cpu-percent=70
```

## Vertical Scaling

### Resource Allocation

#### CPU

```yaml
resources:
  requests:
    cpu: "2"
  limits:
    cpu: "8"
```

#### Memory

```yaml
resources:
  requests:
    memory: "4Gi"
  limits:
    memory: "16Gi"
```

### Performance Tuning

#### Connection Pooling

```yaml
performance:
  connection_pool_size: 200
  max_idle_conns: 50
  max_idle_conns_per_host: 10
```

#### Worker Pools

```yaml
performance:
  worker_pool_size: 20
  max_concurrent_requests: 2000
```

## Storage Scaling

### Distributed Storage

#### CAS Scaling

```yaml
storage:
  cas:
    sharding:
      enabled: true
      shards: 10
    replication:
      factor: 3
```

#### Object Storage Scaling

- Use S3-compatible storage
- Configure multiple buckets
- Enable versioning
- Use lifecycle policies

### Storage Optimization

#### Caching

```yaml
storage:
  cache:
    enabled: true
    size: "10GB"
    ttl: "1h"
    strategy: "lru"
```

#### Compression

```yaml
storage:
  compression:
    algorithm: "lz4"
    level: 4
```

## Network Scaling

### Load Balancing

#### Nginx Configuration

```nginx
upstream decube {
    least_conn;
    server node-1:8080;
    server node-2:8080;
    server node-3:8080;
}

server {
    listen 80;
    location / {
        proxy_pass http://decube;
    }
}
```

#### HAProxy Configuration

```haproxy
backend decube
    balance roundrobin
    server node1 node-1:8080 check
    server node2 node-2:8080 check
    server node3 node-3:8080 check
```

### Network Optimization

#### Connection Multiplexing

```yaml
network:
  http2:
    enabled: true
  keep_alive:
    enabled: true
    timeout: "30s"
```

## Performance Optimization

### Database Optimization

#### Indexing

```sql
-- Create indexes for common queries
CREATE INDEX idx_snapshots_cluster ON snapshots(cluster);
CREATE INDEX idx_snapshots_created ON snapshots(created_at);
```

#### Query Optimization

- Use appropriate indexes
- Limit result sets
- Use pagination
- Cache frequently accessed data

### Caching Strategy

#### Multi-Level Caching

1. **L1 Cache**: In-memory (local)
2. **L2 Cache**: Distributed (Redis)
3. **L3 Cache**: CDN (for static content)

#### Cache Configuration

```yaml
cache:
  levels:
    - type: "memory"
      size: "1GB"
      ttl: "5m"
    - type: "redis"
      endpoint: "redis:6379"
      ttl: "1h"
```

## Scaling Patterns

### Sharding

#### Data Sharding

```yaml
sharding:
  strategy: "consistent-hashing"
  shards: 10
  replication_factor: 3
```

### Replication

#### Read Replicas

```yaml
replication:
  read_replicas: 5
  write_primary: true
```

### Partitioning

#### Partition by Cluster

```yaml
partitioning:
  strategy: "cluster-based"
  partitions:
    - cluster: "cluster-a"
      nodes: ["node-1", "node-2"]
    - cluster: "cluster-b"
      nodes: ["node-3", "node-4"]
```

## Monitoring Scaling

### Metrics to Monitor

- Request rate per node
- Response time distribution
- Resource utilization
- Queue depths
- Error rates

### Scaling Triggers

```yaml
autoscaling:
  triggers:
    - metric: "cpu_usage"
      threshold: 70
      action: "scale_up"
    - metric: "request_rate"
      threshold: 1000
      action: "scale_up"
    - metric: "cpu_usage"
      threshold: 30
      action: "scale_down"
```

## Best Practices

### Planning

1. **Capacity Planning**
   - Estimate load
   - Plan for growth
   - Reserve capacity

2. **Gradual Scaling**
   - Scale incrementally
   - Monitor impact
   - Adjust as needed

3. **Load Testing**
   - Test scaling scenarios
   - Identify bottlenecks
   - Optimize before scaling

### Implementation

1. **Start Small**
   - Begin with minimal resources
   - Scale based on actual needs
   - Monitor and adjust

2. **Use Auto-Scaling**
   - Configure auto-scaling rules
   - Set appropriate thresholds
   - Monitor auto-scaling behavior

3. **Optimize First**
   - Optimize before scaling
   - Fix bottlenecks
   - Improve efficiency

## Troubleshooting Scaling Issues

### Common Issues

#### Performance Degradation

- Check resource limits
- Review connection pooling
- Optimize queries
- Check for bottlenecks

#### Uneven Load Distribution

- Review load balancing
- Check health checks
- Verify node health
- Adjust load balancing algorithm

#### Storage Issues

- Check disk space
- Review I/O performance
- Optimize storage configuration
- Consider storage scaling

## References

- [Performance Tuning](performance-tuning.md)
- [Deployment Guide](deployment.md)
- [Monitoring Guide](monitoring.md)

---

*Last updated: January 2024*

