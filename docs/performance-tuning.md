# Performance Tuning Guide

This guide covers performance optimization strategies for DeCube deployments.

## Table of Contents

1. [Performance Metrics](#performance-metrics)
2. [System Tuning](#system-tuning)
3. [Component Optimization](#component-optimization)
4. [Network Optimization](#network-optimization)
5. [Storage Optimization](#storage-optimization)
6. [Benchmarking](#benchmarking)

## Performance Metrics

### Key Metrics to Monitor

- **Throughput**: Transactions per second (TPS)
- **Latency**: P50, P95, P99 response times
- **Resource Usage**: CPU, memory, disk I/O, network
- **Consensus Time**: Time to reach consensus
- **Storage Performance**: Read/write speeds

### Target Performance

- **Local Operations**: <100ms latency, 10k+ TPS
- **Global Consensus**: <2s latency, 1k+ TPS
- **Storage**: 500MB/s write, 1GB/s read
- **Snapshot Creation**: 100GB in <5 minutes

## System Tuning

### Operating System

#### Linux Tuning

```bash
# Increase file descriptor limits
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf

# Optimize network settings
echo "net.core.somaxconn = 4096" >> /etc/sysctl.conf
echo "net.ipv4.tcp_max_syn_backlog = 4096" >> /etc/sysctl.conf
echo "net.ipv4.ip_local_port_range = 1024 65535" >> /etc/sysctl.conf

# Apply changes
sysctl -p
```

#### Memory Settings

```bash
# Disable swap for better performance (if sufficient RAM)
swapoff -a

# Adjust vm.swappiness
echo "vm.swappiness = 1" >> /etc/sysctl.conf
```

### Docker Tuning

```yaml
# docker-compose.yml
services:
  decube:
    deploy:
      resources:
        limits:
          cpus: '4'
          memory: 8G
        reservations:
          cpus: '2'
          memory: 4G
    ulimits:
      nofile:
        soft: 65536
        hard: 65536
```

### Kubernetes Tuning

```yaml
# k8s-manifest.yaml
resources:
  requests:
    cpu: "2"
    memory: "4Gi"
  limits:
    cpu: "4"
    memory: "8Gi"
```

## Component Optimization

### Consensus Layer

#### RAFT Tuning

```yaml
raft:
  heartbeat_timeout: "500ms"  # Reduce for faster elections
  election_timeout: "1s"      # Balance between speed and stability
  snapshot_interval: "30s"    # Adjust based on log size
  snapshot_threshold: 1000   # Number of entries before snapshot
  max_append_entries: 100     # Batch size for log replication
```

#### BFT Tuning

```yaml
gcl:
  batch_size: 100            # Transactions per block
  batch_timeout: "100ms"      # Max wait time for batch
  max_concurrent_requests: 1000
  request_timeout: "5s"
```

### Storage Layer

#### CAS Optimization

```yaml
storage:
  cas:
    chunk_size: 67108864     # 64MB chunks (adjust based on network)
    compression: "lz4"       # lz4 is faster than gzip
    cache_size: "1GB"        # In-memory cache size
    cache_ttl: "5m"           # Cache time-to-live
```

#### Object Storage

```yaml
storage:
  object_storage:
    multipart_threshold: 52428800  # 50MB
    multipart_chunk_size: 10485760 # 10MB
    max_concurrent_uploads: 10
    connection_pool_size: 100
```

### Gossip Protocol

```yaml
gossip:
  sync_interval: "5s"        # Reduce for faster sync
  anti_entropy_interval: "30s" # Balance between consistency and performance
  fanout: 3                  # Number of peers to gossip with
  max_message_size: 1048576  # 1MB max message size
```

### Catalog Service

```yaml
catalog:
  crdt_type: "orset"         # Choose based on use case
  batch_size: 100            # Batch CRDT operations
  compaction_interval: "1h"   # Garbage collection interval
```

## Network Optimization

### Connection Pooling

```yaml
api:
  rest:
    max_idle_conns: 100
    max_idle_conns_per_host: 10
    idle_conn_timeout: "90s"
    tls_handshake_timeout: "10s"
```

### Load Balancing

```yaml
# Use connection pooling
performance:
  connection_pool_size: 100
  keep_alive: true
  keep_alive_timeout: "30s"
```

### Network Tuning

```bash
# Increase TCP buffer sizes
echo "net.core.rmem_max = 16777216" >> /etc/sysctl.conf
echo "net.core.wmem_max = 16777216" >> /etc/sysctl.conf
echo "net.ipv4.tcp_rmem = 4096 87380 16777216" >> /etc/sysctl.conf
echo "net.ipv4.tcp_wmem = 4096 65536 16777216" >> /etc/sysctl.conf
```

## Storage Optimization

### Disk I/O

```bash
# Use fast storage (SSD/NVMe)
# Configure I/O scheduler
echo deadline > /sys/block/sda/queue/scheduler

# Increase read-ahead
blockdev --setra 8192 /dev/sda
```

### File System

```bash
# Use XFS or ext4 with appropriate options
# Mount with noatime for better performance
mount -o noatime /dev/sda1 /var/lib/decube
```

### Database Tuning

```yaml
storage:
  badger:
    value_log_file_size: 1073741824  # 1GB
    value_log_max_entries: 1000000
    num_compactors: 4
    num_goroutines: 16
```

## Benchmarking

### Running Benchmarks

```bash
# Run CRDT benchmarks
cd rechain/pkg/crdt
go test -bench=. -benchmem

# Run storage benchmarks
cd rechain/internal/storage
go test -bench=. -benchmem

# Run consensus benchmarks
cd decub-gcl/go
go test -bench=. -benchmem
```

### Benchmark Configuration

```go
// Example benchmark
func BenchmarkORSetAdd(b *testing.B) {
    orset := crdt.NewORSet()
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        orset.Add(fmt.Sprintf("item-%d", i))
    }
}
```

### Performance Testing

```bash
# Load testing with Apache Bench
ab -n 10000 -c 100 http://localhost:8080/catalog/snapshots

# Load testing with wrk
wrk -t4 -c100 -d30s http://localhost:8080/catalog/snapshots
```

## Monitoring Performance

### Metrics to Track

- Request rate (requests/second)
- Response time percentiles (P50, P95, P99)
- Error rate
- Resource utilization (CPU, memory, disk, network)
- Queue depths
- Cache hit rates

### Profiling

```bash
# CPU profiling
go tool pprof http://localhost:6060/debug/pprof/profile

# Memory profiling
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutine profiling
go tool pprof http://localhost:6060/debug/pprof/goroutine
```

## Optimization Checklist

### System Level
- [ ] OS tuned (file descriptors, network settings)
- [ ] Sufficient resources allocated
- [ ] Fast storage (SSD/NVMe)
- [ ] Network optimized

### Application Level
- [ ] Connection pooling configured
- [ ] Batch sizes optimized
- [ ] Timeouts appropriate
- [ ] Caching enabled where beneficial

### Component Level
- [ ] Consensus timeouts tuned
- [ ] Storage chunk sizes optimized
- [ ] Gossip intervals adjusted
- [ ] CRDT types chosen appropriately

### Monitoring
- [ ] Metrics collection enabled
- [ ] Alerts configured
- [ ] Dashboards set up
- [ ] Profiling available

## Performance Troubleshooting

### High Latency

1. Check network connectivity
2. Review consensus timeouts
3. Check storage I/O performance
4. Review resource utilization
5. Check for bottlenecks

### Low Throughput

1. Increase batch sizes
2. Optimize connection pooling
3. Review resource limits
4. Check for contention
5. Profile application

### High Resource Usage

1. Review memory allocations
2. Check for memory leaks
3. Optimize data structures
4. Review cache sizes
5. Profile CPU usage

## Best Practices

1. **Start Conservative**: Begin with default settings
2. **Measure First**: Profile before optimizing
3. **One Change at a Time**: Make incremental changes
4. **Monitor Impact**: Measure before and after
5. **Document Changes**: Keep track of what works

## References

- [Architecture Guide](architecture.md)
- [Deployment Guide](deployment.md)
- [Monitoring Guide](monitoring.md)
- [Troubleshooting Guide](troubleshooting.md)

---

*Last updated: January 2024*

