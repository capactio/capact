---
title: capact alpha manifest-gen interfacegroup
---

## capact alpha manifest-gen interfacegroup

Generate new InterfaceGroup manifest

```
capact alpha manifest-gen interfacegroup [PATH] [flags]
```

### Examples

```
# Generate manifests for the cap.interface.database.postgresql InterfaceGroup
capact alpha manifest-gen interfacegroup cap.interface.database.postgresql
```

### Options

```
  -h, --help              help for interfacegroup
  -r, --revision string   Revision of the InterfaceGroup manifest (default "0.1.0")
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -o, --output string                 Path to the output directory for the generated manifests (default "generated")
      --overwrite                     Overwrite existing manifest files
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact alpha manifest-gen](capact_alpha_manifest-gen.md)	 - Manifests generation

