variable "database_version" {
  type = string
  default = "POSTGRES_12"
  description = "CloudSQL database version"
}

variable "region" {
  type = string
  default = "us-central"
  description = "Google cloud zone"
}

variable "tier" {
  type = string
  default = "db-f1-micro"
  description = "CloudSQL instance tier"
}

variable "user_name" {
  type = string
  description = "Database user name"
}

variable "user_password" {
  type = string
  description = "Database user password"
  sensitive = true
}
