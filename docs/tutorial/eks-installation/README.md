# Capact EKS deployment

This tutorial shows how to set up a private Amazon Elastic Kubernetes Service (Amazon EKS) cluster with full Capact installation using Terraform.

<!-- toc -->

- [Architecture](#architecture)
- [Prerequisites](#prerequisites)
- [Installation](#installation)
- [Access API server from the bastion host](#access-api-server-from-the-bastion-host)
- [Use Capact CLI from the bastion host](#use-capact-cli-from-the-bastion-host)
- [Connect to Capact Gateway from local machine](#connect-to-capact-gateway-from-local-machine)
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

1. Set the required environment variables by running:
   
  ```bash
  export CAPACT_NAME={name_of_the_environment}
  export CAPACT_REGION={aws_region_in_which_to_deploy_capact}
  export CAPACT_DOMAIN_NAME={domain_name_used_for_the_capact_environment}
  export CAPACT_DOCKER_TAG={capact_version_to_install}
  export TERRAFORM_STATE_BUCKET={s3_bucket_for_the_remote_statefile}
  ```

1. Configure optional parameters.

   - By default, the cluster worker nodes are created in a single availability zone. To increase the number of availability zones, where the cluster worker nodes are created, run:
     
     ```bash
     export EKS_AZ_COUNT={number_of_availability_zones}
     ```
     
   - To enable [Amazon Elastic File System](https://aws.amazon.com/efs/) configuration for the EKS cluster, run:
      
     ```bash
     export EKS_EFS_ENABLED=true
     ```
     
     If this option is enabled, after following this tutorial, the `efs-sc` StorageClass will be available to use in your Kubernetes cluster. 

   - To add custom flags for `terraform apply` command, set the `CAPACT_TERRAFORM_OPTS` environmental variable. For example, run:
      
     ```bash
     export CAPACT_TERRAFORM_OPTS="-var worker_group_max_size=4"` 
     ```

1. Run the installation script:
   
   ```bash
   ./hack/eks/install.sh
   ```

   - When you see the "Do you want to perform these actions?" question, provide `yes` value in the command line and press enter. 
     
   This operation can take around to 20 minutes to finish.
   
1. Configure the name servers for the Capact Route53 Hosted Zone in your DNS provider. To get the name server for the hosted zone check the [`config/route53_zone_name_servers`](./config/route53_zone_name_servers) file.
  ```bash
  cat hack/eks/config/route53_zone_name_servers
  ```
  ```bash
  {
    "aws-1.cluster.capact.dev": [
      "ns-1260.awsdns-29.org",
      "ns-1586.awsdns-06.co.uk",
      "ns-444.awsdns-55.com",
      "ns-945.awsdns-54.net"
    ]
  }
  ```

## Access API server from the bastion host

The bastion hosts has `kubectl` preinstalled and `kubeconfig` configured to the EKS cluster API server. SSH to the bastion using the following command from:
```bash
ssh -i hack/eks/config/bastion_ssh_private_key ubuntu@$(cat hack/eks/config/bastion_public_ip)
```

Now you should be able to query the API server:
```bash
kubectl get nodes
```

## Use Capact CLI from the bastion host

The bastion host can access the Capact gateway and has `capectl` preinstalled.

> **NOTE**: The current version of the released capectl does not support Gateway API. You have to build it from the source and upload to the bastion host.

1. SSH to the bastion host:
  ```bash
  ssh -i hack/eks/config/bastion_ssh_private_key ubuntu@$(cat hack/eks/config/bastion_public_ip)
  ```

1. Get the address and credentials to the Capact Gateway:
  ```bash
  # get the gateway address
  kubectl -n capact-system get ingress capact-gateway -ojsonpath='{.spec.rules[0].host}'

  # get the gateway username
  kubectl -n capact-system get deployment capact-gateway -oyaml | grep -A1 "name: APP_AUTH_USERNAME" | tail -1 | awk -F ' ' '{print $2}'

  # get the gateway password
  kubectl -n capact-system get deployment capact-gateway -oyaml | grep -A1 "name: APP_AUTH_PASSWORD" | tail -1 | awk -F ' ' '{print $2}'
  ```

1. Login to the cluster:
  ```bash
  capectl login {gateway-address}
  ```

1. Verify, if you can query the Capact Gateway and list all Interfaces in the OCH:
  ```bash
  capectl hub interfaces search
  ```

## Connect to Capact Gateway from local machine

Only the bastion host can access the Capact Gateway. To be able to connect to the Gateway, you need to proxy your traffic.

1. Open SSH tunnel:
   ```bash
   ssh -f -M -N -S /tmp/gateway.${CAPACT_DOMAIN_NAME}.sock -i hack/eks/config/bastion_ssh_private_key ubuntu@$(cat hack/eks/config/bastion_public_ip) -L 127.0.0.1:8081:gateway.${CAPACT_DOMAIN_NAME}:443
   ``` 

1. Add new entry to `/etc/hosts`:
   ```bash
   export LINE_TO_APPEND="127.0.0.1 gateway.${CAPACT_DOMAIN_NAME}"
   export HOSTS_FILE="/etc/hosts"
   
   grep -qxF -- "$LINE_TO_APPEND" "${HOSTS_FILE}" || (echo "$LINE_TO_APPEND" | sudo tee -a "${HOSTS_FILE}" > /dev/null)
   ```

1. Test connection:
   
   1. Using Capact CLI 
   ```bash
   capact login https://gateway.${CAPACT_DOMAIN_NAME}:8081 -u {user} -p {password}
   ```

   2. Using Browser. Navigate to Gateway GraphQL Playground `https://gateway.${CAPACT_DOMAIN_NAME}:8081/graphql`.

1. When you are done, close the connection:

   ```bash
   ssh -S /tmp/gateway.${CAPACT_DOMAIN_NAME}.sock -O exit $(cat hack/eks/config/bastion_public_ip)
   ``` 

## Cleanup

1. Remove the ingress-nginx and public-ingress-nginx Helm releases. This is required to deprovision the AWS ELBs.
  ```bash
  helm delete -n ingress-nginx ingress-nginx
  helm delete -n public-ingress-nginx public-ingress-nginx
  ```

1. Remove the records from the Route53 Hosted Zone from the AWS Console. Only the entries for apex SOA and NS should be left.

1. Deprovision the EKS cluster and VPC.
 
  ```bash
  cd hack/eks/terraform
  # This command might fail. See "Limitations and bugs" section.
  terraform destroy -var domain_name=$CAPACT_DOMAIN_NAME
  ```

  If the previous command failed execute the following commands:
 
  ```bash
  terraform state rm 'module.eks.kubernetes_config_map.aws_auth[0]'
  terraform destroy -var domain_name=$CAPACT_DOMAIN_NAME
  ```

## Limitations and bugs

- There is [an issue](https://github.com/terraform-aws-modules/terraform-aws-eks/issues/1162), with the EKS module, where `terraform destroy` fails on the resource `module.eks.kubernetes_config_map.aws_auth[0]`. You don't have to worry about this, just remove the resource manually from the state file using `terraform state rm 'module.eks.kubernetes_config_map.aws_auth[0]'` and run `terraform destroy again`
