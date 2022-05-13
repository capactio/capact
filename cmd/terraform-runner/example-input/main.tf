variable "database_version" {
  type        = string
  default     = "POSTGRES_12"
  description = "CloudSQL database version"
}

variable "region" {
  type        = string
  default     = "us-central"
  description = "Google cloud zone"
}

variable "tier" {
  type        = string
  default     = "db-f1-micro"
  description = "CloudSQL instance tier"
}

variable "user_name" {
  type        = string
  description = "Database user name"
}

variable "user_password" {
  type        = string
  description = "Database user password"
  sensitive   = true
}

terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "4.20.0"
    }
  }
}

provider "google" {
}

resource "google_sql_database_instance" "master" {
  database_version    = var.database_version
  region              = var.region
  deletion_protection = false

  settings {
    tier = var.tier
    ip_configuration {
      authorized_networks {
        value = "0.0.0.0/0"
      }
    }
    database_flags {
      name  = "max_connections"
      value = 1000
    }
  }
}

resource "google_sql_user" "users" {
  name     = var.user_name
  instance = google_sql_database_instance.master.name
  password = var.user_password
}

output "instance_ip_addr" {
  value = google_sql_database_instance.master.public_ip_address
}

output "username" {
  value = google_sql_user.users.name
}

output "password" {
  value = nonsensitive(google_sql_user.users.password)
}
