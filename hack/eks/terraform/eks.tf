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

  cluster_enabled_log_types            = var.eks_cluster_enabled_log_types
  cluster_endpoint_private_access      = var.eks_cluster_endpoint_private_access
  cluster_endpoint_public_access       = var.eks_cluster_endpoint_public_access
  cluster_endpoint_public_access_cidrs = local.eks_public_access_cidrs

  manage_aws_auth = true
  map_roles = [{
    rolearn  = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/${module.ec2_bastion.role}"
    username = "bastion"
    groups   = ["system:masters"]
  }]
  enable_irsa = true

  write_kubeconfig = false

  kubeconfig_aws_authenticator_command      = "aws"
  kubeconfig_aws_authenticator_command_args = ["eks", "get-token", "--cluster-name", local.eks_cluster_name]

  workers_additional_policies = [
    aws_iam_policy.eks_worker_write_logs.id
  ]

  node_groups = [{
    max_capacity     = var.worker_group_max_size
    desired_capacity = var.worker_group_max_size
    instance_types   = [var.worker_group_instance_type]
    subnets          = local.worker_subnets
    disk_size        = 50
  }]

  tags = local.tags
}

resource "aws_security_group_rule" "bastion_eks_cluster_endpoint" {
  security_group_id = module.eks.cluster_security_group_id

  type                     = "ingress"
  from_port                = 443
  to_port                  = 443
  protocol                 = "tcp"
  source_security_group_id = aws_security_group.bastion.id
}

module "cert_manager_irsa" {
  source                        = "terraform-aws-modules/iam/aws//modules/iam-assumable-role-with-oidc"
  version                       = "3.6.0"
  create_role                   = true
  role_name                     = "${var.namespace}-cert_manager-irsa"
  provider_url                  = replace(module.eks.cluster_oidc_issuer_url, "https://", "")
  role_policy_arns              = [aws_iam_policy.cert_manager_policy.arn]
  oidc_fully_qualified_subjects = ["system:serviceaccount:capact-system:cert-manager"]
  tags                          = local.tags
}

resource "aws_iam_policy" "cert_manager_policy" {
  name        = "${var.namespace}-cert-manager-policy"
  path        = "/"
  description = "Policy, which allows CertManager to create Route53 records"

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : "route53:GetChange",
        "Resource" : "arn:aws:route53:::change/*"
      },
      {
        "Effect" : "Allow",
        "Action" : [
          "route53:ChangeResourceRecordSets",
          "route53:ListResourceRecordSets"
        ],
        "Resource" : "arn:aws:route53:::hostedzone/${local.route53_zone_id}"
      },
    ]
  })
}
