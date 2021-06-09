---
title: capact validate
---

## capact validate

Validate OCF manifests

```
capact validate [flags]
```

### Examples

```
# Validate interface-group.yaml file with OCF specification in default location
capact validate ocf-spec/0.0.1/examples/interface-group.yaml

# Validate multiple files inside test_manifests directory
capact validate pkg/cli/test_manifests/*.yaml

# Validate interface-group.yaml file with custom OCF specification location 
capact validate -s my/ocf/spec/directory ocf-spec/0.0.1/examples/interface-group.yaml

# Validate all OCH manifests
capact validate ./och-content/**/*.yaml
```

### Options

```
  -h, --help             help for validate
  -s, --schemas string   Path to the local directory with OCF JSONSchemas. If not provided, built-in JSONSchemas are used.
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact](capact.md)	 - Collective Capability Manager CLI

