# Google Cloud Platform Deployment Guide

Deploying DeCube on Google Cloud Platform (GCP).

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [GKE Deployment](#gke-deployment)
3. [Compute Engine Deployment](#compute-engine-deployment)
4. [Storage Configuration](#storage-configuration)
5. [Networking](#networking)

## Prerequisites

- GCP project
- gcloud CLI configured
- kubectl installed (for GKE)

## GKE Deployment

### Create Cluster

```bash
gcloud container clusters create decube-cluster \
  --num-nodes=3 \
  --machine-type=n1-standard-4 \
  --zone=us-central1-a
```

### Deploy DeCube

```bash
gcloud container clusters get-credentials decube-cluster
kubectl apply -f k8s/
```

## Compute Engine Deployment

### Create Instances

```bash
gcloud compute instances create decube-node-1 \
  --machine-type=n1-standard-4 \
  --image-family=ubuntu-2004-lts \
  --image-project=ubuntu-os-cloud
```

## Storage Configuration

### Cloud Storage

```yaml
storage:
  object_storage:
    endpoint: "https://storage.googleapis.com"
    access_key: "${GCS_ACCESS_KEY}"
    secret_key: "${GCS_SECRET_KEY}"
    bucket: "decube-storage"
```

### Persistent Disks

```yaml
volumes:
  - name: decube-data
    gcePersistentDisk:
      pdName: decube-disk
      fsType: ext4
```

## Networking

### VPC Configuration

- **Custom VPC**: For production
- **Firewall Rules**: Restrict access
- **Cloud Load Balancing**: For high availability

## References

- [GCP Documentation](https://cloud.google.com/docs/)

---

*Last updated: January 2024*

