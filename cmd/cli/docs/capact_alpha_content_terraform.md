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
	capact alpha content terraform aws.rds deploy ../hub-manifests/manifests/implementation/aws/rds/postgresql/provision-module
```

### Options

```
  -h, --help               help for terraform
  -i, --interface string   Path of the Interface, which is implemented by this Implementation
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
  -o, --output string   Path to the output directory for the generated manifests (default "generated")
```

### SEE ALSO

* [capact alpha content](capact_alpha_content.md)	 - Content generation

