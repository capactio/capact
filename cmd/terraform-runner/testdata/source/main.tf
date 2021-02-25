/*
variable "project" {
  type = string
  description = "Google Cloud project"
} 
*/

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
  sensitive = true
  description = "Database user password"
}

terraform {
  required_providers {
    google = {
      source = "hashicorp/google"
      version = ">=3.55.0"
    }
  }
}

provider "google" {
}


resource "google_sql_database_instance" "master" {
  database_version = var.database_version
  region           = var.region

  settings {
    tier = var.tier
    ip_configuration {
      authorized_networks {
        value = "0.0.0.0/0"
      }
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
