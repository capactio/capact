---
title: capact alpha generate implementation terraform
---

## capact alpha generate implementation terraform

Generate Terraform based manifests

### Synopsis

Generate Implementation manifests based on a Terraform module

```
capact alpha generate implementation terraform [MANIFEST_PATH] [TERRAFORM_MODULE_PATH] [flags]
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
  -c, --config string                 Path to the YAML config file
  -o, --output string                 Path to the output directory for the generated manifests (default "generated")
      --overwrite                     Overwrite existing manifest files
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact alpha generate implementation](capact_alpha_generate_implementation.md)	 - Generate new Implementation manifests

