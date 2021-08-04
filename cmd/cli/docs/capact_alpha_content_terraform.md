---
title: capact alpha content terraform
---

## capact alpha content terraform

Bootstrap Terraform based manifests

### Synopsis

Bootstrap Terraform based manifests based on a Terraform module

```
capact alpha content terraform [PREFIX] [NAME] [TERRAFORM_MODULE_PATH] [flags]
```

### Examples

```
# Bootstrap manifests 
	capact alpha content terraform aws.rds deploy ./terraform-modules/aws-rds
```

### Options

```
  -h, --help               help for terraform
  -i, --interface string   Path with revision of the Interface, which is implemented by this Implementation
  -r, --revision string    Revision of the Implementation manifest (default "0.1.0")
  -s, --source string      URL to the tarball with the Terraform module (default "https://example.com/terraform-module.tgz")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
  -o, --output string   Path to the output directory for the generated manifests (default "generated")
```

### SEE ALSO

* [capact alpha content](capact_alpha_content.md)	 - Content generation

