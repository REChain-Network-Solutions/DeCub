# Getting Started with DeCube

Welcome to DeCube! This guide will help you get started with the decentralized compute platform.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Quick Start](#quick-start)
4. [Configuration](#configuration)
5. [Next Steps](#next-steps)

## Prerequisites

Before you begin, ensure you have the following installed:

- **Go 1.19+**: [Download Go](https://golang.org/dl/)
- **Docker & Docker Compose**: [Install Docker](https://docs.docker.com/get-docker/)
- **Git**: [Install Git](https://git-scm.com/downloads)
- **Make**: Usually pre-installed on Linux/macOS, [install on Windows](https://www.gnu.org/software/make/)

### System Requirements

- **CPU**: 2+ cores recommended
- **RAM**: 4GB minimum, 8GB+ recommended
- **Disk**: 10GB+ free space
- **Network**: Internet connection for downloading dependencies

## Installation

### Option 1: From Source

```bash
# Clone the repository
git clone https://github.com/REChain-Network-Solutions/DeCub.git
cd DeCub

# Run setup script
chmod +x scripts/setup-dev.sh
./scripts/setup-dev.sh

# Or use Make
make setup
```

### Option 2: Using Docker

```bash
# Clone the repository
git clone https://github.com/REChain-Network-Solutions/DeCub.git
cd DeCub

# Start all services
docker-compose up -d

# Check service status
docker-compose ps
```

### Option 3: Pre-built Binaries

Download pre-built binaries from the [Releases](https://github.com/REChain-Network-Solutions/DeCub/releases) page.

## Quick Start

### 1. Start Services

```bash
# Using Docker Compose (recommended for beginners)
docker-compose up -d

# Or build and run locally
make build
make run
```

### 2. Verify Installation

```bash
# Check if services are running
curl http://localhost:8080/health

# Or check Docker containers
docker-compose ps
```

### 3. Create Your First Snapshot

```bash
# Using the CLI
./bin/decub-snapshot create my-first-snapshot /data/etcd /data/volumes

# Or using the API
curl -X POST http://localhost:8080/snapshots \
  -H "Content-Type: application/json" \
  -d '{
    "id": "my-first-snapshot",
    "metadata": {
      "size": 1073741824,
      "created": "2024-01-15T10:30:00Z",
      "cluster": "cluster-a"
    }
  }'
```

### 4. Query the Catalog

```bash
# List all snapshots
curl http://localhost:8080/catalog/snapshots

# Query specific snapshot
curl http://localhost:8080/catalog/snapshots/my-first-snapshot
```

## Configuration

### Basic Configuration

1. Copy the example configuration:
```bash
cp config/config.example.yaml config/config.yaml
```

2. Edit `config/config.yaml` with your settings:
```yaml
cluster:
  id: "my-cluster"
  raft:
    bind_addr: "0.0.0.0:7000"
    data_dir: "/var/lib/decube/raft"

storage:
  cas:
    endpoint: "http://cas:9000"
    access_key: "your-access-key"
    secret_key: "your-secret-key"
```

3. Start services with your configuration:
```bash
docker-compose up -d
```

### Environment Variables

You can also configure DeCube using environment variables:

```bash
export DECUBE_CLUSTER_ID="my-cluster"
export DECUBE_RAFT_BIND_ADDR="0.0.0.0:7000"
export DECUBE_STORAGE_CAS_ENDPOINT="http://cas:9000"

# Start services
docker-compose up -d
```

## Next Steps

Now that you have DeCube running, explore:

1. **API Documentation**: See [API Reference](api.md) for detailed API documentation
2. **Architecture**: Understand the system with [Architecture Guide](architecture.md)
3. **Examples**: Try the [examples](../examples/) to learn common patterns
4. **Deployment**: Follow the [Deployment Guide](deployment.md) for production setups
5. **Monitoring**: Set up monitoring with [Monitoring Guide](monitoring.md)

## Troubleshooting

### Services Won't Start

- Check Docker is running: `docker ps`
- Check ports are available: `netstat -an | grep -E '8080|7000|8000'`
- Review logs: `docker-compose logs`

### Connection Errors

- Verify firewall settings
- Check network connectivity between nodes
- Review TLS certificate configuration

### Performance Issues

- Check system resources: `docker stats`
- Review [Performance Guide](performance.md)
- Adjust configuration parameters

## Getting Help

- **Documentation**: Browse the [docs](.) directory
- **Issues**: Report bugs on [GitHub Issues](https://github.com/REChain-Network-Solutions/DeCub/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/REChain-Network-Solutions/DeCub/discussions)
- **Security**: Email security@decube.io for security issues

## Contributing

We welcome contributions! See [CONTRIBUTING.md](../CONTRIBUTING.md) for guidelines.

---

Happy computing with DeCube! 🚀

