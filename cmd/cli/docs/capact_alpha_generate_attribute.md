---
title: capact alpha generate attribute
---

## capact alpha generate attribute

Generate new Attribute manifests

```
capact alpha generate attribute [PATH] [flags]
```

### Examples

```
# Generate manifests for the cap.attribute.cloud.provider.aws Attribute
capact alpha manifest-gen attribute cap.attribute.cloud.provider.aws
```

### Options

```
  -h, --help              help for attribute
  -r, --revision string   Revision of the Attribute manifest (default "0.1.0")
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -o, --output string                 Path to the output directory for the generated manifests (default "generated")
      --overwrite                     Overwrite existing manifest files
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact alpha generate](capact_alpha_generate.md)	 - OCF Manifests generation

