
output "bastion_public_ip" {
  value = module.ec2_bastion.public_ip
  description = "Public IP address of the bastion host"
}

output "eks_cluster_endpoint" {
  value = module.eks.cluster_endpoint
  description = "EKS cluster endpoint address"
}

output "route53_zone_name_servers" {
  value = module.zones.this_route53_zone_name_servers
  description = "Name servers for the Route53 Hosted Zone, which is responsible for the Capact domain."
}

output "route53_zone_id" {
  value = local.route53_zone_id
  description = "ID of the Route53 Hosted Zone, which is responsible for the Capact domain."
}

output "cert_manager_irsa_role_arn" {
  value = module.cert_manager_irsa.this_iam_role_arn
  description = "ARN of the IAM Role for the Cert Manager service account. The role is used to perform DNS01 challenges."
}

output "bastion_ssh_private_key" {
  value     = module.bastion_key_pair.private_key
  sensitive = true
  description = "Private SSH key to access the bastion host. The username is ec2-user."
}

output "eks_kubeconfig" {
  value     = module.eks.kubeconfig
  sensitive = true
  description = "Kubeconfig for the EKS cluster endpoint."
}
