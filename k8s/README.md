# Kubernetes Manifests

Kubernetes manifests for deploying DeCube.

## Files

- `deployment.yaml` - Main deployment
- `service.yaml` - Service definition
- `configmap.yaml` - Configuration
- `pvc.yaml` - Persistent volume claim
- `namespace.yaml` - Namespace (create if needed)

## Quick Start

```bash
# Create namespace
kubectl create namespace decube

# Apply manifests
kubectl apply -f namespace.yaml
kubectl apply -f configmap.yaml
kubectl apply -f pvc.yaml
kubectl apply -f deployment.yaml
kubectl apply -f service.yaml

# Check status
kubectl get pods -n decube
kubectl get svc -n decube
```

## Customization

### Update Configuration

Edit `configmap.yaml` and apply:

```bash
kubectl apply -f configmap.yaml
kubectl rollout restart deployment/decube -n decube
```

### Scale Deployment

```bash
kubectl scale deployment decube --replicas=5 -n decube
```

### Update Image

```bash
kubectl set image deployment/decube decube=decube/decube:v0.2.0 -n decube
```

## Prerequisites

- Kubernetes cluster (1.19+)
- Storage class named `fast-ssd` (or update `pvc.yaml`)
- MinIO or S3-compatible storage for CAS

## See Also

- [Kubernetes Deployment Guide](../docs/kubernetes.md)
- [Deployment Guide](../docs/deployment.md)

