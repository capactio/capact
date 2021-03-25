resource "aws_iam_policy" "eks_worker_write_logs" {
  name        = "${var.namespace}-cloudwatch-eks-worker-write-policy"
  path        = "/"
  description = "Policy, which allows EKS worker nodes to write logs to CloudWatch"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Effect   = "Allow"
        Resource = "*"
      },
    ]
  })
}

module "eks" {
  source          = "terraform-aws-modules/eks/aws"
  cluster_name    = local.eks_cluster_name
  cluster_version = var.eks_cluster_version

  vpc_id  = module.vpc.vpc_id
  subnets = concat(module.vpc.private_subnets, module.vpc.public_subnets)

  cluster_enabled_log_types       = var.eks_cluster_enabled_log_types
  cluster_endpoint_private_access = var.eks_cluster_endpoint_private_access
  cluster_endpoint_public_access  = var.eks_cluster_endpoint_public_access

  manage_aws_auth = true # TODO won't work with private cluster endpoint

  workers_additional_policies = [
    aws_iam_policy.eks_worker_write_logs.id
  ]
  worker_groups = [
    {
      instance_type    = var.worker_group_instance_type
      asg_max_size     = var.worker_group_max_size
      asg_desired_capacity = var.worker_group_max_size
      root_volume_type = "gp2"
      subnets          = module.vpc.private_subnets
    }
  ]
}

resource "aws_security_group_rule" "bastion_eks_cluster_endpoint" {
  security_group_id = module.eks.cluster_security_group_id

  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.bastion.id
}
