# Kubernetes Deployment Guide

This guide covers deploying DeCube on Kubernetes.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Helm Installation](#helm-installation)
4. [Manual Installation](#manual-installation)
5. [Configuration](#configuration)
6. [Scaling](#scaling)
7. [Monitoring](#monitoring)
8. [Troubleshooting](#troubleshooting)

## Prerequisites

- Kubernetes cluster (1.19+)
- kubectl configured
- Helm 3.x (optional, for Helm installation)
- Storage class for persistent volumes

## Quick Start

### Using Helm

```bash
# Add Helm repository
helm repo add decube https://charts.decube.io
helm repo update

# Install DeCube
helm install decube decube/decube \
  --set cluster.id=cluster-001 \
  --set storage.cas.endpoint=http://minio:9000
```

### Using kubectl

```bash
# Apply manifests
kubectl apply -f k8s-manifest.yaml

# Check status
kubectl get pods -l app=decube
```

## Helm Installation

### Values Configuration

```yaml
# values.yaml
cluster:
  id: "cluster-001"
  replicas: 3

storage:
  cas:
    endpoint: "http://minio:9000"
    accessKey: "minioadmin"
    secretKey: "minioadmin"

gcl:
  enabled: true
  replicas: 3

gossip:
  enabled: true
  replicas: 3

resources:
  requests:
    cpu: "500m"
    memory: "512Mi"
  limits:
    cpu: "2"
    memory: "2Gi"

persistence:
  enabled: true
  size: "10Gi"
  storageClass: "fast-ssd"
```

### Install with Custom Values

```bash
helm install decube decube/decube \
  -f values.yaml
```

### Upgrade

```bash
helm upgrade decube decube/decube \
  -f values.yaml
```

## Manual Installation

### Namespace

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: decube
```

### ConfigMap

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: decube-config
  namespace: decube
data:
  config.yaml: |
    cluster:
      id: "cluster-001"
    storage:
      cas:
        endpoint: "http://minio:9000"
```

### Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: decube
  namespace: decube
spec:
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
        image: decube/decube:latest
        ports:
        - containerPort: 8080
        - containerPort: 9090
        volumeMounts:
        - name: config
          mountPath: /etc/decube
        - name: data
          mountPath: /var/lib/decube
        resources:
          requests:
            cpu: "500m"
            memory: "512Mi"
          limits:
            cpu: "2"
            memory: "2Gi"
      volumes:
      - name: config
        configMap:
          name: decube-config
      - name: data
        persistentVolumeClaim:
          claimName: decube-data
```

### Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: decube
  namespace: decube
spec:
  selector:
    app: decube
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  type: ClusterIP
```

### PersistentVolumeClaim

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: decube-data
  namespace: decube
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
  storageClassName: fast-ssd
```

## Configuration

### Environment Variables

```yaml
env:
- name: DECUBE_CLUSTER_ID
  value: "cluster-001"
- name: DECUBE_STORAGE_CAS_ENDPOINT
  value: "http://minio:9000"
```

### Secrets

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: decube-secrets
  namespace: decube
type: Opaque
stringData:
  cas-access-key: "minioadmin"
  cas-secret-key: "minioadmin"
```

## Scaling

### Horizontal Scaling

```bash
# Scale deployment
kubectl scale deployment decube --replicas=5 -n decube

# Or update Helm values
helm upgrade decube decube/decube \
  --set cluster.replicas=5
```

### Vertical Scaling

```yaml
resources:
  requests:
    cpu: "1"
    memory: "1Gi"
  limits:
    cpu: "4"
    memory: "4Gi"
```

### Auto-scaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: decube-hpa
  namespace: decube
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: decube
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Monitoring

### ServiceMonitor (Prometheus)

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: decube
  namespace: decube
spec:
  selector:
    matchLabels:
      app: decube
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

### PodDisruptionBudget

```yaml
apiVersion: policy/v1
kind: PodDisruptionBudget
metadata:
  name: decube-pdb
  namespace: decube
spec:
  minAvailable: 2
  selector:
    matchLabels:
      app: decube
```

## Networking

### Ingress

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: decube-ingress
  namespace: decube
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - decube.example.com
    secretName: decube-tls
  rules:
  - host: decube.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: decube
            port:
              number: 8080
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n decube
kubectl describe pod <pod-name> -n decube
kubectl logs <pod-name> -n decube
```

### Check Services

```bash
kubectl get svc -n decube
kubectl describe svc decube -n decube
```

### Debug Container

```bash
kubectl exec -it <pod-name> -n decube -- /bin/sh
```

### Common Issues

#### Pods Not Starting
- Check resource limits
- Verify storage class
- Review configuration

#### Connection Issues
- Check service endpoints
- Verify network policies
- Review ingress configuration

#### Storage Issues
- Verify PVC status
- Check storage class
- Review permissions

## Best Practices

1. **Use StatefulSets for Stateful Components**
2. **Configure Resource Limits**
3. **Use PersistentVolumes for Data**
4. **Implement Health Checks**
5. **Set Up Monitoring**
6. **Use Secrets for Sensitive Data**
7. **Implement Network Policies**
8. **Regular Backups**

## References

- [Deployment Guide](deployment.md)
- [Monitoring Guide](monitoring.md)
- [Troubleshooting Guide](troubleshooting.md)

---

*Last updated: January 2024*

