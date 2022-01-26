---
title: capact manifest generate attribute
---

## capact manifest generate attribute

Generate new Attribute manifests

```
capact manifest generate attribute [PATH] [flags]
```

### Examples

```
# Generate manifests for the cap.attribute.cloud.provider.aws Attribute
capact manifest generate attribute cap.attribute.cloud.provider.aws
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

* [capact manifest generate](capact_manifest_generate.md)	 - OCF Manifests generation

