# DeCube Deployment Guide

This guide covers production deployment strategies for DeCube across different environments and scales.

## Deployment Architectures

### Single-Node Development
For development and testing with minimal resources.

```
┌─────────────────┐
│   Single Node   │
│                 │
│ ┌─────────────┐ │
│ │  All Services│ │
│ │   (Docker)   │ │
│ └─────────────┘ │
└─────────────────┘
```

**Requirements:**
- 4 CPU cores
- 8GB RAM
- 50GB storage

**Use Case:** Local development, CI/CD testing

### Multi-Node Cluster
Production-ready deployment with high availability.

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Node 1    │    │   Node 2    │    │   Node 3    │
│             │    │             │    │             │
│ Control     │    │  Catalog    │    │   Gossip    │
│   Plane     │    │             │    │             │
│             │    │             │    │             │
│   GCL       │    │   Storage   │    │    CAS      │
└─────────────┘    └─────────────┘    └─────────────┘
```

**Requirements per node:**
- 8 CPU cores
- 16GB RAM
- 500GB SSD storage
- 10Gbps network

**Use Case:** Production clusters, multi-region deployments

### Cloud-Native Deployment
Kubernetes-based deployment with auto-scaling.

```
┌─────────────────────────────────────────────────┐
│              Kubernetes Cluster                 │
├─────────────────────────────────────────────────┤
│                                                 │
│  ┌─────────────┐ ┌─────────────┐ ┌─────────────┐ │
│  │  Control    │ │   Catalog   │ │   Gossip    │ │
│  │   Plane     │ │             │ │             │ │
│  │             │ │             │ │             │ │
│  │     GCL     │ │   Storage   │ │     CAS     │ │
│  └─────────────┘ └─────────────┘ └─────────────┘ │
│                                                 │
│  ┌─────────────────────────────────────────────┐ │
│  │            etcd Cluster (3 nodes)           │ │
│  └─────────────────────────────────────────────┘ │
│                                                 │
│  ┌─────────────────────────────────────────────┐ │
│  │         Object Storage (S3/MinIO)           │ │
│  └─────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────┘
```

**Requirements:**
- Kubernetes 1.24+
- 3+ worker nodes
- External load balancer
- Persistent storage classes

## Prerequisites

### Infrastructure Requirements

#### Hardware Specifications
| Component | CPU | RAM | Storage | Network |
|-----------|-----|-----|---------|---------|
| Control Plane | 4 cores | 8GB | 100GB SSD | 1Gbps |
| Catalog Node | 4 cores | 8GB | 200GB SSD | 1Gbps |
| Gossip Node | 4 cores | 8GB | 100GB SSD | 10Gbps |
| GCL Node | 8 cores | 16GB | 500GB SSD | 10Gbps |
| Storage Node | 4 cores | 8GB | 1TB SSD | 10Gbps |
| CAS Node | 4 cores | 8GB | 1TB SSD | 10Gbps |

#### Network Requirements
- **Internal Network**: 10Gbps for cluster communication
- **External Network**: 1Gbps minimum for client access
- **Latency**: <1ms between cluster nodes
- **Firewall**: Open ports 2379, 2380 (etcd), 8080-8085 (services), 9000 (storage)

### Software Dependencies

#### Required Software
- **Operating System**: Ubuntu 20.04 LTS or RHEL 8+
- **Container Runtime**: Docker 20.10+ or containerd 1.6+
- **Orchestration**: Kubernetes 1.24+ or Docker Compose 2.0+
- **Load Balancer**: NGINX, HAProxy, or cloud load balancer
- **Monitoring**: Prometheus + Grafana (recommended)

#### Optional Software
- **Service Mesh**: Istio 1.16+ or Linkerd 2.11+
- **Logging**: ELK Stack or Loki
- **Backup**: Velero or custom backup solutions
- **Security**: Vault for secrets management

## Deployment Methods

### Docker Compose (Development)

1. **Clone Repository**
```bash
git clone https://github.com/REChain-Network-Solutions/DeCub.git
cd DeCub
```

2. **Configure Environment**
```bash
cp config/docker-compose.yml.example docker-compose.yml
# Edit docker-compose.yml with your settings
```

3. **Deploy Services**
```bash
docker-compose up -d
```

4. **Verify Deployment**
```bash
docker-compose ps
curl http://localhost:8080/api/v1/status
```

### Kubernetes Deployment

1. **Prepare Cluster**
```bash
# Create namespace
kubectl create namespace decube

# Add helm repo (if using helm)
helm repo add decube https://charts.decube.io
helm repo update
```

2. **Deploy etcd**
```bash
kubectl apply -f k8s-manifests/etcd/
```

3. **Deploy DeCube Services**
```bash
# Using kubectl
kubectl apply -f k8s-manifests/

# Or using helm
helm install decube decube/decube \
  --namespace decube \
  --values values.yaml
```

4. **Configure Ingress**
```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: decube-ingress
  namespace: decube
spec:
  rules:
  - host: api.decube.example.com
    http:
      paths:
      - path: /catalog
        pathType: Prefix
        backend:
          service:
            name: catalog-service
            port:
              number: 8080
      - path: /gcl
        pathType: Prefix
        backend:
          service:
            name: gcl-service
            port:
              number: 8081
```

### Cloud Provider Deployments

#### AWS EKS
```bash
# Create EKS cluster
eksctl create cluster --name decube-cluster --region us-east-1

# Deploy DeCube
helm install decube decube/decube \
  --set storage.type=s3 \
  --set storage.s3.bucket=decube-storage
```

#### Google Cloud GKE
```bash
# Create GKE cluster
gcloud container clusters create decube-cluster \
  --num-nodes=3 \
  --machine-type=n1-standard-4

# Deploy DeCube
helm install decube decube/decube \
  --set storage.type=gcs \
  --set storage.gcs.bucket=decube-storage
```

#### Azure AKS
```bash
# Create AKS cluster
az aks create --resource-group decube-rg \
  --name decube-cluster \
  --node-count 3 \
  --node-vm-size Standard_D4s_v3

# Deploy DeCube
helm install decube decube/decube \
  --set storage.type=azure \
  --set storage.azure.container=decube-storage
```

## Configuration Management

### Configuration Sources

#### ConfigMaps for Static Configuration
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: decube-config
  namespace: decube
data:
  cluster-id: "prod-cluster-001"
  gossip-interval: "10s"
  consensus-timeout: "30s"
```

#### Secrets for Sensitive Data
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: decube-secrets
  namespace: decube
type: Opaque
data:
  tls-cert: <base64-encoded-cert>
  tls-key: <base64-encoded-key>
  storage-key: <base64-encoded-key>
```

#### Environment Variables
```yaml
env:
- name: DECUB_NODE_ID
  valueFrom:
    fieldRef:
      fieldPath: metadata.name
- name: DECUB_CLUSTER_ID
  value: "prod-cluster-001"
- name: DECUB_STORAGE_ENDPOINT
  value: "https://storage.example.com"
```

### Configuration Validation

DeCube includes configuration validation at startup:

```bash
# Validate configuration
decubectl config validate --config config.yaml

# Check for common issues
decubectl config check --cluster
```

## Storage Configuration

### Object Storage Setup

#### MinIO (Self-Hosted)
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: minio
  namespace: decube
spec:
  replicas: 4
  template:
    spec:
      containers:
      - name: minio
        image: minio/minio:latest
        args:
        - server
        - /data
        env:
        - name: MINIO_ACCESS_KEY
          value: "decube"
        - name: MINIO_SECRET_KEY
          valueFrom:
            secretKeyRef:
              name: minio-secret
              key: secret-key
        volumeMounts:
        - name: data
          mountPath: /data
```

#### Cloud Storage Integration
```yaml
# AWS S3
storage:
  type: s3
  s3:
    endpoint: https://s3.amazonaws.com
    bucket: decube-storage
    region: us-east-1
    accessKey: <access-key>
    secretKey: <secret-key>

# Google Cloud Storage
storage:
  type: gcs
  gcs:
    bucket: decube-storage
    credentials: /secrets/gcs-key.json

# Azure Blob Storage
storage:
  type: azure
  azure:
    account: decubestorage
    container: snapshots
    key: <account-key>
```

### Database Configuration

#### etcd Cluster Setup
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: etcd
  namespace: decube
spec:
  replicas: 3
  serviceName: etcd
  template:
    spec:
      containers:
      - name: etcd
        image: quay.io/coreos/etcd:v3.5.5
        env:
        - name: ETCD_NAME
          value: $(POD_NAME)
        - name: ETCD_DATA_DIR
          value: /etcd-data
        - name: ETCD_LISTEN_PEER_URLS
          value: http://0.0.0.0:2380
        - name: ETCD_LISTEN_CLIENT_URLS
          value: http://0.0.0.0:2379
        - name: ETCD_INITIAL_ADVERTISE_PEER_URLS
          value: http://$(POD_NAME).etcd:2380
        - name: ETCD_ADVERTISE_CLIENT_URLS
          value: http://$(POD_NAME).etcd:2379
        - name: ETCD_INITIAL_CLUSTER
          value: etcd-0=http://etcd-0.etcd:2380,etcd-1=http://etcd-1.etcd:2380,etcd-2=http://etcd-2.etcd:2380
```

## Networking Configuration

### Service Mesh Integration

#### Istio Integration
```yaml
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: decube-gateway
  namespace: decube
spec:
  selector:
    istio: ingressgateway
  servers:
  - port:
      number: 80
      name: http
      protocol: HTTP
    hosts:
    - api.decube.example.com
---
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: decube-api
  namespace: decube
spec:
  hosts:
  - api.decube.example.com
  gateways:
  - decube-gateway
  http:
  - match:
    - uri:
        prefix: /catalog
    route:
    - destination:
        host: catalog-service
  - match:
    - uri:
        prefix: /gcl
    route:
    - destination:
        host: gcl-service
```

### Load Balancing

#### External Load Balancer
```yaml
apiVersion: v1
kind: Service
metadata:
  name: decube-external-lb
  namespace: decube
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: nlb
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  - port: 443
    targetPort: 8443
    protocol: TCP
    name: https
  selector:
    app: decube-api-gateway
```

## Security Configuration

### TLS Configuration

#### Certificate Management
```yaml
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@decube.example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
---
apiVersion: cert-manager.io/v1
kind: Certificate
metadata:
  name: decube-tls
  namespace: decube
spec:
  secretName: decube-tls-secret
  issuerRef:
    name: letsencrypt-prod
    kind: ClusterIssuer
  dnsNames:
  - api.decube.example.com
  - gossip.decube.example.com
```

#### mTLS Configuration
```yaml
apiVersion: security.istio.io/v1beta1
kind: PeerAuthentication
metadata:
  name: default
  namespace: decube
spec:
  mtls:
    mode: STRICT
```

### Network Policies

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: decube-network-policy
  namespace: decube
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: decube
    ports:
    - protocol: TCP
      port: 8080
    - protocol: TCP
      port: 8081
  egress:
  - to:
    - namespaceSelector:
        matchLabels:
          name: decube
  - to: []
    ports:
    - protocol: TCP
      port: 53
    - protocol: UDP
      port: 53
```

## Monitoring and Observability

### Prometheus Metrics

DeCube exposes metrics at `/metrics` endpoint:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: decube-servicemonitor
  namespace: monitoring
spec:
  selector:
    matchLabels:
      app: decube
  endpoints:
  - port: metrics
    interval: 30s
```

### Logging Configuration

#### Centralized Logging with Loki
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: promtail-config
  namespace: decube
data:
  promtail.yaml: |
    server:
      http_listen_port: 9080
      grpc_listen_port: 0
    positions:
      filename: /tmp/positions.yaml
    clients:
    - url: http://loki.monitoring.svc.cluster.local:3100/loki/api/v1/push
    scrape_configs:
    - job_name: decube
      static_configs:
      - targets:
        - localhost
        labels:
          job: decube
          __path__: /var/log/decube/*.log
```

### Health Checks

```yaml
apiVersion: v1
kind: Service
metadata:
  name: decube-health
  namespace: decube
spec:
  ports:
  - port: 8080
    targetPort: 8080
  selector:
    app: decube
---
apiVersion: monitoring.coreos.com/v1
kind: Probe
metadata:
  name: decube-health-probe
  namespace: decube
spec:
  prober:
    url: http://decube-health:8080
    path: /health
  targets:
    staticConfig:
      static:
      - api.decube.example.com
```

## Backup and Recovery

### Automated Backups

```yaml
apiVersion: velero.io/v1
kind: Schedule
metadata:
  name: decube-backup
  namespace: velero
spec:
  schedule: "0 1 * * *"
  template:
    includedNamespaces:
    - decube
    storageLocation: aws-s3-backup
    ttl: 720h0m0s
```

### Disaster Recovery

1. **Prepare Recovery Environment**
```bash
# Create new cluster
kubectl create namespace decube-recovery

# Restore from backup
velero restore create --from-backup decube-backup-20240115
```

2. **Verify Recovery**
```bash
# Check service health
kubectl get pods -n decube-recovery

# Verify data integrity
decubectl status --cluster recovery-cluster
```

## Scaling and Performance

### Horizontal Scaling

#### Pod Autoscaling
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: catalog-hpa
  namespace: decube
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: catalog
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

#### Vertical Scaling
```bash
# Scale deployment
kubectl scale deployment catalog --replicas=5 -n decube

# Update resource requests/limits
kubectl set resources deployment catalog \
  --requests=cpu=2,memory=4Gi \
  --limits=cpu=4,memory=8Gi -n decube
```

### Performance Tuning

#### JVM Tuning (if applicable)
```yaml
env:
- name: JAVA_OPTS
  value: "-Xmx4g -Xms2g -XX:+UseG1GC -XX:MaxGCPauseMillis=200"
```

#### Database Tuning
```yaml
# etcd tuning
env:
- name: ETCD_QUOTA_BACKEND_BYTES
  value: "4294967296"  # 4GB
- name: ETCD_SNAPSHOT_COUNT
  value: "10000"
```

## Troubleshooting

### Common Issues

#### Service Unavailable
```bash
# Check pod status
kubectl get pods -n decube

# Check service endpoints
kubectl get endpoints -n decube

# View logs
kubectl logs -l app=decube -n decube --tail=100
```

#### Performance Issues
```bash
# Check resource usage
kubectl top pods -n decube

# Check network policies
kubectl get networkpolicies -n decube

# Profile application
go tool pprof http://localhost:8080/debug/pprof/profile
```

#### Storage Issues
```bash
# Check PVC status
kubectl get pvc -n decube

# Verify storage class
kubectl get storageclass

# Check storage logs
kubectl logs -l app=minio -n decube
```

### Debug Commands

```bash
# Enter pod for debugging
kubectl exec -it <pod-name> -n decube -- /bin/bash

# Port forward for local access
kubectl port-forward svc/catalog-service 8080:8080 -n decube

# Run diagnostic commands
decubectl diagnose --cluster prod-cluster-001
```

## Maintenance Procedures

### Rolling Updates
```bash
# Update deployment image
kubectl set image deployment/catalog catalog=decube/catalog:v1.1.0 -n decube

# Check rollout status
kubectl rollout status deployment/catalog -n decube

# Rollback if needed
kubectl rollout undo deployment/catalog -n decube
```

### Cluster Maintenance
```bash
# Drain node for maintenance
kubectl drain <node-name> --ignore-daemonsets

# Perform maintenance tasks
# ...

# Uncordon node
kubectl uncordon <node-name>
```

### Certificate Rotation
```bash
# Rotate certificates
kubectl delete secret decube-tls-secret -n decube

# New certificates will be automatically provisioned by cert-manager
```

## Compliance and Auditing

### Audit Logging
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: audit-config
  namespace: decube
data:
  audit-policy.yaml: |
    apiVersion: audit.k8s.io/v1
    kind: Policy
    rules:
    - level: Metadata
      resources:
      - group: ""
        resources: ["pods", "services"]
      namespaces: ["decube"]
```

### Compliance Checks
```bash
# Run compliance scans
decubectl compliance check --standard SOC2

# Generate compliance reports
decubectl compliance report --format pdf --output compliance-2024.pdf
```

This deployment guide provides comprehensive instructions for deploying DeCube in various environments. For specific configuration examples or additional assistance, refer to the documentation or community forums.
