locals {
  eks_cluster_name        = "${var.namespace}-cluster"
  eks_public_access_cidrs = concat(var.eks_public_access_cidrs, ["${data.http.public_ip.body}/32"])

  worker_subnets = slice(var.vpc_private_subnets, 0, var.az_count)

  route53_zone_id = module.zones.this_route53_zone_zone_id[var.domain_name]

  tags = {
    Application = "Capact"
    "Domain-Name" = var.domain_name
  }
}
