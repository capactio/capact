variable "engine" {
  type = string
  description = "Database engine"
  default = "postgres"
}

variable "engine_version" {
  type = string
  default = "11.10"
  description = "RDS database engine version"
}

variable "major_engine_version" {
  type = string
  default = "11"
  description = "PostgreSQL major engine version"
}

variable "region" {
  type = string
  description = "AWS region"
}

variable "tier" {
  type = string
  default = "db.t3.micro"
  description = "AWS RDS instance tier"
}

variable "ingress_rule_cidr_blocks" {
  description = "CIDR blocks for ingress rule. For public access provide '0.0.0.0/0'."
  type = string
  default = ""
}

variable "res_name" {
  type = string
  description = "Name used for the resources"
  default = ""
}

variable "publicly_accessible" {
  description = "Bool to control if instance is publicly accessible"
  type = bool
  default = false
}

variable "allocated_storage" {
  description = "The allocated storage in gigabytes"
  type = string
  default = 20
}

variable "max_allocated_storage" {
  description = "Specifies the value for Storage Autoscaling"
  type = number
  default = 100
}

variable "storage_encrypted" {
  description = "Specifies whether the DB instance is encrypted"
  type = bool
  default = true
}

variable "multi_az" {
  description = "Specifies if the RDS instance is multi-AZ"
  type = bool
  default = false
}

variable "deletion_protection" {
  description = "The database can't be deleted when this value is set to true."
  type = bool
  default = false
}

variable "backup_retention_period" {
  description = "The days to retain backups for"
  type = number
  default = null
}

variable "performance_insights_enabled" {
  description = "Specifies whether Performance Insights are enabled"
  type = bool
  default = false
}

variable "performance_insights_retention_period" {
  description = "The amount of time in days to retain Performance Insights data. Either 7 (7 days) or 731 (2 years)."
  type = number
  default = 7
}

variable "monitoring_interval" {
  description = "The interval, in seconds, between points when Enhanced Monitoring metrics are collected for the DB instance. To disable collecting Enhanced Monitoring metrics, specify 0. The default is 0. Valid Values: 0, 1, 5, 10, 15, 30, 60."
  type = number
  default = 60
}

variable "maintenance_window" {
  description = "The window to perform maintenance in. Syntax: 'ddd:hh24:mi-ddd:hh24:mi'. Eg: 'Mon:00:00-Mon:03:00'"
  type = string
  default = "Mon:00:00-Mon:03:00"
}

variable "backup_window" {
  description = "The daily time range (in UTC) during which automated backups are created if they are enabled. Example: '09:46-10:16'. Must not overlap with maintenance_window"
  type = string
  default = "03:00-06:00"
}

variable "skip_final_snapshot" {
  description = "Determines whether a final DB snapshot is created before the DB instance is deleted. If true is specified, no DBSnapshot is created. If false is specified, a DB snapshot is created before the DB instance is deleted, using the value from final_snapshot_identifier"
  type = bool
  default = false
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
