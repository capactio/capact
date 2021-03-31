locals {
  eks_cluster_name = "${var.namespace}-cluster"
  eks_public_access_cidrs = concat(var.eks_public_access_cidrs, ["${data.http.public_ip.body}/32"])

  worker_subnets = slice(module.vpc.private_subnets, 0, var.az_count)
}
