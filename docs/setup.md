# DeCube Setup Guide

This guide provides step-by-step instructions for setting up DeCube for development, testing, and production use.

## Prerequisites

### System Requirements

- **Operating System**: Linux (Ubuntu 20.04+), macOS (10.15+), or Windows 10+
- **CPU**: 4+ cores recommended
- **RAM**: 8GB minimum, 16GB recommended
- **Storage**: 50GB free space for development
- **Network**: Stable internet connection

### Required Software

#### Go Development Environment
```bash
# Install Go 1.19+ (Linux/macOS)
curl -fsSL https://golang.org/dl/go1.19.5.linux-amd64.tar.gz | sudo tar -C /usr/local -xzf -

# Add to PATH
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc

# Verify installation
go version
```

#### Docker and Docker Compose
```bash
# Install Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/download/v2.12.0/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Verify installations
docker --version
docker-compose --version
```

#### Git
```bash
# Install Git
sudo apt-get update && sudo apt-get install -y git

# Configure Git
git config --global user.name "Your Name"
git config --global user.email "your.email@example.com"
```

#### Optional Tools
```bash
# Install Make
sudo apt-get install -y build-essential

# Install kubectl (for Kubernetes deployments)
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install Helm (for Kubernetes deployments)
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

## Quick Start

### Clone and Setup
```bash
# Clone the repository
git clone https://github.com/REChain-Network-Solutions/DeCub.git
cd DeCub

# Initialize submodules (if any)
git submodule update --init --recursive

# Copy configuration template
cp config/docker-compose.yml.example docker-compose.yml
cp rechain/config/config.yaml.example rechain/config/config.yaml
cp decube/config/config.yaml.example decube/config/config.yaml
```

### Local Development Setup
```bash
# Start all services
docker-compose up -d

# Wait for services to be healthy
docker-compose ps

# Check service health
curl http://localhost:8080/api/v1/status
curl http://localhost:8082/api/v1/status
curl http://localhost:8081/api/v1/status
```

### Build from Source
```bash
# Build all components
make build

# Or build individual components
cd decube && go build -o bin/decube ./cmd/decube
cd ../rechain && go build -o bin/rechain ./cmd/rechain
cd ../decub-catalog && go build -o bin/decub-catalog .
```

## Development Environment

### IDE Setup

#### Visual Studio Code
1. Install VS Code from https://code.visualstudio.com/
2. Install recommended extensions:
   - Go
   - Docker
   - Kubernetes
   - YAML
   - Markdown Preview Enhanced
3. Open the project folder

#### GoLand
1. Install GoLand from https://www.jetbrains.com/go/
2. Import the project
3. Configure Go SDK (1.19+)

### Environment Variables
```bash
# Create .env file
cat > .env << EOF
# Development environment
DECUB_ENV=development
DECUB_LOG_LEVEL=debug

# Service endpoints
CATALOG_ENDPOINT=http://localhost:8080
GOSSIP_ENDPOINT=http://localhost:8082
GCL_ENDPOINT=http://localhost:8081

# Storage configuration
STORAGE_TYPE=minio
MINIO_ENDPOINT=http://localhost:9000
MINIO_ACCESS_KEY=decube
MINIO_SECRET_KEY=decube-secret

# etcd configuration
ETCD_ENDPOINTS=http://localhost:2379
EOF
```

### Database Setup

#### etcd Setup
```bash
# Start etcd cluster
docker run -d \
  --name etcd \
  -p 2379:2379 \
  -p 2380:2380 \
  quay.io/coreos/etcd:v3.5.5 \
  etcd \
  --advertise-client-urls http://0.0.0.0:2379 \
  --listen-client-urls http://0.0.0.0:2379 \
  --data-dir /etcd-data \
  --name etcd-node-1 \
  --initial-advertise-peer-urls http://0.0.0.0:2380 \
  --listen-peer-urls http://0.0.0.0:2380 \
  --initial-cluster etcd-node-1=http://0.0.0.0:2380 \
  --initial-cluster-state new \
  --auto-compaction-retention=1 \
  --quota-backend-bytes=4294967296 \
  --snapshot-count=50000
```

#### MinIO Setup
```bash
# Start MinIO server
docker run -d \
  --name minio \
  -p 9000:9000 \
  -p 9001:9001 \
  -e "MINIO_ACCESS_KEY=decube" \
  -e "MINIO_SECRET_KEY=decube-secret" \
  -v /tmp/minio-data:/data \
  minio/minio server /data --console-address ":9001"
```

### Testing Setup

#### Unit Tests
```bash
# Run all unit tests
make test

# Run specific package tests
go test ./decub-crypto/...
go test ./decub-gossip/...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Integration Tests
```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make integration-test

# Clean up
docker-compose -f docker-compose.test.yml down
```

#### Benchmark Tests
```bash
# Run benchmarks
go test -bench=. ./decub-crypto/
go test -bench=. ./decub-gossip/

# Memory profiling
go test -bench=. -benchmem ./decub-crypto/
```

## Configuration

### Service Configuration

#### Catalog Service (config/catalog.yaml)
```yaml
server:
  host: 0.0.0.0
  port: 8080
  tls:
    enabled: false
    cert_file: ""
    key_file: ""

storage:
  type: etcd
  etcd:
    endpoints:
      - http://localhost:2379
    dial_timeout: 5s
    request_timeout: 10s

gossip:
  enabled: true
  port: 8082
  peers: []

logging:
  level: info
  format: json
  output: stdout
```

#### Gossip Service (config/gossip.yaml)
```yaml
node:
  id: "gossip-node-1"
  address: "0.0.0.0:8082"

cluster:
  seeds: []
  gossip_interval: 1s
  probe_interval: 5s
  probe_timeout: 500ms

storage:
  merkle_root_refresh: 30s
  delta_retention: 1h

logging:
  level: info
  format: json
```

#### GCL Service (config/gcl.yaml)
```yaml
node:
  id: "gcl-node-1"
  listen_addr: "0.0.0.0:8081"

consensus:
  algorithm: tendermint
  timeout_commit: 1s
  timeout_propose: 3s
  block_size: 1000

validators:
  - id: "validator-1"
    pub_key: "0x..."
    power: 10

storage:
  type: badger
  path: "/tmp/gcl-data"

logging:
  level: info
```

### Docker Compose Configuration

#### docker-compose.yml
```yaml
version: '3.8'

services:
  etcd:
    image: quay.io/coreos/etcd:v3.5.5
    ports:
      - "2379:2379"
      - "2380:2380"
    environment:
      ETCD_NAME: etcd-node-1
      ETCD_DATA_DIR: /etcd-data
      ETCD_LISTEN_PEER_URLS: http://0.0.0.0:2380
      ETCD_LISTEN_CLIENT_URLS: http://0.0.0.0:2379
      ETCD_INITIAL_ADVERTISE_PEER_URLS: http://etcd:2380
      ETCD_ADVERTISE_CLIENT_URLS: http://etcd:2379
      ETCD_INITIAL_CLUSTER: etcd-node-1=http://etcd:2380
      ETCD_INITIAL_CLUSTER_STATE: new
    volumes:
      - etcd-data:/etcd-data

  minio:
    image: minio/minio:latest
    ports:
      - "9000:9000"
      - "9001:9001"
    environment:
      MINIO_ACCESS_KEY: decube
      MINIO_SECRET_KEY: decube-secret
    volumes:
      - minio-data:/data
    command: server /data --console-address ":9001"

  catalog:
    build: ./decub-catalog
    ports:
      - "8080:8080"
    depends_on:
      - etcd
    environment:
      ETCD_ENDPOINTS: http://etcd:2379

  gossip:
    build: ./decub-gossip
    ports:
      - "8082:8082"
    depends_on:
      - catalog

  gcl:
    build: ./decub-gcl/go
    ports:
      - "8081:8081"
    depends_on:
      - etcd

volumes:
  etcd-data:
  minio-data:
```

## CLI Tools Setup

### decubectl Installation
```bash
# Build CLI tool
cd cmd/decubectl
go build -o decubectl .

# Install globally
sudo mv decubectl /usr/local/bin/

# Verify installation
decubectl version
```

### rechainctl Installation
```bash
# Build CLI tool
cd cmd/rechainctl
go build -o rechainctl .

# Install globally
sudo mv rechainctl /usr/local/bin/

# Verify installation
rechainctl version
```

### CLI Configuration
```bash
# Initialize CLI configuration
decubectl config init

# Set cluster endpoint
decubectl config set-cluster prod --server https://api.decube.example.com

# Set default context
decubectl config use-context prod
```

## Production Setup

### Security Hardening

#### TLS Configuration
```bash
# Generate self-signed certificates (for testing)
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes

# Configure services to use TLS
# Update service configurations with cert_file and key_file paths
```

#### Firewall Configuration
```bash
# UFW example (Ubuntu)
sudo ufw enable
sudo ufw allow 22/tcp
sudo ufw allow 8080/tcp
sudo ufw allow 8081/tcp
sudo ufw allow 8082/tcp
sudo ufw allow 9000/tcp
sudo ufw allow 2379/tcp
sudo ufw allow 2380/tcp
```

#### User Management
```bash
# Create dedicated user
sudo useradd -r -s /bin/false decube

# Set proper permissions
sudo chown -R decube:decube /var/lib/decube
sudo chmod 700 /var/lib/decube
```

### Monitoring Setup

#### Prometheus Configuration
```yaml
# prometheus.yml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: 'decube-catalog'
    static_configs:
      - targets: ['localhost:8080']
  - job_name: 'decube-gossip'
    static_configs:
      - targets: ['localhost:8082']
  - job_name: 'decube-gcl'
    static_configs:
      - targets: ['localhost:8081']
```

#### Grafana Dashboard
```bash
# Import DeCube dashboard
# Dashboard ID: 12345 (example)
```

### Backup Configuration

#### Automated Backups
```bash
# Create backup script
cat > backup.sh << 'EOF'
#!/bin/bash
BACKUP_DIR="/var/backups/decube"
TIMESTAMP=$(date +%Y%m%d_%H%M%S)

# Backup etcd data
etcdctl snapshot save $BACKUP_DIR/etcd_$TIMESTAMP.db

# Backup MinIO data
mc mirror minio/decube $BACKUP_DIR/minio_$TIMESTAMP/

# Backup configurations
tar -czf $BACKUP_DIR/config_$TIMESTAMP.tar.gz /etc/decube/
EOF

chmod +x backup.sh

# Schedule with cron
echo "0 2 * * * /path/to/backup.sh" | crontab -
```

## Troubleshooting

### Common Issues

#### Service Won't Start
```bash
# Check logs
docker-compose logs <service-name>

# Check dependencies
docker-compose ps

# Verify configuration
decubectl config validate
```

#### Connection Refused
```bash
# Check network connectivity
telnet localhost 8080

# Verify service ports
netstat -tlnp | grep :8080

# Check firewall rules
sudo ufw status
```

#### Performance Issues
```bash
# Monitor resource usage
docker stats

# Check system resources
top
df -h
free -h

# Profile application
go tool pprof http://localhost:8080/debug/pprof/profile
```

### Debug Mode
```bash
# Enable debug logging
export DECUB_LOG_LEVEL=debug

# Start services with debug flags
docker-compose up --scale catalog=1 --scale gossip=1 --scale gcl=1

# Run with verbose output
decubectl --verbose command
```

## Next Steps

After completing the setup:

1. **Run the test suite**: `make test`
2. **Create your first snapshot**: `decubectl snapshot create test-snapshot`
3. **Explore the API**: Visit http://localhost:8080/docs
4. **Check the documentation**: See docs/ for detailed guides
5. **Join the community**: Visit our GitHub discussions

For production deployments, refer to the [deployment guide](deployment.md) for advanced configuration options.
