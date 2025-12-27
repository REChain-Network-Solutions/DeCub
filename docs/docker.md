# Docker Deployment Guide

This guide covers deploying DeCube using Docker and Docker Compose.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Docker Compose](#docker-compose)
3. [Custom Images](#custom-images)
4. [Production Deployment](#production-deployment)
5. [Troubleshooting](#troubleshooting)

## Quick Start

### Using Docker Compose

```bash
# Start all services
docker-compose up -d

# Check status
docker-compose ps

# View logs
docker-compose logs -f

# Stop services
docker-compose down
```

### Using Docker

```bash
# Build image
docker build -t decube:latest .

# Run container
docker run -d \
  --name decube \
  -p 8080:8080 \
  -p 9090:9090 \
  -v decube-data:/var/lib/decube \
  decube:latest
```

## Docker Compose

### Development

```bash
# Start development environment
docker-compose up -d

# Rebuild and restart
docker-compose up -d --build

# View logs
docker-compose logs -f decube
```

### Production

```bash
# Use production compose file
docker-compose -f docker-compose.production.yml up -d

# Scale services
docker-compose -f docker-compose.production.yml scale decube=5
```

## Custom Images

### Building Custom Image

```dockerfile
FROM decube:latest

# Add custom configuration
COPY config/custom.yaml /etc/decube/config.yaml

# Add custom scripts
COPY scripts/ /usr/local/bin/
```

### Multi-Architecture Builds

```bash
# Build for multiple architectures
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t decube:latest \
  .
```

## Production Deployment

### Resource Limits

```yaml
deploy:
  resources:
    limits:
      cpus: '4'
      memory: 8G
    reservations:
      cpus: '2'
      memory: 4G
```

### Health Checks

```yaml
healthcheck:
  test: ["CMD", "wget", "--spider", "http://localhost:8080/health"]
  interval: 30s
  timeout: 10s
  retries: 3
  start_period: 40s
```

### Volumes

```yaml
volumes:
  - decube-data:/var/lib/decube
  - decube-config:/etc/decube
  - decube-logs:/var/log/decube
```

## Troubleshooting

### Container Won't Start

```bash
# Check logs
docker logs decube

# Check container status
docker ps -a

# Inspect container
docker inspect decube
```

### Permission Issues

```bash
# Fix volume permissions
docker run --rm -v decube-data:/data \
  alpine chown -R 1000:1000 /data
```

### Network Issues

```bash
# Check network
docker network ls
docker network inspect decube-network

# Test connectivity
docker exec decube wget -O- http://localhost:8080/health
```

## Best Practices

1. **Use Named Volumes**: For persistent data
2. **Set Resource Limits**: Prevent resource exhaustion
3. **Enable Health Checks**: Automatic recovery
4. **Use Secrets**: For sensitive data
5. **Regular Updates**: Keep images updated

## References

- [Deployment Guide](deployment.md)
- [Docker Documentation](https://docs.docker.com/)

---

*Last updated: January 2024*

