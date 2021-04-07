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
