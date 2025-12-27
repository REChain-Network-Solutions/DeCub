terraform {
  required_version = ">= 1.0"
  
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.20"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.cluster.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.cluster.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.cluster.token
}

# VPC
resource "aws_vpc" "decube" {
  cidr_block           = var.vpc_cidr
  enable_dns_hostnames = true
  enable_dns_support   = true

  tags = {
    Name = "decube-vpc"
  }
}

# Internet Gateway
resource "aws_internet_gateway" "decube" {
  vpc_id = aws_vpc.decube.id

  tags = {
    Name = "decube-igw"
  }
}

# Subnets
resource "aws_subnet" "decube_public" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.decube.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index)
  availability_zone = var.availability_zones[count.index]

  map_public_ip_on_launch = true

  tags = {
    Name = "decube-public-${count.index + 1}"
  }
}

resource "aws_subnet" "decube_private" {
  count             = length(var.availability_zones)
  vpc_id            = aws_vpc.decube.id
  cidr_block        = cidrsubnet(var.vpc_cidr, 8, count.index + length(var.availability_zones))
  availability_zone = var.availability_zones[count.index]

  tags = {
    Name = "decube-private-${count.index + 1}"
  }
}

# Route Tables
resource "aws_route_table" "decube_public" {
  vpc_id = aws_vpc.decube.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.decube.id
  }

  tags = {
    Name = "decube-public-rt"
  }
}

resource "aws_route_table_association" "decube_public" {
  count          = length(aws_subnet.decube_public)
  subnet_id      = aws_subnet.decube_public[count.index].id
  route_table_id = aws_route_table.decube_public.id
}

# EKS Cluster
resource "aws_eks_cluster" "decube" {
  name     = var.cluster_name
  role_arn = aws_iam_role.eks_cluster.arn
  version  = var.kubernetes_version

  vpc_config {
    subnet_ids = concat(aws_subnet.decube_public[*].id, aws_subnet.decube_private[*].id)
  }

  depends_on = [
    aws_iam_role_policy_attachment.eks_cluster_policy,
  ]
}

# EKS Node Group
resource "aws_eks_node_group" "decube" {
  cluster_name    = aws_eks_cluster.decube.name
  node_group_name = "decube-nodes"
  node_role_arn   = aws_iam_role.eks_node.arn
  subnet_ids      = aws_subnet.decube_private[*].id

  scaling_config {
    desired_size = var.node_desired_size
    max_size     = var.node_max_size
    min_size     = var.node_min_size
  }

  instance_types = [var.instance_type]

  depends_on = [
    aws_iam_role_policy_attachment.eks_worker_node_policy,
    aws_iam_role_policy_attachment.eks_cni_policy,
    aws_iam_role_policy_attachment.eks_container_registry_policy,
  ]
}

# S3 Bucket for Storage
resource "aws_s3_bucket" "decube_storage" {
  bucket = var.s3_bucket_name

  tags = {
    Name = "decube-storage"
  }
}

resource "aws_s3_bucket_versioning" "decube_storage" {
  bucket = aws_s3_bucket.decube_storage.id
  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_server_side_encryption_configuration" "decube_storage" {
  bucket = aws_s3_bucket.decube_storage.id

  rule {
    apply_server_side_encryption_by_default {
      sse_algorithm = "AES256"
    }
  }
}

# Data sources
data "aws_eks_cluster" "cluster" {
  name = aws_eks_cluster.decube.name
}

data "aws_eks_cluster_auth" "cluster" {
  name = aws_eks_cluster.decube.name
}

# IAM Roles
resource "aws_iam_role" "eks_cluster" {
  name = "decube-eks-cluster-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "eks.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role" "eks_node" {
  name = "decube-eks-node-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "ec2.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "eks_cluster_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSClusterPolicy"
  role       = aws_iam_role.eks_cluster.name
}

resource "aws_iam_role_policy_attachment" "eks_worker_node_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy"
  role       = aws_iam_role.eks_node.name
}

resource "aws_iam_role_policy_attachment" "eks_cni_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy"
  role       = aws_iam_role.eks_node.name
}

resource "aws_iam_role_policy_attachment" "eks_container_registry_policy" {
  policy_arn = "arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly"
  role       = aws_iam_role.eks_node.name
}

# Outputs
output "cluster_endpoint" {
  value = aws_eks_cluster.decube.endpoint
}

output "cluster_name" {
  value = aws_eks_cluster.decube.name
}

output "s3_bucket_name" {
  value = aws_s3_bucket.decube_storage.id
}

