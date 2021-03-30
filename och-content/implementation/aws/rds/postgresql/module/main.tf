variable "database_version" {
  type = string
  default = "11"
  description = "PostgreSQL database version"
}

variable "region" {
  type = string
  default = "eu-west-1"
  description = "AWS region"
}

variable "tier" {
  type = string
  default = "db.t3.micro"
  description = "AWS RDS instance tier"
}

variable "user_name" {
  type = string
  description = "Database user name"
  default = "postgres"
}

variable "user_password" {
  type = string
  description = "Database user password"
  default = "voltronpostgres"
  //  can't be sensitive as we need it on output
}


terraform {
  required_version = ">= 0.12.26"

  required_providers {
    aws = {
      source = "hashicorp/aws"
      version = ">= 2.49"
    }
  }
}

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
  name = random_string.name.id
  tags = {}
}

################################################################################
# Supporting Resources
################################################################################

module "vpc" {
  source = "terraform-aws-modules/vpc/aws"
  version = "~> 2"

  name = local.name
  cidr = "10.99.0.0/18"

  enable_dns_hostnames = true
  enable_dns_support = true

  azs = [
    "${var.region}a",
    "${var.region}b",
    "${var.region}c"]
  public_subnets = [
    "10.99.0.0/24",
    "10.99.1.0/24",
    "10.99.2.0/24"]
  private_subnets = [
    "10.99.3.0/24",
    "10.99.4.0/24",
    "10.99.5.0/24"]

  create_database_subnet_group = false

  tags = local.tags
}

module "security_group" {
  source = "terraform-aws-modules/security-group/aws"
  version = "~> 3"

  name = local.name
  description = "Complete PostgreSQL example security group"
  vpc_id = module.vpc.vpc_id

  # ingress
  ingress_with_cidr_blocks = [
    {
      from_port = 5432
      to_port = 5432
      protocol = "tcp"
      description = "PostgreSQL access from anywhere"
      cidr_blocks = "0.0.0.0/0"
    },
  ]

  tags = local.tags
}

################################################################################
# RDS Module
################################################################################

module "db" {
  source = "terraform-aws-modules/rds/aws"
  version = "~> 2.0"

  identifier = local.name

  # All available versions: https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts
  engine = "postgres"
  engine_version = "11.10"
  family = "postgres11"
  # DB parameter group
  major_engine_version = var.database_version
  # DB option group
  instance_class = var.tier

  publicly_accessible = true

  allocated_storage = 20
  max_allocated_storage = 100
  storage_encrypted = false

  # NOTE: Do NOT use 'user' as the value for 'username' as it throws:
  # "Error creating DB Instance: InvalidParameterValue: MasterUsername
  # user cannot be used as it is a reserved word used by the engine"
  name = local.name
  username = var.user_name
  password = var.user_password
  port = 5432

  multi_az = false
  subnet_ids = module.vpc.public_subnets
  vpc_security_group_ids = [
    module.security_group.this_security_group_id]

  maintenance_window = "Mon:00:00-Mon:03:00"
  backup_window = "03:00-06:00"
  enabled_cloudwatch_logs_exports = [
    "postgresql",
    "upgrade"]

  backup_retention_period = 0
  skip_final_snapshot = true
  deletion_protection = false

  performance_insights_enabled = true
  performance_insights_retention_period = 7
  create_monitoring_role = true
  monitoring_interval = 60

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
  db_option_group_tags = {
    "Sensitive" = "low"
  }
  db_parameter_group_tags = {
    "Sensitive" = "low"
  }
  db_subnet_group_tags = {
    "Sensitive" = "high"
  }
}

output "instance_ip_addr" {
  description = "The address of the RDS instance"
  value = module.db.this_db_instance_address
}

output "username" {
  description = "The master username for the database"
  value = module.db.this_db_instance_username
}

output "password" {
  description = "The database password"
  value = var.user_password
  sensitive = false
}
