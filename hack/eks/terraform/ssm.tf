resource "aws_ssm_parameter" "bastion_public_ip" {
  name  = "/${var.namespace}/bastion/public_ip"
  type  = "SecureString"
  value = module.ec2_bastion.public_ip
}

resource "aws_ssm_parameter" "bastion_ssh_private_key" {
  name  = "/${var.namespace}/bastion/ssh_private_key"
  type  = "SecureString"
  value = base64encode(module.bastion_key_pair.private_key)
}

resource "aws_ssm_parameter" "eks_kubeconfig" {
  name  = "/${var.namespace}/eks/kubeconfig"
  type  = "SecureString"
  value = base64encode(module.eks.kubeconfig)
}
