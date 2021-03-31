
output "bastion_public_ip" {
  value = module.ec2_bastion.public_ip
}

output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
}

output "route53_zone_name_servers" {
  value = module.zones.this_route53_zone_name_servers
}

output "route53_zone_id" {
  value = local.route53_zone_id
}

output "cert_manager_irsa_role_arn" {
  value = module.cert_manager_irsa.this_iam_role_arn
}

output "bastion_ssh_private_key" {
  value = module.bastion_key_pair.private_key
  sensitive = true
}

output "eks_kubeconfig" {
  value = module.eks.kubeconfig
  sensitive = true
}