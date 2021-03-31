# Capact EKS deployment

This directory stores Terraform modules and scripts for an Capact deployment on Amazon EKS.

## Prerequisites

- S3 bucket for the remote Terraform state file
- AWS account
- A domain name for the Capact installation

## Installation

1. Run `install.sh`
2. Configure the name servers for the Capact Route53 Hosted Zone in your DNS provider.
3. Register the ALIAS record in the Hosted Zone for the Capact gateway, pointing to the internal load balancer.

## Running Capact actions from the bastion host

The bastion host can access the Capact gateway and has `capectl` preinstalled.

1. SSH to the bastion host
2. Login the gateway
3. Run an action

> Domain names for the application ingress are not automatically configured on the Route53 Hosted Zone and must be configured manually by the Capact administrator

## Limitations

- You have to configure the DNS entries in the Route53 hosted zone manually
