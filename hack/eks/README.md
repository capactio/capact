# Capact EKS deployment

This directory stores Terraform modules and scripts for a Capact deployment on Amazon EKS.

## Architecture

![Diagram](./assets/Capact_EKS.svg)

> For now the worker nodes are deployed only in a single availability zone.

## Prerequisites

- S3 bucket for the remote Terraform state file
- AWS account with AdministratorAccess permissions
- [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
- A domain name for the Capact installation

## Installation

1. Set the following environment variables:
```
export CAPACT_NAME=<name_of_the_environment>
export CAPACT_REGION=<aws_region_in_which_to_deploy_capact>
export CAPACT_DOMAIN_NAME=<domain_name_used_for_the_capact_environment>
export CAPACT_DOCKER_TAG=<capact_version_to_install>
export TERRAFORM_STATE_BUCKET=<s3_bucket_for_the_remote_statefile>
```

2. Run `./install.sh`. This can take around to 20 minutes to finish.
3. Configure the name servers for the Capact Route53 Hosted Zone in your DNS provider. To get the name server for the hosted zone check `config/route53_zone_name_servers`
```bash
cat config/route53_zone_name_servers
```
```bash
{
  "aws-1.cluster.projectvoltron.dev": [
    "ns-1260.awsdns-29.org",
    "ns-1586.awsdns-06.co.uk",
    "ns-444.awsdns-55.com",
    "ns-945.awsdns-54.net"
  ]
}
```

## Access API server from the bastion host

The bastion hosts has `kubectl` preinstalled and `kubeconfig` configured to the EKS cluster API server. SSH to the bastion using the following command:
```bash
ssh -i config/bastion_ssh_private_key ec2-user@$(cat config/bastion_public_ip)
```

Now you should be able to query the API server:
```bash
kubectl get nodes
```

## Using capact CLI from the bastion host

The bastion host can access the Capact gateway and has `capectl` preinstalled.

> The current version of the released capectl does not support Gateway API. You have to build it from source and upload to the bastion

1. SSH to the bastion host:
```bash
ssh -i config/bastion_ssh_private_key ec2-user@$(cat config/bastion_public_ip)
```

2. Login the gateway:
```bash
capectl login <capact-gateway-address>
```

3. Verify, if you can query the Capact Gateway and list all interfaces in the OCH:
```bash
capectl hub interfaces search
```

## Limitations and bugs

- Before running `terraform destroy` you have to remove all the entries from the Route53 Hosted Zone and the Load Balancers created for the LoadBalancer service. In other case destroy will fail.
- There is [an issue](https://github.com/terraform-aws-modules/terraform-aws-eks/issues/1162), with the EKS module, where `terraform destroy` fails on the resource `module.eks.kubernetes_config_map.aws_auth[0]`. You don't have to worry about this, just confirm all other resources are destroyed and remove the resource manually from the state file using `terraform state rm 'module.eks.kubernetes_config_map.aws_auth[0]'`.
