# Terraform Configuration for DeCube

This directory contains Terraform configurations for deploying DeCube on AWS.

## Prerequisites

- Terraform >= 1.0
- AWS CLI configured
- Appropriate AWS permissions

## Quick Start

```bash
# Initialize Terraform
terraform init

# Plan deployment
terraform plan

# Apply configuration
terraform apply

# Destroy resources
terraform destroy
```

## Configuration

Edit `variables.tf` to customize:

- AWS region
- Cluster name
- Node group size
- Instance types
- VPC configuration

## Outputs

After deployment, outputs include:

- Cluster endpoint
- Cluster name
- S3 bucket name
- VPC and subnet IDs

## Resources Created

- VPC with public and private subnets
- EKS cluster
- EKS node group
- S3 bucket for storage
- IAM roles and policies
- Internet gateway and route tables

## See Also

- [Kubernetes Deployment Guide](../docs/kubernetes.md)
- [Deployment Guide](../docs/deployment.md)

