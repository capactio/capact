output "arn" {
  description = "Amazon Resource Name (ARN) of the domain"
  value       = module.main.arn
}

output "domain_id" {
  description = "Unique identifier for the domain"
  value       = module.main.domain_id
}

output "domain_name" {
  description = "Unique identifier for the domain"
  value       = var.domain_name
}

output "endpoint" {
  description = "Domain-specific endpoint used to submit index, search, and data upload requests"
  value       = module.main.endpoint
}

output "kibana_endpoint" {
  description = "Domain-specific endpoint for kibana without https scheme"
  value       = module.main.kibana_endpoint
}

output "vpc_options_availability_zones" {
  description = "If the domain was created inside a VPC, the names of the availability zones the configured subnet_ids were created inside"
  value       = module.main.vpc_options_availability_zones
}

output "vpc_options_vpc_id" {
  description = "If the domain was created inside a VPC, the ID of the VPC"
  value       = module.main.vpc_options_vpc_id
}

output "elasticsearch_version" {
  description = "Version of the Elasticsearch domain"
  value = var.elasticsearch_version
}
