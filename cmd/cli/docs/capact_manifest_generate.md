---
title: capact manifest generate
---

## capact manifest generate

OCF Manifests generation

### Synopsis

Subcommand for various manifest generation operations

```
capact manifest generate [flags]
```

### Examples

```
# To generate manifests interactively, run: 
capact manifest generate
# Then, select which manifests kinds you want to generate.
# If the Interface is selected, the Type kind toggles
# input and output Type generation for a given Interface.
```

### Options

```
  -h, --help            help for generate
  -o, --output string   Path to the output directory for the generated manifests (default "generated")
      --overwrite       Overwrite existing manifest files
```

### Options inherited from parent commands

```
  -C, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact manifest](capact_manifest.md)	 - This command consists of multiple subcommands to interact with OCF manifests
* [capact manifest generate attribute](capact_manifest_generate_attribute.md)	 - Generate new Attribute manifests
* [capact manifest generate implementation](capact_manifest_generate_implementation.md)	 - Generate new Implementation manifests
* [capact manifest generate interface](capact_manifest_generate_interface.md)	 - Generate new Interface-related manifests
* [capact manifest generate interfacegroup](capact_manifest_generate_interfacegroup.md)	 - Generate new InterfaceGroup manifest
* [capact manifest generate type](capact_manifest_generate_type.md)	 - Generate new Type manifests

