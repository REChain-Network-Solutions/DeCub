# Terraform Deployment Guide

This guide covers deploying DeCube infrastructure using Terraform.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Deployment](#deployment)
5. [Customization](#customization)

## Prerequisites

- Terraform >= 1.0
- AWS CLI configured
- Appropriate AWS permissions
- kubectl installed

## Quick Start

```bash
# Navigate to terraform directory
cd terraform

# Initialize Terraform
terraform init

# Review plan
terraform plan

# Apply configuration
terraform apply

# Get outputs
terraform output
```

## Configuration

### Variables

Edit `variables.tf` to customize:

- AWS region
- Cluster name
- Node group size
- Instance types
- VPC configuration

### Example

```hcl
variable "cluster_name" {
  description = "EKS cluster name"
  type        = string
  default     = "decube-production"
}

variable "node_desired_size" {
  description = "Desired number of nodes"
  type        = number
  default     = 5
}
```

## Deployment

### Step 1: Initialize

```bash
terraform init
```

### Step 2: Plan

```bash
terraform plan -out=tfplan
```

### Step 3: Apply

```bash
terraform apply tfplan
```

### Step 4: Configure kubectl

```bash
aws eks update-kubeconfig --name $(terraform output -raw cluster_name)
```

### Step 5: Deploy DeCube

```bash
kubectl apply -f ../k8s/
```

## Customization

### Multiple Environments

```bash
# Development
terraform workspace new dev
terraform apply -var-file=dev.tfvars

# Production
terraform workspace new prod
terraform apply -var-file=prod.tfvars
```

### Custom VPC

```hcl
data "aws_vpc" "existing" {
  id = "vpc-xxxxx"
}
```

## Outputs

After deployment:

```bash
# Get cluster endpoint
terraform output cluster_endpoint

# Get S3 bucket name
terraform output s3_bucket_name
```

## Cleanup

```bash
# Destroy resources
terraform destroy
```

## Best Practices

1. **Use Workspaces**: Separate environments
2. **Version Control**: Track Terraform files
3. **State Management**: Use remote state
4. **Modular Design**: Reusable modules
5. **Documentation**: Document variables

## References

- [Terraform Documentation](https://www.terraform.io/docs/)
- [AWS Provider](https://registry.terraform.io/providers/hashicorp/aws/)
- [Kubernetes Deployment Guide](kubernetes.md)

---

*Last updated: January 2024*

