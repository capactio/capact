resource "aws_iam_policy" "efs_csi_driver_policy" {
  count = var.efs_enabled ? 1 : 0
  name = "${var.namespace}-efs-csi-driver-policy"
  tags = local.tags

  path = "/"
  description = "Policy, which allows EFS CSI Driver to interact with the file system"

  policy = jsonencode({
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Action": [
          "elasticfilesystem:DescribeAccessPoints",
          "elasticfilesystem:DescribeFileSystems"
        ],
        "Resource": "*"
      },
      {
        "Effect": "Allow",
        "Action": [
          "elasticfilesystem:CreateAccessPoint"
        ],
        "Resource": "*",
        "Condition": {
          "StringLike": {
            "aws:RequestTag/efs.csi.aws.com/cluster": "true"
          }
        }
      },
      {
        "Effect": "Allow",
        "Action": "elasticfilesystem:DeleteAccessPoint",
        "Resource": "*",
        "Condition": {
          "StringEquals": {
            "aws:ResourceTag/efs.csi.aws.com/cluster": "true"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role" "efs_csi_driver_role" {
  count = var.efs_enabled ? 1 : 0
  name = "${var.namespace}-efs-csi-driver-role"
  tags = local.tags

  assume_role_policy = jsonencode({
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "Federated": module.eks.oidc_provider_arn
        },
        "Action": "sts:AssumeRoleWithWebIdentity",
        "Condition": {
          "StringEquals": {
            "${replace(module.eks.cluster_oidc_issuer_url, "https://", "")}:sub": "system:serviceaccount:kube-system:efs-csi-controller-sa"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "efs_csi_driver_policy_attachment" {
  count = var.efs_enabled ? 1 : 0

  role = aws_iam_role.efs_csi_driver_role[count.index].name
  policy_arn = aws_iam_policy.efs_csi_driver_policy[count.index].arn
}

resource "kubernetes_service_account" "efs_csi_driver_ctrl_sa" {
  count = var.efs_enabled ? 1 : 0
  metadata {
    name = "efs-csi-controller-sa"
    namespace = "kube-system"
    annotations = {
      "eks.amazonaws.com/role-arn": aws_iam_role.efs_csi_driver_role[count.index].arn
    }
  }
}

module "efs_security_group" {
  source = "terraform-aws-modules/security-group/aws"
  version = "~> 3"

  name = "${var.namespace}-efs-security-group"
  description = "Security group for EFS"
  vpc_id = module.vpc.vpc_id

  ingress_with_cidr_blocks = [
    {
      from_port = 2049
      to_port = 2049
      protocol = "tcp"
      description = "EFS access rule"
      cidr_blocks = module.vpc.vpc_cidr_block
    },
  ]

  tags = local.tags
}

resource "aws_efs_file_system" "eks_efs" {
  count = var.efs_enabled ? 1 : 0
  performance_mode = "generalPurpose"
  tags = local.tags
}

resource "aws_efs_mount_target" "eks_efs_mount" {
  for_each = var.efs_enabled ? toset(local.worker_subnets) : []

  subnet_id = module.vpc.private_subnets[index(local.worker_subnets, each.value)]
  file_system_id = aws_efs_file_system.eks_efs[0].id
  security_groups = [module.efs_security_group.this_security_group_id]
}

resource "kubernetes_storage_class" "efs_storage_class" {
  count = var.efs_enabled ? 1 : 0
  metadata {
    name = "efs-sc"
  }
  storage_provisioner = "efs.csi.aws.com"
  parameters = {
    provisioningMode: "efs-ap"
    fileSystemId: aws_efs_file_system.eks_efs[count.index].id
    directoryPerms: "700"
  }
}
