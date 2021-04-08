# Capact EKS deployment

This tutorial shows how to set up a private Amazon Elastic Kubernetes Service (Amazon EKS) cluster with full Voltron installation using Terraform.

<!-- toc -->

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Access API server from the bastion host](#access-api-server-from-the-bastion-host)
- [Using capact CLI from the bastion host](#using-capact-cli-from-the-bastion-host)
- [Cleanup](#cleanup)
- [Limitations and bugs](#limitations-and-bugs)

<!-- tocstop -->

## Architecture

![Diagram](./assets/Capact_EKS.svg)

> **NOTE**: For now the worker nodes are deployed only in a single availability zone.

## Prerequisites

- S3 bucket for the remote Terraform state file
- AWS account with **AdministratorAccess** permissions on it
- [AWS CLI v2](https://docs.aws.amazon.com/cli/latest/userguide/install-cliv2.html)
- A domain name for the Capact installation

To configure the AWS CLI follow [this](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html) guide.

If you use AWS SSO on your account, then you can also configure SSO for AWS CLI instead of creating an IAM user. [This page](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html) shows how to configure AWS CLI with AWS SSO.

## Installation

1. Set the following environment variables:
  ```
  export CAPACT_NAME=<name_of_the_environment>
  export CAPACT_REGION=<aws_region_in_which_to_deploy_capact>
  export CAPACT_DOMAIN_NAME=<domain_name_used_for_the_capact_environment>
  export CAPACT_DOCKER_TAG=<capact_version_to_install>
  export TERRAFORM_STATE_BUCKET=<s3_bucket_for_the_remote_statefile>
  ```

> **NOTE:** You can add flags to the terraform apply, by settings the CAPACT_TERRAFORM_OPTS environment variable, e.g.
>
> `export CAPACT_TERRAFORM_OPTS="-var worker_group_max_size=4"`

2. Run `./install.sh`. This can take around to 20 minutes to finish.
3. Configure the name servers for the Capact Route53 Hosted Zone in your DNS provider. To get the name server for the hosted zone check the [`config/route53_zone_name_servers`](./config/route53_zone_name_servers) file.
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

The bastion hosts has `kubectl` preinstalled and `kubeconfig` configured to the EKS cluster API server. SSH to the bastion using the following command from the:
```bash
ssh -i hack/eks/config/bastion_ssh_private_key ubuntu@$(cat hack/eks/config/bastion_public_ip)
```

Now you should be able to query the API server:
```bash
kubectl get nodes
```

## Using capact CLI from the bastion host

The bastion host can access the Capact gateway and has `capectl` preinstalled.

> **NOTE**: The current version of the released capectl does not support Gateway API. You have to build it from source and upload to the bastion

1. SSH to the bastion host:
  ```bash
  ssh -i hack/eks/config/bastion_ssh_private_key ubuntu@$(cat hack/eks/config/bastion_public_ip)
  ```

2. Get the address and credentials to the Capact gateway:
  ```bash
  # get the gateway address
  kubectl -n voltron-system get ingress voltron-gateway -ojsonpath='{.spec.rules[0].host}'

  # get the gateway username
  kubectl -n voltron-system get deployment voltron-gateway -oyaml | grep -A1 "name: APP_AUTH_USERNAME" | tail -1 | awk -F ' ' '{print $2}'

  # get the gateway password
  kubectl -n voltron-system get deployment voltron-gateway -oyaml | grep -A1 "name: APP_AUTH_PASSWORD" | tail -1 | awk -F ' ' '{print $2}'
  ```

3. Login the gateway:
  ```bash
  capectl login <gateway-address>
  ```

4. Verify, if you can query the Capact Gateway and list all interfaces in the OCH:
  ```bash
  capectl hub interfaces search
  ```

## Cleanup

1. Remove the ingress-nginx and public-ingress-nginx Helm releases. This is required to deprovision the AWS ELBs.
  ```bash
  helm delete -n ingress-nginx ingress-nginx
  helm delete -n public-ingress-nginx public-ingress-nginx
  ```

2. Remove the A entries from the Route53 Hosted Zone from the AWS Console. Only the entries for apex SOA and NS should be left.

3. Deprovision the EKS cluster and VPC.
  ```bash
  cd hack/eks/terraform

  # This command might fail. See "Limitations and bugs" section.
  terraform destroy -var domain_name=$CAPACT_DOMAIN_NAME

  # If the previous command failed execute the following two commands.
  terraform state rm 'module.eks.kubernetes_config_map.aws_auth[0]'
  terraform destroy -var domain_name=$CAPACT_DOMAIN_NAME
  ```

## Limitations and bugs

- Before running `terraform destroy` you have to remove all the entries from the Route53 Hosted Zone and the Load Balancers created for the LoadBalancer service. In other case destroy will fail.
- There is [an issue](https://github.com/terraform-aws-modules/terraform-aws-eks/issues/1162), with the EKS module, where `terraform destroy` fails on the resource `module.eks.kubernetes_config_map.aws_auth[0]`. You don't have to worry about this, just remove the resource manually from the state file using `terraform state rm 'module.eks.kubernetes_config_map.aws_auth[0]'` and run `terraform destroy again`
