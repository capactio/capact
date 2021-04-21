
variable "namespace" {
  type        = string
  default     = "capact"
  description = "Prefix, used in all resource names"
}

variable "region" {
  type    = string
  default = "eu-west-1"
  description = "Region, in which Capact will be deployed"
}

variable "az_count" {
  type    = number
  default = 1
  description = "Number of AZs, in which the worker nodes will be created"
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
  description = "CIDR of the created VPC"
}

variable "vpc_private_subnets" {
  type    = list(string)
  default = ["10.0.0.0/23", "10.0.2.0/23", "10.0.4.0/23"]
  description = "CIDRs for the private subnets"
}

variable "vpc_public_subnets" {
  type    = list(string)
  default = ["10.0.100.0/23", "10.0.102.0/23", "10.0.104.0/23"]
  description = "CIDRs for the public subnets"
}

variable "vpc_single_nat_gateway" {
  type    = bool
  default = true
  description = "Boolean indicating, if only a single NAT gateway will be deployed in the VPC"
}

variable "eks_cluster_version" {
  default = "1.18"
  description = "Version of the EKS cluster"
}

variable "eks_cluster_endpoint_private_access" {
  default = true
  description = "Enable EKS private cluster endpoint"
}

variable "eks_cluster_endpoint_public_access" {
  default = true
  description = "Enable EKS public cluster endpoint"
}

variable "eks_public_access_cidrs" {
  type    = list(string)
  default = []
  description = "Additional CIDRs allowed to access the EKS public cluster endpoint"
}

variable "eks_cluster_enabled_log_types" {
  default = ["api", "controllerManager", "scheduler"]
  description = "List of EKS logs, which will be pushed to CloudWatch Logs"
}

variable "worker_group_instance_type" {
  default = "t3a.large"
  description = "Instance type of the worker nodes"
}

variable "worker_group_max_size" {
  default = 3
  description = "Maximum size of the worker nodes Autoscaling Group"
}

variable "domain_name" {
  description = "Domain, under which this Capact installation will be available, e.g. 'capact.my-domain.com'"
}

variable "capectl_version" {
  default = "v0.2.1"
  description = "Version of the capectl binary, installed on the bastion host"
}

variable "efs_enabled" {
  default = false
  description = "Enables EFS storage configuration for EKS cluster"
}
