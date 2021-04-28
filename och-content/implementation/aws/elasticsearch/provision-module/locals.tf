data "aws_caller_identity" "current" {}

locals {
  account_id = data.aws_caller_identity.current.account_id

  access_policies = var.access_policies != "" ? var.access_policies : jsonencode({
    "Version": "2012-10-17",
    "Statement": [
      {
        "Effect": "Allow",
        "Principal": {
          "AWS": "*"
        },
        "Action": "es:*",
        "Resource": "arn:aws:es:${var.region}:${local.account_id}:domain/${var.domain_name}/*"
      }
    ]
  })
}
