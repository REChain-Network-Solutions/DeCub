# Microsoft Azure Deployment Guide

Deploying DeCube on Microsoft Azure.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [AKS Deployment](#aks-deployment)
3. [Virtual Machines](#virtual-machines)
4. [Storage Configuration](#storage-configuration)

## Prerequisites

- Azure subscription
- Azure CLI installed
- kubectl installed (for AKS)

## AKS Deployment

### Create Cluster

```bash
az aks create \
  --resource-group decube-rg \
  --name decube-cluster \
  --node-count 3 \
  --node-vm-size Standard_D4s_v3
```

### Deploy DeCube

```bash
az aks get-credentials --resource-group decube-rg --name decube-cluster
kubectl apply -f k8s/
```

## Virtual Machines

### Create VM

```bash
az vm create \
  --resource-group decube-rg \
  --name decube-vm \
  --image UbuntuLTS \
  --size Standard_D4s_v3
```

## Storage Configuration

### Blob Storage

```yaml
storage:
  object_storage:
    endpoint: "https://${STORAGE_ACCOUNT}.blob.core.windows.net"
    access_key: "${AZURE_STORAGE_KEY}"
    bucket: "decube-storage"
```

## References

- [Azure Documentation](https://docs.microsoft.com/azure/)

---

*Last updated: January 2024*

