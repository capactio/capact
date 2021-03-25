
module "vpc" {
  source = "terraform-aws-modules/vpc/aws"

  name = "${var.namespace}-vpc"
  cidr = var.vpc_cidr

  azs             = var.azs
  private_subnets = var.vpc_private_subnets
  public_subnets  = var.vpc_public_subnets

  enable_nat_gateway = true
  single_nat_gateway = var.vpc_single_nat_gateway

  private_subnet_tags = {
    "kubernetes.io/cluster/${local.eks_cluster_name}" : "shared"
  }
  public_subnet_tags = {
    "kubernetes.io/cluster/${local.eks_cluster_name}" : "shared"
  }
}

module "aws_key_pair" {
  source  = "cloudposse/key-pair/aws"
  version = "0.16.1"

  name                = "${var.namespace}-bastion-key"
  attributes          = ["ssh", "key"]
  ssh_public_key_path = false
  generate_ssh_key    = true
}

resource "aws_security_group" "bastion" {
  name        = "bastion_sg"
  description = "Allow SSH inbound traffic"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "SSH access"
    from_port   = 22
    to_port     = 22
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

module "ec2_bastion" {
  source  = "cloudposse/ec2-bastion-server/aws"
  version = "0.25.0"

  name          = "${var.namespace}-bastion"
  key_name      = module.aws_key_pair.key_name
  instance_type = "t3a.micro"

  vpc_id                      = module.vpc.vpc_id
  subnets                     = module.vpc.public_subnets
  security_groups             = [aws_security_group.bastion.id]
  associate_public_ip_address = true
}
