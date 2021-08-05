---
title: capact alpha manifest-gen interface
---

## capact alpha manifest-gen interface

Bootstrap new Interface manifests

### Synopsis

Bootstrap new Interface and associated Type manifests

```
capact alpha manifest-gen interface [PATH] [flags]
```

### Examples

```
# Bootstrap manifests for the cap.interface.database.postgresql.install Interface
capact alpha content interface database.postgresql install
```

### Options

```
  -h, --help              help for interface
  -r, --revision string   Revision of the Interface manifest (default "0.1.0")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
  -o, --output string   Path to the output directory for the generated manifests (default "generated")
      --override        Override existing manifest files
```

### SEE ALSO

* [capact alpha manifest-gen](capact_alpha_manifest-gen.md)	 - Manifests generation

