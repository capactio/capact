# Capact EKS deployment

This directory stores Terraform modules and scripts for an Capact deployment on Amazon EKS.

## How to

1. Initialize the Terraform module. It's recommended to use remote state file on S3:
```
terraform init -backend-config=bucket=voltron-terraform-states -backend-config=key=voltron-eks.tfstate -backend-config=region=eu-west-1
```

2. Apply the Terraform module. Terraform needs API server access to finish the configuration, so we have to add our IP to the allowed ones.
```
# TODO maybe this can be moved to a terraform local-exec
terraform apply -var "eks_public_access_cidrs=[\"$(curl 'https://api.ipify.org')/32\"]"
```

## Limitations

- You have to add the self-signed certificate to the bastion host manually
- You have to configure the DNS pointing to your application manually
