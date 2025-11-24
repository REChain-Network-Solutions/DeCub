# DeCube Local Control-Plane

A minimal local control-plane implementation using embedded etcd for DeCube, providing strong consistency, REST/gRPC APIs, and peer replication.

## Features

- **Embedded etcd**: RAFT-based key-value store with strong consistency
- **REST API**: Full CRUD operations for pods, snapshots, and leases
- **gRPC API**: High-performance service with protobuf definitions
- **Peer Replication**: gRPC-based state synchronization between nodes
- **Snapshot/Restore**: etcd snapshot creation and WAL replay
- **Multi-node Deployment**: Docker Compose setup for 3-node cluster

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   DeCube Node   │    │   DeCube Node   │    │   DeCube Node   │
│                 │    │                 │    │                 │
│  ┌──────────┐   │    │  ┌──────────┐   │    │  ┌──────────┐   │
│  │  REST    │   │    │  │  REST    │   │    │  │  REST    │   │
│  │  API     │   │    │  │  API     │   │    │  │  API     │   │
│  └──────────┘   │    │  └──────────┘   │    │  └──────────┘   │
│  ┌──────────┐   │    │  ┌──────────┐   │    │  ┌──────────┐   │
│  │  gRPC    │◄─►│    │  │  gRPC    │◄─►│    │  │  gRPC    │   │
│  │  API     │   │    │  │  API     │   │    │  │  API     │   │
│  └──────────┘   │    │  └──────────┘   │    │  └──────────┘   │
│  ┌──────────┐   │    │  ┌──────────┐   │    │  ┌──────────┐   │
│  │ Embedded │   │    │  │ Embedded │   │    │  │ Embedded │   │
│  │   etcd   │◄─►│    │  │   etcd   │◄─►│    │  │   etcd   │   │
│  │  (RAFT)  │   │    │  │  (RAFT)  │   │    │  │  (RAFT)  │   │
│  └──────────┘   │    │  └──────────┘   │    │  └──────────┘   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Quick Start

### Single Node

```bash
# Clone and build
git clone https://github.com/decube/decube.git
cd decube
go mod download
go build -o decube ./cmd/decube

# Run
./decube --config ./config/config.yaml
```

### Multi-Node Cluster

```bash
# Start 3-node cluster
docker-compose up -d

# Check cluster health
curl http://localhost:8080/health
curl http://localhost:8081/health
curl http://localhost:8082/health
```

## API Endpoints

### REST API

#### Pods
- `GET /api/v1/pods` - List pods
- `POST /api/v1/pods` - Create pod
- `GET /api/v1/pods/{name}` - Get pod
- `PUT /api/v1/pods/{name}` - Update pod
- `DELETE /api/v1/pods/{name}` - Delete pod

#### Snapshots
- `GET /api/v1/snapshots` - List snapshots
- `POST /api/v1/snapshots` - Create snapshot
- `GET /api/v1/snapshots/{id}` - Get snapshot
- `POST /api/v1/snapshots/{id}/restore` - Restore snapshot
- `DELETE /api/v1/snapshots/{id}` - Delete snapshot

#### Leases
- `GET /api/v1/leases` - List leases
- `POST /api/v1/leases` - Create lease
- `GET /api/v1/leases/{id}` - Get lease
- `POST /api/v1/leases/{id}/renew` - Renew lease
- `DELETE /api/v1/leases/{id}` - Delete lease

#### Node Info
- `GET /node/info` - Get node information
- `GET /health` - Health check

### gRPC API

Full protobuf definitions available in `api/proto/decube.proto`.

## Configuration

### YAML Configuration

```yaml
# Node configuration
node:
  id: "node-1"
  data_dir: "/var/lib/decube"
  listen_address: "0.0.0.0:2379"
  peer_addresses:
    - "node-1:2380"
    - "node-2:2380"
    - "node-3:2380"

# etcd configuration
etcd:
  name: "node-1"
  data_dir: "/var/lib/decube/etcd"
  wal_dir: "/var/lib/decube/etcd/wal"
  snapshot_count: 10000
  heartbeat_interval: 100
  election_timeout: 1000
  max_snapshots: 5
  max_wals: 5
  auto_compaction_retention: "1h"
  quota_backend_bytes: 4294967296

# API configuration
api:
  rest:
    enabled: true
    address: "0.0.0.0:8080"
    cors_origins:
      - "*"
  grpc:
    enabled: true
    address: "0.0.0.0:9090"

# Replication configuration
replication:
  enabled: true
  peer_timeout: 5s
  retry_interval: 1s
  max_retries: 3

# Snapshot configuration
snapshot:
  enabled: true
  interval: 1h
  retention_count: 10
  compression: true
```

### Environment Variables

All configuration values can be overridden with environment variables prefixed with `DECUBE_`:

```bash
export DECUBE_NODE_ID=node-1
export DECUBE_API_REST_ADDRESS=0.0.0.0:8080
export DECUBE_ETCD_NAME=node-1
```

## Deployment

### Systemd Service

```bash
# Install
sudo cp decube.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable decube
sudo systemctl start decube

# Check status
sudo systemctl status decube
sudo journalctl -u decube -f
```

### Docker

```bash
# Build
docker build -t decube:latest .

# Run single node
docker run -p 8080:8080 -p 9090:9090 -p 2379:2379 -p 2380:2380 \
  -v $(pwd)/config:/var/lib/decube/config:ro \
  -v decube-data:/var/lib/decube \
  decube:latest
```

### Kubernetes

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: decube
spec:
  serviceName: decube
  replicas: 3
  selector:
    matchLabels:
      app: decube
  template:
    metadata:
      labels:
        app: decube
    spec:
      containers:
      - name: decube
        image: decube:latest
        ports:
        - containerPort: 8080
          name: rest
        - containerPort: 9090
          name: grpc
        - containerPort: 2379
          name: etcd-client
        - containerPort: 2380
          name: etcd-peer
        volumeMounts:
        - name: data
          mountPath: /var/lib/decube
        env:
        - name: DECUBE_NODE_ID
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
  volumeClaimTemplates:
  - metadata:
    name: data
  spec:
    accessModes: ["ReadWriteOnce"]
    resources:
      requests:
        storage: 10Gi
```

## Examples

### Create a Pod

```bash
curl -X POST http://localhost:8080/api/v1/pods \
  -H "Content-Type: application/json" \
  -d '{
    "name": "nginx-pod",
    "namespace": "default",
    "status": "running",
    "node_name": "node-1",
    "labels": {
      "app": "nginx",
      "version": "1.21"
    }
  }'
```

### Create a Lease

```bash
curl -X POST http://localhost:8080/api/v1/leases \
  -H "Content-Type: application/json" \
  -d '{
    "holder": "scheduler-1",
    "ttl_seconds": 30,
    "metadata": {
      "purpose": "leader-election"
    }
  }'
```

### Create a Snapshot

```bash
curl -X POST http://localhost:8080/api/v1/snapshots \
  -H "Content-Type: application/json" \
  -d '{
    "name": "backup-2023-12-01",
    "metadata": {
      "created_by": "admin",
      "purpose": "weekly-backup"
    }
  }'
```

## Monitoring

### Health Checks

```bash
# REST health check
curl http://localhost:8080/health

# Response
{
  "status": "healthy",
  "timestamp": "2023-12-01T10:00:00Z",
  "is_leader": true
}
```

### Metrics

DeCube exposes Prometheus metrics on the REST API endpoint `/metrics` (when enabled).

## Security

### TLS Configuration

Enable TLS by setting the following in configuration:

```yaml
security:
  tls_enabled: true
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
  ca_file: "/path/to/ca.pem"
```

### Authentication

Currently, DeCube does not implement authentication. For production use, consider:

- Mutual TLS authentication
- JWT-based authentication
- Integration with external auth providers

## Backup and Recovery

### Manual Snapshot

```bash
# Create snapshot
curl -X POST http://localhost:8080/api/v1/snapshots \
  -H "Content-Type: application/json" \
  -d '{"name": "manual-backup"}'

# List snapshots
curl http://localhost:8080/api/v1/snapshots
```

### Automated Snapshots

Configure automated snapshots in the configuration:

```yaml
snapshot:
  enabled: true
  interval: 1h
  retention_count: 10
  compression: true
```

## Troubleshooting

### Common Issues

1. **etcd cluster not forming**
   - Check peer addresses in configuration
   - Ensure firewall allows etcd peer communication (2380)
   - Verify node names are unique

2. **API server not responding**
   - Check if ports are available
   - Verify configuration addresses
   - Check logs for binding errors

3. **Data not persisting**
   - Ensure data directory has proper permissions
   - Check disk space availability
   - Verify volume mounts in Docker

### Logs

```bash
# Docker logs
docker logs decube-node1

# Systemd logs
journalctl -u decube -f

# Application logs (when file logging enabled)
tail -f /var/log/decube/decube.log
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request

## License

This project is licensed under the Apache License 2.0.
