# REChain - Decentralized Multi-Layer Storage and Compute Fabric

REChain is a production-ready decentralized system that provides a multi-layer storage and compute fabric with the following components:

- **Local RAFT Storage**: High-performance local storage with RAFT consensus
- **Global BFT Ledger**: Byzantine Fault Tolerant global ledger for cross-cluster coordination
- **Gossip + CRDT Sync**: Epidemic broadcast with Conflict-Free Replicated Data Types for metadata synchronization
- **CAS Object Store**: Content-Addressed Storage with S3 compatibility and Merkle proofs
- **Merkle Proofs**: Cryptographic verification of data integrity and state

## Features

### 🔒 Security First
- mTLS for all network communications
- Client-side AES-256-GCM encryption
- RSA-PSS transaction signing
- HSM integration support
- Comprehensive audit logging

### 🚀 High Performance
- Sub-second block finalization
- 1000+ TPS throughput
- Horizontal scaling support
- Optimized storage engines
- Connection pooling and caching

### 🛡️ Production Ready
- Comprehensive monitoring with Prometheus
- Structured logging with multiple outputs
- Automated backups and disaster recovery
- Rolling updates with zero downtime
- Multi-region deployment support

### 🔧 Developer Friendly
- REST and gRPC APIs
- Comprehensive SDKs
- CLI tools with shell completion
- Extensive documentation
- Development and production configs

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │    │     REChain     │    │   Storage       │
│                 │    │                 │    │                 │
│  ┌────────────┐ │    │  ┌────────────┐ │    │  ┌────────────┐ │
│  │   REST     │◄┼────┼─►│   API      │◄┼────┼─►│   CAS      │ │
│  │   API      │ │    │  │   Server   │ │    │  │   Store    │ │
│  └────────────┘ │    │  └────────────┘ │    │  └────────────┘ │
│                 │    │                 │    │                 │
│  ┌────────────┐ │    │  ┌────────────┐ │    │  ┌────────────┐ │
│  │   gRPC     │◄┼────┼─►│  Consensus │◄┼────┼─►│   Local    │ │
│  │   API      │ │    │  │   Engine   │ │    │  │   Store    │ │
│  └────────────┘ │    │  └────────────┘ │    │  └────────────┘ │
└─────────────────┘    │                 │    └─────────────────┘
                       │  ┌────────────┐ │
                       │  │   Gossip   │ │
                       │  │   Protocol │ │
                       │  └────────────┘ │
                       └─────────────────┘
```

## Quick Start

### Prerequisites

- Docker and Docker Compose
- Go 1.19+ (for development)
- 4GB RAM minimum
- 10GB disk space

### Single Node Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/rechain/rechain.git
   cd rechain
   ```

2. **Start the services**
   ```bash
   docker-compose up -d
   ```

3. **Check health**
   ```bash
   curl http://localhost:1317/health
   ```

4. **Store an object**
   ```bash
   echo "Hello, REChain!" | curl -X POST \
     -H "Content-Type: text/plain" \
     --data-binary @- \
     http://localhost:1317/cas/objects
   ```

### Multi-Node Setup

1. **Scale up nodes**
   ```bash
   docker-compose up -d --scale rechain-node=3
   ```

2. **Check cluster status**
   ```bash
   curl http://localhost:1317/node/peers
   ```

## Configuration

REChain uses YAML configuration files. See `config/config.yaml` for all available options.

### Environment Variables

Override configuration with environment variables:

```bash
export RECHAIN_NODE_ID=my-node
export RECHAIN_LOG_LEVEL=debug
export RECHAIN_API_REST_ADDRESS=0.0.0.0:8080
```

### TLS Configuration

Enable TLS by providing certificates:

```yaml
security:
  tls_enabled: true
  cert_file: ./certs/server.crt
  key_file: ./certs/server.key
  ca_file: ./certs/ca.crt
```

## API Usage

### REST API

#### Store Object
```bash
curl -X POST \
  -H "Content-Type: application/octet-stream" \
  --data-binary @file.txt \
  http://localhost:1317/cas/objects
```

#### Retrieve Object
```bash
curl http://localhost:1317/cas/objects/{cid} -o retrieved-file.txt
```

#### Submit Transaction
```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"type": "data", "payload": {"key": "value"}}' \
  http://localhost:1317/txs
```

#### Get Block
```bash
curl http://localhost:1317/blocks/1
```

### gRPC API

```go
conn, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

client := proto.NewRechainServiceClient(conn)

// Get node info
resp, err := client.GetNodeInfo(context.Background(), &proto.NodeInfoRequest{})
```

## CLI Tools

### Install CLI

```bash
go install ./cmd/rechainctl
```

### Usage

```bash
# Get node info
rechainctl node info

# Store file
rechainctl cas store myfile.txt

# Get transaction
rechainctl tx get tx-123

# List blocks
rechainctl block list --limit 10
```

## Monitoring

### Prometheus Metrics

Access metrics at `http://localhost:9091/metrics`

### Grafana Dashboards

Access Grafana at `http://localhost:3000` (admin/admin)

### Health Checks

```bash
# API health
curl http://localhost:1317/health

# Consensus health
curl http://localhost:1317/consensus/state
```

## Development

### Build from Source

```bash
# Clone repository
git clone https://github.com/rechain/rechain.git
cd rechain

# Install dependencies
go mod download

# Build
go build -o bin/rechain ./cmd/rechain

# Run
./bin/rechain --config ./config/config.yaml
```

### Run Tests

```bash
# Unit tests
go test ./...

# Integration tests
go test -tags=integration ./tests/...

# Benchmarks
go test -bench=. ./...
```

### Development Setup

```bash
# Start development environment
docker-compose -f docker-compose.dev.yml up -d

# Run with hot reload
air
```

## Security

### Best Practices

- Always use TLS in production
- Rotate encryption keys regularly
- Enable audit logging
- Use HSM for key management
- Implement proper access controls

### Security Audits

REChain undergoes regular security audits. See `SECURITY.md` for details.

## Performance

### Benchmarks

- **Throughput**: 1000+ TPS
- **Latency**: <100ms p95
- **Storage**: 10GB+ per node
- **Network**: 1Gbps+ recommended

### Scaling

- Horizontal scaling with multiple nodes
- Load balancing across regions
- Automatic failover and recovery
- Dynamic resource allocation

## Deployment

### Production Deployment

```bash
# Build production image
docker build -t rechain:latest .

# Deploy with Kubernetes
kubectl apply -f k8s/

# Or use Docker Compose
docker-compose -f docker-compose.prod.yml up -d
```

### Backup and Recovery

```bash
# Create backup
rechainctl backup create

# Restore from backup
rechainctl backup restore backup-2023-12-01.tar.gz
```

## Contributing

We welcome contributions! Please see `CONTRIBUTING.md` for guidelines.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make changes with tests
4. Submit a pull request
5. Wait for review and merge

## License

REChain is licensed under the Apache 2.0 License. See `LICENSE` for details.

## Support

- **Documentation**: https://docs.rechain.io
- **Issues**: https://github.com/rechain/rechain/issues
- **Discussions**: https://github.com/rechain/rechain/discussions
- **Slack**: https://rechain.slack.com

## Roadmap

### Phase 1 (Current)
- ✅ Core consensus implementation
- ✅ Gossip protocol with CRDTs
- ✅ CAS object storage
- ✅ Security and encryption
- ✅ REST and gRPC APIs

### Phase 2 (Next)
- 🔄 Smart contracts support
- 🔄 Cross-chain interoperability
- 🔄 Advanced CRDT types
- 🔄 Performance optimizations
- 🔄 Enterprise features

### Phase 3 (Future)
- 🔄 Decentralized identity
- 🔄 Privacy-preserving computation
- 🔄 AI/ML integration
- 🔄 Quantum-resistant cryptography

---

**REChain** - Building the future of decentralized storage and compute.
