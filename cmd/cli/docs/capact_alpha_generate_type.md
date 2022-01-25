---
title: capact alpha generate type
---

## capact alpha generate type

Generate new Type manifests

```
capact alpha generate type [PATH] [flags]
```

### Examples

```
# Generate manifests for the cap.type.database.postgresql.config Type
capact alpha manifest-gen type cap.type.database.postgresql.config
```

### Options

```
  -h, --help              help for type
  -r, --revision string   Revision of the Type manifest (default "0.1.0")
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

