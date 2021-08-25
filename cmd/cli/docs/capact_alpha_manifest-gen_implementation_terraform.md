---
title: capact alpha manifest-gen implementation terraform
---

## capact alpha manifest-gen implementation terraform

Generate Terraform based manifests

### Synopsis

Generate Implementation manifests based on a Terraform module

```
capact alpha manifest-gen implementation terraform [MANIFEST_PATH] [TERRAFORM_MODULE_PATH] [flags]
```

### Examples

```
# Generate Implementation manifests 
capact alpha manifest-gen implementation terraform cap.implementation.aws.rds.deploy ./terraform-modules/aws-rds

# Generate Implementation manifests for an AWS Terraform module
capact alpha manifest-gen implementation terraform cap.implementation.aws.rds.deploy ./terraform-modules/aws-rds -p aws
	
# Generate Implementation manifests for an GCP Terraform module
capact alpha manifest-gen implementation terraform cap.implementation.gcp.cloudsql.deploy ./terraform-modules/cloud-sql -p gcp
```

### Options

```
  -h, --help               help for terraform
  -i, --interface string   Path with revision of the Interface, which is implemented by this Implementation
  -p, --provider string    Create a provider-specific workflow. Possible values: "aws", "gcp"
  -r, --revision string    Revision of the Implementation manifest (default "0.1.0")
  -s, --source string      Path to the Terraform module, such as URL to Tarball or Git repository (default "https://example.com/terraform-module.tgz")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
  -o, --output string   Path to the output directory for the generated manifests (default "generated")
      --overwrite       Overwrite existing manifest files
```

### SEE ALSO

* [capact alpha manifest-gen implementation](capact_alpha_manifest-gen_implementation.md)	 - Generate new Implementation manifests

