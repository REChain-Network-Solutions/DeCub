# AWS Deployment Guide

Deploying DeCube on Amazon Web Services (AWS).

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [EKS Deployment](#eks-deployment)
4. [EC2 Deployment](#ec2-deployment)
5. [Storage Configuration](#storage-configuration)
6. [Networking](#networking)
7. [Monitoring](#monitoring)

## Prerequisites

- AWS account
- AWS CLI configured
- kubectl installed (for EKS)
- Terraform installed (optional)

## Quick Start

### Using Terraform

```bash
cd terraform
terraform init
terraform plan
terraform apply
```

### Using AWS Console

1. Create EKS cluster
2. Configure node groups
3. Deploy DeCube manifests
4. Configure storage

## EKS Deployment

### Create Cluster

```bash
aws eks create-cluster \
  --name decube-cluster \
  --role-arn arn:aws:iam::ACCOUNT:role/EKSClusterRole \
  --resources-vpc-config subnetIds=subnet-xxx,subnet-yyy
```

### Deploy DeCube

```bash
kubectl apply -f k8s/
```

## EC2 Deployment

### Launch Instances

```bash
aws ec2 run-instances \
  --image-id ami-xxx \
  --instance-type t3.medium \
  --key-name my-key \
  --security-groups sg-xxx \
  --user-data file://user-data.sh
```

### Configure Services

```bash
# Install DeCube
./scripts/setup-dev.sh

# Start services
docker-compose up -d
```

## Storage Configuration

### S3 for Object Storage

```yaml
storage:
  object_storage:
    endpoint: "https://s3.amazonaws.com"
    access_key: "${AWS_ACCESS_KEY}"
    secret_key: "${AWS_SECRET_KEY}"
    bucket: "decube-storage"
    region: "us-east-1"
```

### EBS for Block Storage

```yaml
volumes:
  - name: decube-data
    ebs:
      volumeSize: 100
      volumeType: gp3
```

## Networking

### VPC Configuration

- **Public Subnets**: For load balancers
- **Private Subnets**: For application nodes
- **NAT Gateway**: For outbound internet

### Security Groups

```yaml
security_groups:
  - name: decube-api
    ports: [8080, 9090]
    source: 0.0.0.0/0
  - name: decube-internal
    ports: [7000, 8000]
    source: 10.0.0.0/16
```

## Monitoring

### CloudWatch Integration

```yaml
monitoring:
  cloudwatch:
    enabled: true
    namespace: "DeCube"
    metrics:
      - cpu_usage
      - memory_usage
      - request_rate
```

## Cost Optimization

- Use Reserved Instances
- Enable Auto Scaling
- Use S3 Intelligent-Tiering
- Optimize EBS volumes

## References

- [Terraform Guide](../terraform.md)
- [Kubernetes Guide](../kubernetes.md)
- [AWS Documentation](https://docs.aws.amazon.com/)

---

*Last updated: January 2024*

