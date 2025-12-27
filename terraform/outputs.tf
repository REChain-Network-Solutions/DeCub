output "cluster_endpoint" {
  description = "EKS cluster endpoint"
  value       = aws_eks_cluster.decube.endpoint
}

output "cluster_name" {
  description = "EKS cluster name"
  value       = aws_eks_cluster.decube.name
}

output "s3_bucket_name" {
  description = "S3 bucket name"
  value       = aws_s3_bucket.decube_storage.id
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.decube.id
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = aws_subnet.decube_public[*].id
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = aws_subnet.decube_private[*].id
}

