
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "${var.namespace}-vpc"
  cidr = var.vpc_cidr

  azs             = data.aws_availability_zones.all.names
  private_subnets = var.vpc_private_subnets
  public_subnets  = var.vpc_public_subnets

  enable_dns_hostnames = true

  enable_nat_gateway = true
  single_nat_gateway = var.vpc_single_nat_gateway

  private_subnet_tags = {
    "kubernetes.io/cluster/${local.eks_cluster_name}" : "shared"
  }
  public_subnet_tags = {
    "kubernetes.io/cluster/${local.eks_cluster_name}" : "shared"
  }
}
