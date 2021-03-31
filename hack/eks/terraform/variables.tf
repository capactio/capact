
variable "namespace" {
  type        = string
  default     = "capact"
  description = "Prefix, used in all resource names"
}

variable "region" {
  type    = string
  default = "eu-west-1"
}

variable "az_count" {
  type    = number
  default = 1
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "vpc_private_subnets" {
  type    = list(string)
  default = ["10.0.0.0/23", "10.0.2.0/23", "10.0.4.0/23"]
}

variable "vpc_public_subnets" {
  type    = list(string)
  default = ["10.0.100.0/23", "10.0.102.0/23", "10.0.104.0/23"]
}

variable "vpc_single_nat_gateway" {
  type    = bool
  default = true
}

variable "eks_cluster_version" {
  default = "1.18"
}

variable "eks_cluster_endpoint_private_access" {
  default = true
}

variable "eks_cluster_endpoint_public_access" {
  default = true
}

variable "eks_public_access_cidrs" {
  type    = list(string)
  default = []
}

variable "eks_cluster_enabled_log_types" {
  default = ["api", "controllerManager", "scheduler"]
}

variable "worker_group_instance_type" {
  default = "t3a.large"
}

variable "worker_group_max_size" {
  default = 3
}

variable "domain_name" {
  description = "Domain, under which this Capact installation will be available, e.g. 'capact.my-domain.com'"
}

variable "capectl_version" {
  default = "v0.2.1"
}
