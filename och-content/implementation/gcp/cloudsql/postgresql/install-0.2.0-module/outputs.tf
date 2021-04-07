output "instance_ip_addr" {
  value = google_sql_database_instance.master.public_ip_address
}

output "username" {
  value = google_sql_user.users.name
}

output "password" {
  value = google_sql_user.users.password
  sensitive = true
}
