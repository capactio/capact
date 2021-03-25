
output "bastion_public_ip" {
  value = module.ec2_bastion.public_ip
}

output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}
