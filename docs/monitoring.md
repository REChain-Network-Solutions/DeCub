# Monitoring and Observability Guide

This guide covers monitoring, logging, and observability setup for DeCube deployments.

## Overview

Effective monitoring is crucial for maintaining the health, performance, and reliability of DeCube clusters. This guide provides comprehensive monitoring strategies and configurations.

## Monitoring Architecture

### Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DeCube Nodes  │    │  Prometheus     │    │   Grafana       │
│                 │    │                 │    │                 │
│ • Service Metrics│───▶│ • Metrics       │───▶│ • Dashboards    │
│ • System Metrics │    │   Collection    │    │ • Alerts        │
│ • Custom Metrics │    │ • Alerting      │    │ • Analytics     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   AlertManager  │
                    │                 │
                    │ • Alert Routing │
                    │ • Notification  │
                    │ • Silencing     │
                    └─────────────────┘
```

### Metrics Collection

#### Service Metrics
DeCube exposes metrics via HTTP endpoints:

- **Catalog Service**: `http://localhost:8080/metrics`
- **Gossip Service**: `http://localhost:8082/metrics`
- **GCL Service**: `http://localhost:8081/metrics`
- **Control Plane**: `http://localhost:8083/metrics`

#### System Metrics
Node-level metrics collected via node exporters:

- CPU, memory, disk, and network utilization
- System load and process information
- Filesystem and mount point details

#### Custom Metrics
Application-specific metrics:

- Snapshot operation counters
- Consensus round duration
- CRDT merge conflicts
- Gossip message propagation

## Prometheus Configuration

### Basic Setup

```yaml
# prometheus.yml
global:
  scrape_interval: 15s
  evaluation_interval: 15s
  scrape_timeout: 10s

rule_files:
  - "alert_rules.yml"

scrape_configs:
  - job_name: 'decube-catalog'
    static_configs:
      - targets: ['catalog-1:8080', 'catalog-2:8080', 'catalog-3:8080']
    metrics_path: '/metrics'
    scrape_interval: 30s

  - job_name: 'decube-gossip'
    static_configs:
      - targets: ['gossip-1:8082', 'gossip-2:8082']
    metrics_path: '/metrics'
    scrape_interval: 30s

  - job_name: 'decube-gcl'
    static_configs:
      - targets: ['gcl-1:8081', 'gcl-2:8081', 'gcl-3:8081', 'gcl-4:8081']
    metrics_path: '/metrics'
    scrape_interval: 30s

  - job_name: 'decube-control-plane'
    static_configs:
      - targets: ['control-plane:8083']
    metrics_path: '/metrics'
    scrape_interval: 30s

  - job_name: 'node'
    static_configs:
      - targets: ['node1:9100', 'node2:9100', 'node3:9100']
    scrape_interval: 30s
```

### Service Discovery

#### Kubernetes Service Discovery
```yaml
scrape_configs:
  - job_name: 'decube-services'
    kubernetes_sd_configs:
      - role: pod
    relabel_configs:
      - source_labels: [__meta_kubernetes_pod_label_app]
        regex: decube-.*
        action: keep
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
        regex: "true"
        action: keep
      - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_path]
        target_label: __metrics_path__
        regex: (.+)
        replacement: $1
      - source_labels: [__address__, __meta_kubernetes_pod_annotation_prometheus_io_port]
        target_label: __address__
        regex: ([^:]+)(?::\d+)?;(\d+)
        replacement: $1:$2
```

#### Consul Service Discovery
```yaml
scrape_configs:
  - job_name: 'decube-consul'
    consul_sd_configs:
      - server: 'consul:8500'
        services: ['decube-catalog', 'decube-gossip', 'decube-gcl']
    relabel_configs:
      - source_labels: ['__meta_consul_service']
        regex: 'decube-(.*)'
        target_label: 'service'
        replacement: '$1'
```

## Grafana Dashboards

### System Overview Dashboard

```json
{
  "dashboard": {
    "title": "DeCube System Overview",
    "tags": ["decube", "system"],
    "timezone": "browser",
    "panels": [
      {
        "title": "Cluster Health",
        "type": "stat",
        "targets": [
          {
            "expr": "up{job=~\"decube-.*\"}",
            "legendFormat": "{{job}}"
          }
        ]
      },
      {
        "title": "CPU Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "100 - (avg by(instance) (irate(node_cpu_seconds_total{mode=\"idle\"}[5m])) * 100)",
            "legendFormat": "{{instance}}"
          }
        ]
      },
      {
        "title": "Memory Usage",
        "type": "graph",
        "targets": [
          {
            "expr": "(1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100",
            "legendFormat": "{{instance}}"
          }
        ]
      }
    ]
  }
}
```

### Application Performance Dashboard

```json
{
  "dashboard": {
    "title": "DeCube Application Performance",
    "tags": ["decube", "performance"],
    "panels": [
      {
        "title": "Request Latency",
        "type": "heatmap",
        "targets": [
          {
            "expr": "rate(http_request_duration_seconds_bucket[5m])",
            "legendFormat": "{{le}}"
          }
        ]
      },
      {
        "title": "Request Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total[5m])",
            "legendFormat": "{{method}} {{path}}"
          }
        ]
      },
      {
        "title": "Error Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(http_requests_total{status=~\"5..\"}[5m]) / rate(http_requests_total[5m]) * 100",
            "legendFormat": "{{service}}"
          }
        ]
      }
    ]
  }
}
```

### Consensus Dashboard

```json
{
  "dashboard": {
    "title": "DeCube Consensus Performance",
    "tags": ["decube", "consensus"],
    "panels": [
      {
        "title": "Block Production Rate",
        "type": "graph",
        "targets": [
          {
            "expr": "rate(gcl_blocks_total[5m])",
            "legendFormat": "Blocks/s"
          }
        ]
      },
      {
        "title": "Transaction Finality Time",
        "type": "heatmap",
        "targets": [
          {
            "expr": "gcl_transaction_finality_duration_seconds_bucket",
            "legendFormat": "{{le}}"
          }
        ]
      },
      {
        "title": "Validator Participation",
        "type": "bargauge",
        "targets": [
          {
            "expr": "gcl_validator_participation_ratio",
            "legendFormat": "{{validator}}"
          }
        ]
      }
    ]
  }
}
```

## Alerting

### Alert Rules

```yaml
groups:
  - name: decube
    rules:
      - alert: DeCubeServiceDown
        expr: up{job=~"^decube-.*"} == 0
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "DeCube service {{ $labels.job }} is down"
          description: "DeCube service {{ $labels.job }} on {{ $labels.instance }} has been down for more than 5 minutes."

      - alert: HighCPUUsage
        expr: 100 - (avg by(instance) (irate(node_cpu_seconds_total{mode="idle"}[5m])) * 100) > 85
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High CPU usage on {{ $labels.instance }}"
          description: "CPU usage is {{ $value }}% on {{ $labels.instance }}"

      - alert: HighMemoryUsage
        expr: (1 - node_memory_MemAvailable_bytes / node_memory_MemTotal_bytes) * 100 > 90
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High memory usage on {{ $labels.instance }}"
          description: "Memory usage is {{ $value }}% on {{ $labels.instance }}"

      - alert: DiskSpaceLow
        expr: (node_filesystem_avail_bytes / node_filesystem_size_bytes) * 100 < 10
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "Low disk space on {{ $labels.instance }}"
          description: "Disk space is {{ $value }}% available on {{ $labels.instance }}"

      - alert: HighRequestLatency
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 1
        for: 10m
        labels:
          severity: warning
        annotations:
          summary: "High request latency on {{ $labels.service }}"
          description: "95th percentile latency is {{ $value }}s on {{ $labels.service }}"

      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"^5.."}[5m]) / rate(http_requests_total[5m]) * 100 > 5
        for: 5m
        labels:
          severity: critical
        annotations:
          summary: "High error rate on {{ $labels.service }}"
          description: "Error rate is {{ $value }}% on {{ $labels.service }}"
```

### Alertmanager Configuration

```yaml
global:
  smtp_smarthost: 'smtp.gmail.com:587'
  smtp_from: 'alerts@decube.io'
  smtp_auth_username: 'alerts@decube.io'
  smtp_auth_password: 'password'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'team'
  routes:
    - match:
        severity: critical
      receiver: 'team-critical'

receivers:
  - name: 'team'
    email_configs:
      - to: 'team@decube.io'
        send_resolved: true
    slack_configs:
      - api_url: 'https://hooks.slack.com/services/.../.../...'
        channel: '#alerts'
        send_resolved: true

  - name: 'team-critical'
    email_configs:
      - to: 'team@decube.io'
        send_resolved: true
    pagerduty_configs:
      - service_key: 'pagerduty-service-key'
```

## Logging

### Centralized Logging Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DeCube Nodes  │    │   Fluent Bit    │    │   Elasticsearch │
│                 │    │                 │    │                 │
│ • Application   │───▶│ • Log Collection│───▶│ • Log Storage   │
│ • System Logs   │    │ • Filtering     │    │ • Search        │
│ • Audit Logs    │    │ • Enrichment    │    │ • Analytics     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                         │
                                            ┌─────────────────┐
                                            │     Kibana      │
                                            │                 │
                                            │ • Dashboards    │
                                            │ • Search UI     │
                                            │ • Analytics     │
                                            └─────────────────┘
```

### Log Collection Configuration

#### Fluent Bit Configuration
```ini
[INPUT]
    Name              tail
    Path              /var/log/decube/*.log
    Parser            json
    Tag               decube.*
    Refresh_Interval  5

[INPUT]
    Name              systemd
    Tag               systemd.*
    Systemd_Filter    _SYSTEMD_UNIT=decube.service

[FILTER]
    Name                record_modifier
    Match               decube.*
    Record              cluster ${CLUSTER_NAME}
    Record              node ${NODE_NAME}

[OUTPUT]
    Name  elasticsearch
    Match *
    Host  elasticsearch
    Port  9200
    Index decube-%Y.%m.%d
    Type  _doc
```

#### Logstash Configuration
```ruby
input {
  beats {
    port => 5044
  }
}

filter {
  if [kubernetes] {
    mutate {
      add_field => {
        "cluster" => "%{[kubernetes][namespace]}"
        "pod" => "%{[kubernetes][pod][name]}"
        "container" => "%{[kubernetes][container][name]}"
      }
    }
  }

  grok {
    match => { "message" => "%{TIMESTAMP_ISO8601:timestamp} %{LOGLEVEL:level} %{DATA:component} %{GREEDYDATA:message}" }
  }

  date {
    match => [ "timestamp", "ISO8601" ]
    target => "@timestamp"
  }
}

output {
  elasticsearch {
    hosts => ["elasticsearch:9200"]
    index => "decube-%{+YYYY.MM.dd}"
  }
}
```

### Structured Logging

#### Application Logging
```go
// Structured logging example
logger := logrus.New()
logger.SetFormatter(&logrus.JSONFormatter{
    TimestampFormat: time.RFC3339,
})

// Log with context
logger.WithFields(logrus.Fields{
    "service": "catalog",
    "operation": "snapshot_create",
    "snapshot_id": snapshotID,
    "user_id": userID,
    "cluster": clusterID,
}).Info("Snapshot creation started")

// Error logging
logger.WithFields(logrus.Fields{
    "service": "gossip",
    "error": err.Error(),
    "peer": peerID,
}).Error("Failed to connect to peer")
```

#### Audit Logging
```go
// Audit log entry
auditLogger.WithFields(logrus.Fields{
    "event": "snapshot_access",
    "user": userID,
    "resource": snapshotID,
    "action": "read",
    "ip": clientIP,
    "timestamp": time.Now().Format(time.RFC3339),
    "result": "success",
}).Info("Audit event")
```

## Tracing

### Distributed Tracing Setup

#### Jaeger Configuration
```yaml
apiVersion: jaegertracing.io/v1
kind: Jaeger
metadata:
  name: decube-tracing
spec:
  strategy: allInOne
  allInOne:
    image: jaegertracing/all-in-one:latest
    options:
      log-level: info
  storage:
    type: memory
    options:
      memory:
        max-traces: 100000
  ui:
    options:
      dependencies:
        menuEnabled: false
```

#### Application Instrumentation
```go
// Initialize tracer
tracer, closer := jaeger.NewTracer(
    "decube-catalog",
    jaeger.NewConstSampler(true),
    jaeger.NewInMemoryReporter(),
)
defer closer.Close()
opentracing.SetGlobalTracer(tracer)

// Create span for operation
span := opentracing.StartSpan("catalog.snapshot.create")
defer span.Finish()

span.SetTag("snapshot.id", snapshotID)
span.SetTag("user.id", userID)

// Child span for database operation
dbSpan := opentracing.StartSpan("catalog.db.insert", opentracing.ChildOf(span.Context()))
defer dbSpan.Finish()

// Log events
span.LogKV("event", "starting_snapshot_creation")
```

### Tracing Best Practices

- **Span Naming**: Use consistent, descriptive names
- **Tag Important Data**: Include relevant IDs and metadata
- **Log Key Events**: Record state changes and errors
- **Set Appropriate Sampling**: Balance observability with performance
- **Propagate Context**: Ensure trace continuity across services

## Health Checks

### Service Health Endpoints

#### Liveness Probe
```go
func (s *Server) livenessHandler(w http.ResponseWriter, r *http.Request) {
    // Check basic connectivity
    if err := s.db.Ping(); err != nil {
        http.Error(w, "Database unreachable", http.StatusServiceUnavailable)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
}
```

#### Readiness Probe
```go
func (s *Server) readinessHandler(w http.ResponseWriter, r *http.Request) {
    // Check dependencies
    if !s.gossip.Ready() {
        http.Error(w, "Gossip not ready", http.StatusServiceUnavailable)
        return
    }

    if !s.consensus.Ready() {
        http.Error(w, "Consensus not ready", http.StatusServiceUnavailable)
        return
    }

    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Ready"))
}
```

#### Kubernetes Probes
```yaml
apiVersion: v1
kind: Pod
spec:
  containers:
  - name: catalog
    livenessProbe:
      httpGet:
        path: /health/live
        port: 8080
      initialDelaySeconds: 30
      periodSeconds: 10
    readinessProbe:
      httpGet:
        path: /health/ready
        port: 8080
      initialDelaySeconds: 5
      periodSeconds: 5
```

## Metrics Collection

### Custom Metrics

#### Business Metrics
```go
// Snapshot metrics
snapshotCreated := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "decube_snapshot_created_total",
        Help: "Total number of snapshots created",
    },
    []string{"cluster", "user"},
)

snapshotSize := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "decube_snapshot_size_bytes",
        Help: "Size of created snapshots",
        Buckets: prometheus.DefBuckets,
    },
    []string{"cluster"},
)

// Register metrics
prometheus.MustRegister(snapshotCreated, snapshotSize)

// Usage
snapshotCreated.WithLabelValues(clusterID, userID).Inc()
snapshotSize.WithLabelValues(clusterID).Observe(float64(size))
```

#### Performance Metrics
```go
// Request duration
requestDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "decube_http_request_duration_seconds",
        Help: "HTTP request duration in seconds",
        Buckets: []float64{.001, .005, .01, .05, .1, .5, 1, 2.5, 5, 10},
    },
    []string{"method", "endpoint", "status"},
)

// Consensus metrics
consensusRoundDuration := prometheus.NewHistogramVec(
    prometheus.HistogramOpts{
        Name: "decube_consensus_round_duration_seconds",
        Help: "Consensus round duration",
        Buckets: prometheus.DefBuckets,
    },
    []string{"phase"},
)
```

## Monitoring Best Practices

### Alert Design
- **Avoid Alert Fatigue**: Only alert on actionable issues
- **Use Appropriate Severity**: Critical, warning, info levels
- **Include Context**: Provide enough information for diagnosis
- **Set Reasonable Thresholds**: Based on historical data
- **Test Alerts**: Ensure alerts fire correctly

### Dashboard Organization
- **Logical Grouping**: Group related metrics together
- **Consistent Layout**: Use similar layouts across dashboards
- **Clear Labels**: Descriptive titles and axis labels
- **Color Coding**: Use colors to indicate status/severity
- **Time Ranges**: Include multiple time range options

### Data Retention
- **Metrics**: 30-90 days for detailed metrics
- **Logs**: 30 days for application logs, 1 year for audit logs
- **Traces**: 7-30 days depending on volume
- **Cost Optimization**: Use downsampling for older data

### Security Considerations
- **Access Control**: Restrict access to monitoring systems
- **Data Encryption**: Encrypt sensitive log data
- **Network Security**: Secure communication between components
- **Audit Logging**: Log access to monitoring systems

## Troubleshooting Monitoring

### Common Issues

#### Missing Metrics
- Check service endpoints are accessible
- Verify Prometheus scrape configuration
- Ensure metrics are properly registered
- Check firewall rules

#### High Cardinality
- Limit label combinations
- Use appropriate aggregation
- Implement metric expiration
- Consider sampling for high-volume metrics

#### Alert Storm
- Review alert thresholds
- Implement alert grouping
- Use alert dependencies
- Implement alert silencing

#### Performance Impact
- Monitor monitoring system performance
- Optimize queries and dashboards
- Use appropriate scrape intervals
- Implement caching where possible

This monitoring guide provides comprehensive coverage of observability practices for DeCube. Proper monitoring is essential for maintaining system reliability and performance.
