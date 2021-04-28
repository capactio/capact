variable "region" {
  type        = string
  description = "AWS region"
}

variable "domain_name" {
  description = "Name of the domain"
  type        = string
}

variable "elasticsearch_version" {
  description = "The version of Elasticsearch to deploy."
  type        = string
  default     = "7.10"
}

variable "access_policies" {
  description = "IAM policy document specifying the access policies for the domain"
  type        = string
  default     = ""
}

# Advanced security options
variable "advanced_security_options" {
  description = "Options for fine-grained access control"
  type        = any
  default     = {}
}

variable "advanced_security_options_enabled" {
  description = "Whether advanced security is enabled (Forces new resource)"
  type        = bool
  default     = false
}

variable "advanced_security_options_internal_user_database_enabled" {
  description = "Whether the internal user database is enabled. If not set, defaults to false by the AWS API."
  type        = bool
  default     = false
}

variable "advanced_security_options_master_user_arn" {
  description = "ARN for the master user. Only specify if `internal_user_database_enabled` is not set or set to `false`"
  type        = string
  default     = null
}

variable "advanced_security_options_master_user_username" {
  description = "The master user's username, which is stored in the Amazon Elasticsearch Service domain's internal database. Only specify if `internal_user_database_enabled` is set to `true`."
  type        = string
  default     = null
}

variable "advanced_security_options_master_user_password" {
  description = "The master user's password, which is stored in the Amazon Elasticsearch Service domain's internal database. Only specify if `internal_user_database_enabled` is set to `true`."
  type        = string
  default     = null
}

# Domain endpoint options
variable "domain_endpoint_options" {
  description = "Domain endpoint HTTP(S) related options."
  type        = any
  default     = {}
}

variable "domain_endpoint_options_enforce_https" {
  description = "Whether or not to require HTTPS"
  type        = bool
  default     = true
}

variable "domain_endpoint_options_tls_security_policy" {
  description = "The name of the TLS security policy that needs to be applied to the HTTPS endpoint. Valid values: `Policy-Min-TLS-1-0-2019-07` and `Policy-Min-TLS-1-2-2019-07`"
  type        = string
  default     = "Policy-Min-TLS-1-2-2019-07"
}

variable "domain_endpoint_options_custom_endpoint_enabled" {
  description = "Whether to enable custom endpoint for the Elasticsearch domain"
  type        = bool
  default     = false
}

variable "domain_endpoint_options_custom_endpoint" {
  description = "Fully qualified domain for your custom endpoint"
  type        = string
  default     = null
}

variable "domain_endpoint_options_custom_endpoint_certificate_arn" {
  description = "ACM certificate ARN for your custom endpoint"
  type        = string
  default     = null
}

# Advanced options
variable "advanced_options" {
  description = "Key-value string pairs to specify advanced configuration options. Note that the values for these configuration options must be strings (wrapped in quotes) or they may be wrong and cause a perpetual diff, causing Terraform to want to recreate your Elasticsearch domain on every apply"
  type        = map(string)
  default     = {}
}

# ebs_options
variable "ebs_options" {
  description = "EBS related options, may be required based on chosen instance size"
  type        = map(any)
  default     = {}
}

variable "ebs_enabled" {
  description = "Whether EBS volumes are attached to data nodes in the domain"
  type        = bool
  default     = true
}

variable "ebs_options_volume_type" {
  description = "The type of EBS volumes attached to data nodes"
  type        = string
  default     = "gp2"
}

variable "ebs_options_volume_size" {
  description = "The size of EBS volumes attached to data nodes (in GB). Required if ebs_enabled is set to `true`."
  type        = number
  default     = 10
}

variable "ebs_options_iops" {
  description = "The baseline input/output (I/O) performance of EBS volumes attached to data nodes. Applicable only for the Provisioned IOPS EBS volume type"
  type        = number
  default     = 0
}

# encrypt_at_rest
variable "encrypt_at_rest" {
  description = "Encrypt at rest options. Only available for certain instance types"
  type        = map(any)
  default     = {}
}

variable "encrypt_at_rest_enabled" {
  description = "Whether to enable encryption at rest"
  type        = bool
  default     = true
}

variable "encrypt_at_rest_kms_key_id" {
  description = "The KMS key id to encrypt the Elasticsearch domain with. If not specified then it defaults to using the aws/es service KMS key"
  type        = string
  default     = "alias/aws/es"
}

# node_to_node_encryption
variable "node_to_node_encryption" {
  description = "Node-to-node encryption options"
  type        = map(any)
  default     = {}
}

variable "node_to_node_encryption_enabled" {
  description = "Whether to enable node-to-node encryption"
  type        = bool
  default     = true
}

# cluster_config 
variable "cluster_config" {
  description = "Cluster configuration of the domain"
  type        = map(any)
  default     = {}
}

variable "cluster_config_instance_type" {
  description = "Instance type of data nodes in the cluster"
  type        = string
  default     = "t3.small.elasticsearch"
}

variable "cluster_config_instance_count" {
  description = "Number of instances in the cluster"
  type        = number
  default     = 3
}

variable "cluster_config_dedicated_master_enabled" {
  description = "Indicates whether dedicated master nodes are enabled for the cluster"
  type        = bool
  default     = false
}

variable "cluster_config_dedicated_master_type" {
  description = "Instance type of the dedicated master nodes in the cluster"
  type        = string
  default     = "t3.small.elasticsearch"
}

variable "cluster_config_dedicated_master_count" {
  description = "Number of dedicated master nodes in the cluster"
  type        = number
  default     = 1
}

variable "cluster_config_availability_zone_count" {
  description = "Number of Availability Zones for the domain to use with"
  type        = number
  default     = 3
}

variable "cluster_config_zone_awareness_enabled" {
  description = "Indicates whether zone awareness is enabled. To enable awareness with three Availability Zones"
  type        = bool
  default     = false
}

variable "cluster_config_warm_enabled" {
  description = "Indicates whether to enable warm storage"
  type        = bool
  default     = false
}

variable "cluster_config_warm_count" {
  description = "The number of warm nodes in the cluster"
  type        = number
  default     = null
}

variable "cluster_config_warm_type" {
  description = "The instance type for the Elasticsearch cluster's warm nodes"
  type        = string
  default     = null
}

# snapshot_options
variable "snapshot_options" {
  description = "Snapshot related options"
  type        = map(any)
  default     = {}
}

variable "snapshot_options_automated_snapshot_start_hour" {
  description = "Hour during which the service takes an automated daily snapshot of the indices in the domain"
  type        = number
  default     = 0
}

# vpc_options
variable "vpc_options" {
  description = "VPC related options, see below. Adding or removing this configuration forces a new resource"
  type        = map(any)
  default     = {}
}

variable "vpc_options_security_group_ids" {
  description = "List of VPC Security Group IDs to be applied to the Elasticsearch domain endpoints. If omitted, the default Security Group for the VPC will be used"
  type        = list(any)
  default     = []
}

variable "vpc_options_subnet_ids" {
  description = "List of VPC Subnet IDs for the Elasticsearch domain endpoints to be created in"
  type        = list(any)
  default     = []
}

# log_publishing_options 
variable "log_publishing_options" {
  description = "Options for publishing slow logs to CloudWatch Logs"
  type        = map(any)
  default     = {}
}

variable "log_publishing_options_log_type" {
  description = "A type of Elasticsearch log. Valid values: INDEX_SLOW_LOGS, SEARCH_SLOW_LOGS, ES_APPLICATION_LOGS"
  type        = string
  default     = "INDEX_SLOW_LOGS"
}

variable "log_publishing_options_cloudwatch_log_group_arn" {
  description = "iARN of the Cloudwatch log group to which log needs to be published"
  type        = string
  default     = ""
}

variable "log_publishing_options_enabled" {
  description = "Specifies whether given log publishing option is enabled or not"
  type        = bool
  default     = true
}

variable "log_publishing_options_retention" {
  description = "Retention in days for the created Cloudwatch log group"
  type        = number
  default     = 90
}


# cognito_options  
variable "cognito_options" {
  description = "Options for Amazon Cognito Authentication for Kibana"
  type        = map(any)
  default     = {}
}

variable "cognito_options_enabled" {
  description = "Specifies whether Amazon Cognito authentication with Kibana is enabled or not"
  type        = bool
  default     = false
}

variable "cognito_options_user_pool_id" {
  description = "ID of the Cognito User Pool to use"
  type        = string
  default     = ""
}

variable "cognito_options_identity_pool_id" {
  description = "ID of the Cognito Identity Pool to use"
  type        = string
  default     = ""
}

variable "cognito_options_role_arn" {
  description = "ARN of the IAM role that has the AmazonESCognitoAccess policy attached"
  type        = string
  default     = ""
}

variable "tags" {
  description = "A mapping of tags to assign to the resource"
  type        = map(any)
  default     = {}
}


# Timeouts
variable "timeouts" {
  description = "Timeouts map."
  type        = map(any)
  default     = {}
}

variable "timeouts_update" {
  description = "How long to wait for updates."
  type        = string
  default     = null
}

# Service Link Role
variable "create_service_link_role" {
  description = "Create service link role for AWS Elasticsearch Service"
  type        = bool
  default     = false
}
