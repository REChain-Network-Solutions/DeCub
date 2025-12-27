# Observability Guide

This guide covers setting up comprehensive observability for DeCube.

## Table of Contents

1. [Metrics](#metrics)
2. [Logging](#logging)
3. [Tracing](#tracing)
4. [Dashboards](#dashboards)
5. [Alerts](#alerts)
6. [Best Practices](#best-practices)

## Metrics

### Prometheus Integration

#### Configuration

```yaml
# config.yaml
metrics:
  enabled: true
  bind_addr: "0.0.0.0:9090"
  path: "/metrics"
  
  prometheus:
    enabled: true
    endpoint: "http://prometheus:9090"
```

#### Prometheus Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'decube'
    static_configs:
      - targets: ['decube:9090']
    metrics_path: '/metrics'
    scrape_interval: 30s
```

### Key Metrics

#### System Metrics
- `decube_cpu_usage` - CPU utilization
- `decube_memory_usage` - Memory usage
- `decube_disk_usage` - Disk usage
- `decube_network_bytes` - Network I/O

#### Application Metrics
- `decube_http_requests_total` - Total HTTP requests
- `decube_http_request_duration_seconds` - Request duration
- `decube_consensus_operations_total` - Consensus operations
- `decube_storage_operations_total` - Storage operations

#### Business Metrics
- `decube_snapshots_total` - Total snapshots
- `decube_snapshots_size_bytes` - Total snapshot size
- `decube_catalog_entries_total` - Catalog entries

### Querying Metrics

```promql
# Request rate
rate(decube_http_requests_total[5m])

# Error rate
rate(decube_http_requests_total{status=~"5.."}[5m])

# P95 latency
histogram_quantile(0.95, decube_http_request_duration_seconds_bucket)

# CPU usage
decube_cpu_usage

# Memory usage
decube_memory_usage
```

## Logging

### Structured Logging

#### Configuration

```yaml
logging:
  level: "info"  # debug, info, warn, error
  format: "json"  # json, text
  output: "stdout"  # stdout, file
  file: "/var/log/decube/decube.log"
  max_size: 100  # MB
  max_backups: 10
  max_age: 30  # days
```

#### Log Format

```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "info",
  "component": "catalog",
  "message": "Snapshot created",
  "snapshot_id": "snapshot-001",
  "cluster": "cluster-a"
}
```

### Log Aggregation

#### ELK Stack

```yaml
# Filebeat configuration
filebeat.inputs:
- type: container
  paths:
    - '/var/lib/docker/containers/*/*.log'
  processors:
    - add_kubernetes_metadata:
        host: ${NODE_NAME}
        matchers:
        - logs_path:
            logs_path: "/var/log/containers/"
```

#### Loki

```yaml
# Promtail configuration
scrape_configs:
- job_name: decube
  static_configs:
  - targets:
    - localhost
    labels:
      job: decube
      __path__: /var/log/decube/*.log
```

## Tracing

### OpenTelemetry

#### Configuration

```yaml
tracing:
  enabled: true
  exporter: "jaeger"  # jaeger, zipkin, otlp
  endpoint: "http://jaeger:14268/api/traces"
  sample_rate: 0.1  # 10% sampling
```

#### Instrumentation

```go
import (
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

func createSnapshot(ctx context.Context) {
    ctx, span := otel.Tracer("decube").Start(ctx, "createSnapshot")
    defer span.End()
    
    // Your code here
}
```

### Jaeger Integration

#### Configuration

```yaml
# docker-compose.yml
services:
  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686"  # UI
      - "14268:14268"  # HTTP collector
```

## Dashboards

### Grafana Dashboards

#### System Dashboard

```json
{
  "dashboard": {
    "title": "DeCube System Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(decube_http_requests_total[5m])"
          }
        ]
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(decube_http_requests_total{status=~\"5..\"}[5m])"
          }
        ]
      },
      {
        "title": "P95 Latency",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, decube_http_request_duration_seconds_bucket)"
          }
        ]
      }
    ]
  }
}
```

#### Business Dashboard

```json
{
  "dashboard": {
    "title": "DeCube Business Metrics",
    "panels": [
      {
        "title": "Total Snapshots",
        "targets": [
          {
            "expr": "decube_snapshots_total"
          }
        ]
      },
      {
        "title": "Snapshot Size",
        "targets": [
          {
            "expr": "decube_snapshots_size_bytes"
          }
        ]
      }
    ]
  }
}
```

## Alerts

### Alertmanager Configuration

```yaml
# alertmanager.yml
route:
  group_by: ['alertname', 'cluster']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 12h
  receiver: 'default'
  routes:
  - match:
      severity: critical
    receiver: 'critical'
receivers:
- name: 'default'
  webhook_configs:
  - url: 'http://alertmanager:9093/api/v1/alerts'
- name: 'critical'
  email_configs:
  - to: 'oncall@example.com'
```

### Alert Rules

```yaml
# alerts.yml
groups:
- name: decube
  rules:
  - alert: HighErrorRate
    expr: rate(decube_http_requests_total{status=~"5.."}[5m]) > 0.1
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "High error rate detected"
      
  - alert: HighLatency
    expr: histogram_quantile(0.95, decube_http_request_duration_seconds_bucket) > 1
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "High latency detected"
      
  - alert: DiskSpaceLow
    expr: decube_disk_usage > 0.9
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Disk space running low"
```

## Best Practices

### Metrics

1. **Use Histograms for Latency**: Capture distribution
2. **Label Appropriately**: Don't over-label
3. **Monitor Cardinality**: Keep metric cardinality reasonable
4. **Use Rate Functions**: For counters, use rate()
5. **Set Appropriate Intervals**: Balance detail vs. cost

### Logging

1. **Structured Logging**: Use JSON format
2. **Appropriate Levels**: Use correct log levels
3. **Include Context**: Add relevant fields
4. **Avoid Sensitive Data**: Don't log secrets
5. **Rotate Logs**: Implement log rotation

### Tracing

1. **Sample Appropriately**: Use sampling for high-volume
2. **Trace Key Operations**: Focus on important paths
3. **Keep Traces Short**: Limit trace duration
4. **Use Context**: Propagate trace context

### Dashboards

1. **Keep It Simple**: Focus on key metrics
2. **Use Appropriate Time Ranges**: Match use case
3. **Group Related Metrics**: Logical organization
4. **Set Thresholds**: Visual indicators for alerts

## Tools and Integrations

### Prometheus
- Metrics collection and storage
- Query language (PromQL)
- Alerting rules

### Grafana
- Visualization and dashboards
- Alerting
- Data source integration

### ELK Stack
- Elasticsearch: Log storage
- Logstash: Log processing
- Kibana: Log visualization

### Jaeger
- Distributed tracing
- Service dependency mapping
- Performance analysis

## References

- [Monitoring Guide](monitoring.md)
- [Performance Tuning](performance-tuning.md)
- [Troubleshooting Guide](troubleshooting.md)

---

*Last updated: January 2024*

