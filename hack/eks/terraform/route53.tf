module "zones" {
  source  = "terraform-aws-modules/route53/aws//modules/zones"
  version = "~> 1.0"
  tags    = local.tags

  zones = {
    (var.domain_name) = {
    }
  }
}

