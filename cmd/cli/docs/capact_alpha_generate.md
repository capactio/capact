---
title: capact alpha generate
---

## capact alpha generate

OCF Manifests generation

### Synopsis

Subcommand for various manifest generation operations

```
capact alpha generate [flags]
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
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact alpha](capact_alpha.md)	 - Alpha features
* [capact alpha generate attribute](capact_alpha_generate_attribute.md)	 - Generate new Attribute manifests
* [capact alpha generate implementation](capact_alpha_generate_implementation.md)	 - Generate new Implementation manifests
* [capact alpha generate interface](capact_alpha_generate_interface.md)	 - Generate new Interface-related manifests
* [capact alpha generate interfacegroup](capact_alpha_generate_interfacegroup.md)	 - Generate new InterfaceGroup manifest
* [capact alpha generate type](capact_alpha_generate_type.md)	 - Generate new Type manifests

