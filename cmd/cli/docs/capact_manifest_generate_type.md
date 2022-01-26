---
title: capact manifest generate type
---

## capact manifest generate type

Generate new Type manifests

```
capact manifest generate type [PATH] [flags]
```

### Examples

```
# Generate manifests for the cap.type.database.postgresql.config Type
capact manifest generate type cap.type.database.postgresql.config
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

* [capact manifest generate](capact_manifest_generate.md)	 - OCF Manifests generation

