# Integration Guide

This guide covers integrating DeCube with other systems and services.

## Table of Contents

1. [API Integration](#api-integration)
2. [SDK Usage](#sdk-usage)
3. [Kubernetes Integration](#kubernetes-integration)
4. [Monitoring Integration](#monitoring-integration)
5. [CI/CD Integration](#cicd-integration)
6. [Third-Party Integrations](#third-party-integrations)

## API Integration

### REST API

#### Authentication

```bash
# Get authentication token
TOKEN=$(curl -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}' \
  | jq -r '.token')

# Use token in requests
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/catalog/snapshots
```

#### Creating Resources

```bash
# Create snapshot
curl -X POST http://localhost:8080/snapshots \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "id": "snapshot-001",
    "metadata": {
      "size": 1073741824,
      "cluster": "cluster-a"
    }
  }'
```

#### Querying Resources

```bash
# List snapshots
curl http://localhost:8080/catalog/snapshots

# Get specific snapshot
curl http://localhost:8080/catalog/snapshots/snapshot-001

# Query with filters
curl "http://localhost:8080/catalog/query?type=snapshots&cluster=cluster-a"
```

### gRPC API

#### Go Client Example

```go
package main

import (
    "context"
    "log"
    
    "google.golang.org/grpc"
    pb "github.com/REChain-Network-Solutions/DeCub/decube/api/proto"
)

func main() {
    conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()
    
    client := pb.NewDeCubeClient(conn)
    
    ctx := context.Background()
    req := &pb.CreateSnapshotRequest{
        Id: "snapshot-001",
        Metadata: map[string]string{
            "cluster": "cluster-a",
        },
    }
    
    resp, err := client.CreateSnapshot(ctx, req)
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created snapshot: %s", resp.Id)
}
```

#### Python Client Example

```python
import grpc
from decube import decube_pb2, decube_pb2_grpc

channel = grpc.insecure_channel('localhost:9090')
stub = decube_pb2_grpc.DeCubeStub(channel)

request = decube_pb2.CreateSnapshotRequest(
    id='snapshot-001',
    metadata={'cluster': 'cluster-a'}
)

response = stub.CreateSnapshot(request)
print(f"Created snapshot: {response.id}")
```

## SDK Usage

### Go SDK

#### Installation

```bash
go get github.com/REChain-Network-Solutions/DeCub/sdk/go
```

#### Usage

```go
package main

import (
    "context"
    "log"
    
    "github.com/REChain-Network-Solutions/DeCub/sdk/go/decube"
)

func main() {
    client, err := decube.NewClient("http://localhost:8080")
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Create snapshot
    snapshot, err := client.Snapshots.Create(ctx, &decube.SnapshotRequest{
        ID: "snapshot-001",
        Metadata: map[string]interface{}{
            "cluster": "cluster-a",
        },
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Printf("Created snapshot: %s", snapshot.ID)
    
    // Query catalog
    snapshots, err := client.Catalog.ListSnapshots(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, s := range snapshots {
        log.Printf("Snapshot: %s", s.ID)
    }
}
```

### Python SDK

#### Installation

```bash
pip install decube-sdk
```

#### Usage

```python
from decube import Client

client = Client("http://localhost:8080")

# Create snapshot
snapshot = client.snapshots.create(
    id="snapshot-001",
    metadata={"cluster": "cluster-a"}
)

# Query catalog
snapshots = client.catalog.list_snapshots()
for s in snapshots:
    print(f"Snapshot: {s.id}")
```

## Kubernetes Integration

### Deploying with Helm

```bash
# Add Helm repository
helm repo add decube https://charts.decube.io
helm repo update

# Install DeCube
helm install decube decube/decube \
  --set cluster.id=cluster-001 \
  --set storage.cas.endpoint=http://minio:9000
```

### Custom Resource Definitions

```yaml
apiVersion: decube.io/v1
kind: Snapshot
metadata:
  name: snapshot-001
spec:
  id: snapshot-001
  metadata:
    cluster: cluster-a
    size: 1073741824
```

### Operator Usage

```bash
# Apply snapshot resource
kubectl apply -f snapshot.yaml

# Check status
kubectl get snapshots

# Describe snapshot
kubectl describe snapshot snapshot-001
```

## Monitoring Integration

### Prometheus

#### Configuration

```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'decube'
    static_configs:
      - targets: ['localhost:9090']
    metrics_path: '/metrics'
```

#### Querying Metrics

```promql
# Request rate
rate(decube_http_requests_total[5m])

# Response time
histogram_quantile(0.95, decube_http_request_duration_seconds_bucket)

# Error rate
rate(decube_http_requests_total{status=~"5.."}[5m])
```

### Grafana Dashboards

```json
{
  "dashboard": {
    "title": "DeCube Metrics",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "rate(decube_http_requests_total[5m])"
          }
        ]
      }
    ]
  }
}
```

### Datadog Integration

```yaml
# datadog.yaml
logs_enabled: true
logs_config:
  container_collect_all: true
  container_exclude: "name:datadog-agent"
```

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy to DeCube

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Create Snapshot
        run: |
          curl -X POST ${{ secrets.DECUBE_ENDPOINT }}/snapshots \
            -H "Authorization: Bearer ${{ secrets.DECUBE_TOKEN }}" \
            -H "Content-Type: application/json" \
            -d '{"id":"deploy-${{ github.sha }}"}'
```

### GitLab CI

```yaml
deploy:
  stage: deploy
  script:
    - curl -X POST $DECUBE_ENDPOINT/snapshots \
        -H "Authorization: Bearer $DECUBE_TOKEN" \
        -H "Content-Type: application/json" \
        -d "{\"id\":\"deploy-$CI_COMMIT_SHA\"}"
```

### Jenkins Pipeline

```groovy
pipeline {
    agent any
    stages {
        stage('Deploy') {
            steps {
                sh '''
                    curl -X POST ${DECUBE_ENDPOINT}/snapshots \
                      -H "Authorization: Bearer ${DECUBE_TOKEN}" \
                      -H "Content-Type: application/json" \
                      -d "{\"id\":\"deploy-${BUILD_NUMBER}\"}"
                '''
            }
        }
    }
}
```

## Third-Party Integrations

### Terraform Provider

```hcl
provider "decube" {
  endpoint = "http://localhost:8080"
  token    = var.decube_token
}

resource "decube_snapshot" "example" {
  id = "snapshot-001"
  metadata = {
    cluster = "cluster-a"
  }
}
```

### Ansible Module

```yaml
- name: Create snapshot
  decube_snapshot:
    endpoint: "http://localhost:8080"
    token: "{{ decube_token }}"
    id: "snapshot-001"
    metadata:
      cluster: "cluster-a"
  register: result
```

### Pulumi Integration

```typescript
import * as decube from "@decube/pulumi-decube";

const snapshot = new decube.Snapshot("snapshot-001", {
    id: "snapshot-001",
    metadata: {
        cluster: "cluster-a"
    }
});
```

## Webhook Integration

### Configuring Webhooks

```yaml
webhooks:
  - url: "https://example.com/webhook"
    events:
      - snapshot.created
      - snapshot.deleted
    secret: "webhook-secret"
```

### Webhook Payload

```json
{
  "event": "snapshot.created",
  "timestamp": "2024-01-15T10:30:00Z",
  "data": {
    "id": "snapshot-001",
    "metadata": {
      "cluster": "cluster-a"
    }
  }
}
```

## Best Practices

1. **Use Connection Pooling**: Reuse connections for better performance
2. **Handle Errors**: Implement retry logic with exponential backoff
3. **Monitor Integration**: Set up monitoring for integrated systems
4. **Secure Communication**: Use TLS for all API calls
5. **Rate Limiting**: Respect rate limits and implement backoff

## Troubleshooting

### Connection Issues

- Check network connectivity
- Verify endpoint URLs
- Check firewall rules
- Review TLS certificates

### Authentication Errors

- Verify token validity
- Check token expiration
- Review permissions
- Validate credentials

### Performance Issues

- Review connection pooling
- Check request batching
- Monitor resource usage
- Profile integration code

## References

- [API Documentation](api.md)
- [Getting Started Guide](getting-started.md)
- [Deployment Guide](deployment.md)
- [Monitoring Guide](monitoring.md)

---

*Last updated: January 2024*

