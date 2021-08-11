---
title: capact manifest validate
---

## capact manifest validate

Validate OCF manifests

```
capact manifest validate [flags]
```

### Examples

```
# Validate interface-group.yaml file with OCF specification in default location
capact manifest validate ocf-spec/0.0.1/examples/interface-group.yaml

# Validate multiple files inside test_manifests directory with additional server-side checks
capact manifest validate --server-side pkg/cli/test_manifests/*.yaml

# Validate all Hub manifests with additional server-side checks
capact manifest validate --server-side ./manifests/**/*.yaml

# Validate interface-group.yaml file with custom OCF specification location 
capact manifest validate -s my/ocf/spec/directory ocf-spec/0.0.1/examples/interface-group.yaml
```

### Options

```
      --concurrency int   Maximum number of concurrent workers. (default 5)
  -h, --help              help for validate
  -s, --schemas string    Path to the local directory with OCF JSONSchemas. If not provided, built-in JSONSchemas are used.
      --server-side       Executes additional manifests checks against Capact Hub.
  -v, --verbose           Prints more verbose output.
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact manifest](capact_manifest.md)	 - This command consists of multiple subcommands to interact with OCF manifests

