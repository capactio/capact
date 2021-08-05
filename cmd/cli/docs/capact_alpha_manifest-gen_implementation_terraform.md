---
title: capact alpha manifest-gen implementation terraform
---

## capact alpha manifest-gen implementation terraform

Bootstrap Terraform based manifests

### Synopsis

Bootstrap Terraform based manifests based on a Terraform module

```
capact alpha manifest-gen implementation terraform [PREFIX] [NAME] [TERRAFORM_MODULE_PATH] [flags]
```

### Examples

```
# Bootstrap manifests 
capact alpha content implementation terraform aws.rds.deploy ./terraform-modules/aws-rds

# Bootstrap manifests for an AWS Terraform module
capact alpha content implementation terraform aws.rds.deploy ./terraform-modules/aws-rds -p aws
	
# Bootstrap manifests for an GCP Terraform module
capact alpha content implementation terraform gcp.cloudsql.deploy ./terraform-modules/cloud-sql -p gcp
```

### Options

```
  -h, --help               help for terraform
  -i, --interface string   Path with revision of the Interface, which is implemented by this Implementation
  -p, --provider string    Create a provider-specific workflow. Possible values: "aws", "gcp"
  -r, --revision string    Revision of the Implementation manifest (default "0.1.0")
  -s, --source string      URL to the tarball with the Terraform module (default "https://example.com/terraform-module.tgz")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
  -o, --output string   Path to the output directory for the generated manifests (default "generated")
      --override        Override existing manifest files
```

### SEE ALSO

* [capact alpha manifest-gen implementation](capact_alpha_manifest-gen_implementation.md)	 - Bootstrap new Implementation manifests

