# Performance Guide

This guide covers performance optimization, benchmarking, and monitoring for DeCube deployments.

## Performance Characteristics

### Baseline Performance Metrics

| Operation | Latency (P50) | Latency (P95) | Throughput |
|-----------|---------------|---------------|------------|
| Snapshot Create (1GB) | 30s | 60s | 50 MB/s |
| Snapshot Restore (1GB) | 45s | 90s | 35 MB/s |
| CRDT Query | 5ms | 15ms | 10,000 ops/s |
| Gossip Sync | 100ms | 500ms | 1,000 ops/s |
| GCL Transaction | 2s | 5s | 100 tx/s |
| CAS Store (64MB) | 200ms | 500ms | 320 MB/s |
| CAS Retrieve (64MB) | 150ms | 300ms | 426 MB/s |

*Metrics based on 3-node cluster with 10Gbps network and SSD storage*

## System Requirements

### Minimum Requirements
- **CPU**: 4 cores per node
- **RAM**: 8GB per node
- **Storage**: 500GB SSD per node
- **Network**: 1Gbps

### Recommended Requirements
- **CPU**: 8+ cores per node
- **RAM**: 16GB+ per node
- **Storage**: 1TB+ NVMe per node
- **Network**: 10Gbps

### Scaling Guidelines

#### Vertical Scaling
- **CPU**: Linear scaling up to 16 cores
- **RAM**: 2GB per 1GB dataset
- **Storage**: 3x dataset size for snapshots

#### Horizontal Scaling
- **Catalog**: 3-5 nodes for high availability
- **Gossip**: 5+ nodes for global coverage
- **GCL**: 4+ nodes (BFT requirement: 3f+1)
- **Storage**: 3+ nodes for redundancy

## Performance Optimization

### Application Layer Optimization

#### API Gateway Tuning
```yaml
api_gateway:
  worker_processes: 4
  worker_connections: 1024
  keepalive_timeout: 65
  client_max_body_size: 100M

  # Rate limiting
  rate_limit:
    requests_per_second: 1000
    burst: 2000

  # Caching
  cache:
    size: 1G
    ttl: 300
```

#### Service Mesh Configuration
```yaml
istio:
  pilot:
    resources:
      requests:
        cpu: 500m
        memory: 2Gi
      limits:
        cpu: 1000m
        memory: 4Gi

  # Circuit breakers
  circuit_breaker:
    maxConnections: 100
    maxPendingRequests: 50
    maxRequests: 1000

  # Load balancing
  load_balancer:
    policy: ROUND_ROBIN
    healthy_panic_threshold: 30
```

### Consensus Layer Optimization

#### BFT Parameter Tuning
```go
// Consensus configuration
type ConsensusConfig struct {
    // Timeout settings
    PreprepareTimeout  time.Duration `default:"30s"`
    PrepareTimeout     time.Duration `default:"60s"`
    CommitTimeout      time.Duration `default:"90s"`

    // Batch settings
    MaxBatchSize       int           `default:"100"`
    BatchTimeout       time.Duration `default:"1s"`

    // Network settings
    MaxMessageSize     int           `default:"10MB"`
    GossipFanout       int           `default:"3"`

    // Performance settings
    ParallelWorkers    int           `default:"4"`
    QueueSize          int           `default:"1000"`
}
```

#### RAFT Optimization
```yaml
raft:
  # Timing
  heartbeat_timeout: 1000ms
  election_timeout: 5000ms

  # Performance
  max_append_entries: 64
  batch_apply: true

  # Storage
  wal_sync: true
  snapshot_threshold: 8192
  compaction_overhead: 5
```

### Storage Layer Optimization

#### Object Storage Tuning
```yaml
minio:
  # Performance settings
  drives_per_node: 4
  erasure_sets: 4
  parity: 2

  # Network
  max_connections: 1000
  buffer_size: 16MB

  # Cache
  cache:
    enabled: true
    size: 1G
    expiry: 72h
```

#### Database Optimization

**etcd Tuning**
```bash
# etcd configuration
ETCD_QUOTA_BACKEND_BYTES=8589934592  # 8GB
ETCD_SNAPSHOT_COUNT=100000
ETCD_HEARTBEAT_INTERVAL=100
ETCD_ELECTION_TIMEOUT=1000
ETCD_MAX_REQUEST_BYTES=1048576  # 1MB
ETCD_MAX_CONCURRENT_STREAMS=1024
```

**LevelDB Tuning**
```go
// LevelDB options
options := &opt.Options{
    BlockCacheCapacity: 64 * opt.MiB,
    WriteBuffer:        32 * opt.MiB,
    MaxOpenFiles:       1000,
    Compression:        opt.SnappyCompression,
    Filter:             filter.NewBloomFilter(10),
}
```

### Network Layer Optimization

#### Gossip Protocol Tuning
```yaml
gossip:
  # Timing
  gossip_interval: 200ms
  push_pull_interval: 5000ms
  probe_interval: 1000ms

  # Performance
  indirect_checks: 3
  retransmit_mult: 4
  suspicion_mult: 4

  # Network
  tcp_timeout: 10000ms
  probe_timeout: 500ms
  gossip_nodes: 3
```

#### P2P Network Optimization
```go
// libp2p configuration
host, err := libp2p.New(
    libp2p.ListenAddrStrings(listenAddr),
    libp2p.NATPortMap(),
    libp2p.ConnectionManager(connmgr.NewConnManager(100, 400, time.Minute)),
    libp2p.EnableNATService(),
    libp2p.AutoNATService(),
    libp2p.Routing(routing),
    libp2p.EnableRelay(),
)
```

### Security Layer Optimization

#### Cryptographic Performance
```go
// AES-GCM optimization
cipher, err := aes.NewCipher(key)
if err != nil {
    return err
}

gcm, err := cipher.NewGCM(cipher.BlockSize() * 8)
if err != nil {
    return err
}

// Use hardware acceleration if available
if hasAESNI() {
    // Hardware-accelerated AES
    return gcm.Seal(nonce, nonce, plaintext, additionalData)
}
```

#### TLS Optimization
```yaml
tls:
  # Protocol
  min_version: TLS_1_3
  max_version: TLS_1_3

  # Cipher suites (TLS 1.3 optimized)
  cipher_suites:
    - TLS_AES_256_GCM_SHA384
    - TLS_AES_128_GCM_SHA256

  # Session resumption
  session_tickets: true
  session_cache_size: 1024

  # OCSP stapling
  ocsp_stapling: true
```

## Benchmarking

### Benchmark Suite

#### Snapshot Operations
```bash
# Create benchmark
decubectl benchmark snapshot create \
  --size 1GB \
  --iterations 10 \
  --output results.json

# Restore benchmark
decubectl benchmark snapshot restore \
  --snapshot-id snap-001 \
  --iterations 10 \
  --output results.json
```

#### CRDT Operations
```bash
# CRDT benchmark
decubectl benchmark crdt \
  --operations 10000 \
  --concurrency 10 \
  --type orset \
  --output results.json
```

#### Consensus Operations
```bash
# Consensus benchmark
decubectl benchmark consensus \
  --transactions 1000 \
  --payload-size 1KB \
  --validators 4 \
  --output results.json
```

### Custom Benchmarks

#### Go Benchmarking
```go
func BenchmarkSnapshotCreation(b *testing.B) {
    for i := 0; i < b.N; i++ {
        snapshot := createTestSnapshot(1 * opt.GiB)
        b.SetBytes(int64(snapshot.Size))
    }
}

func BenchmarkCRDTMerge(b *testing.B) {
    crdt := NewORSet()
    b.ResetTimer()

    for i := 0; i < b.N; i++ {
        crdt.Add(fmt.Sprintf("item-%d", i), fmt.Sprintf("value-%d", i))
    }
}
```

#### Load Testing
```bash
# Apache Bench
ab -n 10000 -c 100 http://localhost:8080/api/v1/snapshots

# hey (HTTP load testing)
hey -n 10000 -c 100 http://localhost:8080/api/v1/catalog/query

# Vegeta (advanced load testing)
echo "GET http://localhost:8080/api/v1/status" | \
  vegeta attack -duration=30s -rate=100 | \
  vegeta report
```

## Monitoring and Metrics

### Key Metrics to Monitor

#### System Metrics
- **CPU Usage**: < 80% sustained
- **Memory Usage**: < 90% of allocated
- **Disk I/O**: < 80% utilization
- **Network I/O**: < 70% capacity

#### Application Metrics
- **Request Latency**: P95 < 500ms for APIs
- **Error Rate**: < 1% for critical operations
- **Throughput**: Meet SLA requirements
- **Queue Depth**: < 1000 pending operations

#### Business Metrics
- **Snapshot Success Rate**: > 99.9%
- **Data Durability**: 11 9's (99.999999999%)
- **RTO (Recovery Time Objective)**: < 1 hour
- **RPO (Recovery Point Objective)**: < 5 minutes

### Monitoring Setup

#### Prometheus Configuration
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "alert_rules.yml"

scrape_configs:
  - job_name: 'decube'
    static_configs:
      - targets: ['localhost:8080', 'localhost:8081', 'localhost:8082']
    metrics_path: '/metrics'
    scrape_interval: 5s

  - job_name: 'etcd'
    static_configs:
      - targets: ['localhost:2379']
    metrics_path: '/metrics'
```

#### Grafana Dashboards

**System Overview Dashboard**
- CPU, Memory, Disk, Network usage
- Service health status
- Error rates and latency

**Application Performance Dashboard**
- Request throughput and latency
- Database performance metrics
- Cache hit rates

**Consensus Dashboard**
- Block production rate
- Transaction finality time
- Validator participation

### Alerting Rules

```yaml
groups:
  - name: decube
    rules:
      - alert: HighCPUUsage
        expr: cpu_usage_percent > 85
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage detected"
          description: "CPU usage is {{ $value }}%"

      - alert: ServiceDown
        expr: up == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "Service {{ $labels.job }} is down"
          description: "Service {{ $labels.job }} has been down for more than 1 minute"

      - alert: HighLatency
        expr: http_request_duration_seconds{quantile="0.95"} > 1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High request latency"
          description: "95th percentile latency is {{ $value }}s"
```

## Profiling and Debugging

### CPU Profiling
```bash
# Go pprof
go tool pprof http://localhost:8080/debug/pprof/profile

# Flame graph
go tool pprof -http=:8081 http://localhost:8080/debug/pprof/profile
```

### Memory Profiling
```bash
# Heap profile
go tool pprof http://localhost:8080/debug/pprof/heap

# Memory leak detection
go tool pprof -alloc_space http://localhost:8080/debug/pprof/heap
```

### Goroutine Profiling
```bash
# Goroutine dump
curl http://localhost:8080/debug/pprof/goroutine > goroutines.txt

# Analyze blocking
go tool pprof http://localhost:8080/debug/pprof/block
```

### Tracing
```go
// Distributed tracing setup
tracer, closer := jaeger.NewTracer(
    "decube",
    jaeger.NewConstSampler(true),
    jaeger.NewInMemoryReporter(),
)
opentracing.SetGlobalTracer(tracer)

// Span creation
span := opentracing.StartSpan("snapshot.create")
defer span.Finish()

span.SetTag("snapshot.id", snapshotID)
span.SetTag("snapshot.size", size)
```

## Capacity Planning

### Sizing Calculator

#### Storage Sizing
```
Total Storage = (Dataset Size × 3) + (Snapshot Frequency × Retention Period × Dataset Size × 0.1)
```

#### Network Sizing
```
Network Bandwidth = (Concurrent Users × Average Request Size × 2) / Time Window
```

#### Compute Sizing
```
Required CPU = (Operations per Second × CPU per Operation) / CPU Efficiency
Required RAM = (Active Dataset Size × 2) + (Concurrent Connections × 8MB)
```

### Scaling Strategies

#### Horizontal Scaling Triggers
- CPU usage > 70%
- Memory usage > 80%
- Request latency > 200ms (P95)
- Queue depth > 100

#### Vertical Scaling Limits
- CPU: 32 cores maximum
- RAM: 512GB maximum
- Storage: Limited by hardware

#### Auto-scaling Configuration
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: decube-autoscaler
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: decube-app
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  behavior:
    scaleDown:
      stabilizationWindowSeconds: 300
      policies:
      - type: Percent
        value: 10
        periodSeconds: 60
```

## Performance Testing

### Load Testing Scenarios

#### Normal Load
- 100 concurrent users
- 1000 requests per minute
- 95% of requests < 500ms

#### Peak Load
- 500 concurrent users
- 5000 requests per minute
- 95% of requests < 2s

#### Stress Testing
- 1000+ concurrent users
- Sustained load for 1+ hours
- Monitor system stability

### Chaos Engineering

#### Network Partition Testing
```bash
# Simulate network partition
kubectl run chaos-mesh --image=pingcap/chaos-mesh \
  -- chaos apply network-partition \
  --selector app=decube \
  --direction to \
  --target-selector app=decube \
  --loss 100
```

#### Node Failure Testing
```bash
# Kill random pods
kubectl delete pod $(kubectl get pods -l app=decube -o jsonpath='{.items[0].metadata.name}')
```

#### Resource Exhaustion Testing
```bash
# CPU stress
stress-ng --cpu 4 --timeout 60s

# Memory stress
stress-ng --vm 2 --vm-bytes 4G --timeout 60s
```

## Optimization Checklist

### Pre-Deployment
- [ ] Hardware meets minimum requirements
- [ ] Network bandwidth sufficient
- [ ] Storage IOPS adequate
- [ ] OS tuning applied
- [ ] Security hardening complete

### Post-Deployment
- [ ] Monitoring configured
- [ ] Alerting rules active
- [ ] Backup strategy implemented
- [ ] Performance baselines established
- [ ] Load testing completed

### Ongoing Maintenance
- [ ] Regular performance reviews
- [ ] Capacity planning updates
- [ ] Software updates tested
- [ ] Security patches applied
- [ ] Documentation updated

## Troubleshooting Performance Issues

### High Latency
1. Check network connectivity
2. Profile application code
3. Optimize database queries
4. Scale resources horizontally
5. Implement caching layers

### High CPU Usage
1. Profile CPU usage
2. Optimize algorithms
3. Reduce logging verbosity
4. Scale vertically or horizontally
5. Offload compute-intensive tasks

### High Memory Usage
1. Check for memory leaks
2. Optimize data structures
3. Implement memory pooling
4. Scale memory resources
5. Tune garbage collection

### Low Throughput
1. Identify bottlenecks
2. Optimize concurrent processing
3. Implement request batching
4. Scale horizontally
5. Tune system parameters

This performance guide provides comprehensive information for optimizing and monitoring DeCube deployments. Regular performance testing and monitoring are essential for maintaining optimal system performance.
