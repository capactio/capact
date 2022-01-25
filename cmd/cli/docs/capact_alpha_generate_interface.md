---
title: capact alpha generate interface
---

## capact alpha generate interface

Generate new Interface-related manifests

### Synopsis

Generate new InterfaceGroup, Interface and associated Type manifests

```
capact alpha generate interface [PATH] [flags]
```

### Examples

```
# Generate manifests for the cap.interface.database.postgresql.install Interface
capact alpha manifest-gen interface cap.interface.database.postgresql.install
```

### Options

```
  -h, --help              help for interface
  -r, --revision string   Revision of the Interface manifest (default "0.1.0")
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

