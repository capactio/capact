provider "aws" {
  region = var.region
}

resource "random_string" "name" {
  length = 8
  special = false
  lower = true
  number = false
  upper = false
}


locals {
  name = var.res_name != "" ?  var.res_name : random_string.name.id
  tags = {
    CapactManaged = true
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

################################################################################
# Supporting Resources
################################################################################

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "~> 3.0.0"

  name = local.name
  cidr = "10.99.0.0/18"

  enable_dns_hostnames = true
  enable_dns_support = true

  enable_nat_gateway = false

  azs = data.aws_availability_zones.available.names
  public_subnets = [
    "10.99.0.0/24",
    "10.99.1.0/24",
    "10.99.2.0/24"]
  private_subnets = []

  create_database_subnet_group = false

  tags = local.tags
}

module "security_group" {
  source = "terraform-aws-modules/security-group/aws"
  version = "~> 4.0.0"

  name = local.name
  description = "PostgreSQL security group created by Capact"
  vpc_id = module.vpc.vpc_id

  # ingress
  ingress_with_cidr_blocks = [
    {
      from_port = 5432
      to_port = 5432
      protocol = "tcp"
      description = "RDS access rule"
      cidr_blocks = var.ingress_rule_cidr_blocks
    },
  ]

  tags = local.tags
}

################################################################################
# RDS Module
################################################################################

module "db" {
  source = "terraform-aws-modules/rds/aws"
  version = "~> 2.35.0"

  identifier = local.name

  # All available versions: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts
  engine = var.engine
  engine_version = var.engine_version
  family = "${var.engine}${var.major_engine_version}"
  # DB parameter group
  major_engine_version = var.major_engine_version
  # DB option group
  instance_class = var.tier

  publicly_accessible = var.publicly_accessible

  allocated_storage = var.allocated_storage
  max_allocated_storage = var.max_allocated_storage
  storage_encrypted = var.storage_encrypted

  # NOTE: Do NOT use 'user' as the value for 'username' as it throws:
  # "Error creating DB Instance: InvalidParameterValue: MasterUsername
  # user cannot be used as it is a reserved word used by the engine"
  name = local.name
  username = var.user_name
  password = var.user_password
  port = 5432

  multi_az = var.multi_az
  subnet_ids = module.vpc.public_subnets
  vpc_security_group_ids = [
    module.security_group.security_group_id]

  maintenance_window = var.maintenance_window
  backup_window = var.backup_window
  enabled_cloudwatch_logs_exports = [
    "postgresql",
    "upgrade"]

  backup_retention_period = var.backup_retention_period
  skip_final_snapshot = var.skip_final_snapshot
  deletion_protection = var.deletion_protection

  performance_insights_enabled = var.performance_insights_enabled
  performance_insights_retention_period = var.performance_insights_retention_period

  create_monitoring_role = true
  monitoring_role_name = local.name
  monitoring_interval = var.monitoring_interval

  parameters = [
    {
      name = "autovacuum"
      value = 1
    },
    {
      name = "client_encoding"
      value = "utf8"
    }
  ]

  tags = local.tags
}
