terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.0"
    }
  }
}

provider "aws" {
  region = var.region
}

module "main" {
  source = "git::https://github.com/lgallard/terraform-aws-elasticsearch.git?ref=0.9.0"

  domain_name                                              = var.domain_name
  elasticsearch_version                                    = var.elasticsearch_version
  access_policies                                          = local.access_policies
  advanced_security_options                                = var.advanced_security_options
  advanced_security_options_enabled                        = var.advanced_security_options_enabled
  advanced_security_options_internal_user_database_enabled = var.advanced_security_options_internal_user_database_enabled
  advanced_security_options_master_user_arn                = var.advanced_security_options_master_user_arn
  advanced_security_options_master_user_username           = var.advanced_security_options_master_user_username
  advanced_security_options_master_user_password           = var.advanced_security_options_master_user_password
  domain_endpoint_options                                  = var.domain_endpoint_options
  domain_endpoint_options_enforce_https                    = var.domain_endpoint_options_enforce_https
  domain_endpoint_options_tls_security_policy              = var.domain_endpoint_options_tls_security_policy
  domain_endpoint_options_custom_endpoint_enabled          = var.domain_endpoint_options_custom_endpoint_enabled
  domain_endpoint_options_custom_endpoint                  = var.domain_endpoint_options_custom_endpoint
  domain_endpoint_options_custom_endpoint_certificate_arn  = var.domain_endpoint_options_custom_endpoint_certificate_arn
  advanced_options                                         = var.advanced_options
  ebs_options                                              = var.ebs_options
  ebs_enabled                                              = var.ebs_enabled
  ebs_options_volume_type                                  = var.ebs_options_volume_type
  ebs_options_volume_size                                  = var.ebs_options_volume_size
  ebs_options_iops                                         = var.ebs_options_iops
  encrypt_at_rest                                          = var.encrypt_at_rest
  encrypt_at_rest_enabled                                  = var.encrypt_at_rest_enabled
  encrypt_at_rest_kms_key_id                               = var.encrypt_at_rest_kms_key_id
  node_to_node_encryption                                  = var.node_to_node_encryption
  node_to_node_encryption_enabled                          = var.node_to_node_encryption_enabled
  cluster_config                                           = var.cluster_config
  cluster_config_instance_type                             = var.cluster_config_instance_type
  cluster_config_instance_count                            = var.cluster_config_instance_count
  cluster_config_dedicated_master_enabled                  = var.cluster_config_dedicated_master_enabled
  cluster_config_dedicated_master_type                     = var.cluster_config_dedicated_master_type
  cluster_config_dedicated_master_count                    = var.cluster_config_dedicated_master_count
  cluster_config_availability_zone_count                   = var.cluster_config_availability_zone_count
  cluster_config_zone_awareness_enabled                    = var.cluster_config_zone_awareness_enabled
  cluster_config_warm_enabled                              = var.cluster_config_warm_enabled
  cluster_config_warm_count                                = var.cluster_config_warm_count
  cluster_config_warm_type                                 = var.cluster_config_warm_type
  snapshot_options                                         = var.snapshot_options
  snapshot_options_automated_snapshot_start_hour           = var.snapshot_options_automated_snapshot_start_hour
  vpc_options                                              = var.vpc_options
  vpc_options_security_group_ids                           = var.vpc_options_security_group_ids
  vpc_options_subnet_ids                                   = var.vpc_options_subnet_ids
  log_publishing_options                                   = var.log_publishing_options
  log_publishing_options_log_type                          = var.log_publishing_options_log_type
  log_publishing_options_cloudwatch_log_group_arn          = var.log_publishing_options_cloudwatch_log_group_arn
  log_publishing_options_enabled                           = var.log_publishing_options_enabled
  log_publishing_options_retention                         = var.log_publishing_options_retention
  cognito_options                                          = var.cognito_options
  cognito_options_enabled                                  = var.cognito_options_enabled
  cognito_options_user_pool_id                             = var.cognito_options_user_pool_id
  cognito_options_identity_pool_id                         = var.cognito_options_identity_pool_id
  cognito_options_role_arn                                 = var.cognito_options_role_arn
  tags                                                     = var.tags
  timeouts                                                 = var.timeouts
  timeouts_update                                          = var.timeouts_update
  create_service_link_role                                 = var.create_service_link_role
}
