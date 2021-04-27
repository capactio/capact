output "instance_ip_addr" {
  description = "The address of the RDS instance"
  value = module.db.this_db_instance_address
}

output "port" {
  description = "The database port"
  value = module.db.this_db_instance_port
}

output "defaultDBName" {
  description = "The master username for the database"
  value = var.engine == "postgres" ? "postgres" : "" # no default db for MySQL
}

output "username" {
  description = "The master username for the database"
  value = module.db.this_db_instance_username
  sensitive = true
}

output "password" {
  description = "The database password"
  value = module.db.this_db_instance_password
  sensitive = true
}

output "identifier" {
  description = "The AWS RDS instance identifier"
  value = module.db.this_db_instance_id
}

output "availability_zone" {
  description = "Availability zones"
  value = module.db.this_db_instance_availability_zone
}

output "class" {
  description = "AWS RDS instance class"
  value = var.tier
}
