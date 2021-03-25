
variable "namespace" {
  type    = string
  default = "voltron"
}

variable "region" {
  type    = string
  default = "eu-west-1"
}

variable "azs" {
  type    = list(string)
  default = ["eu-west-1a", "eu-west-1b", "eu-west-1c"]
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "vpc_private_subnets" {
  type    = list(string)
  default = ["10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"]
}

variable "vpc_public_subnets" {
  type    = list(string)
  default = ["10.0.101.0/24", "10.0.102.0/24", "10.0.103.0/24"]
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
  default = true # disable this
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
