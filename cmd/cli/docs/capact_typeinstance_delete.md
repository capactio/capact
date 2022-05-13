---
title: capact typeinstance delete
---

## capact typeinstance delete

Delete a given TypeInstance(s)

```
capact typeinstance delete TYPE_INSTANCE_ID... [flags]
```

### Examples

```
# Delete TypeInstances with IDs 'c49b' and '4793'
capact typeinstance delete c49b 4793

```

### Options

```
  -h, --help               help for delete
      --timeout duration   Timeout for HTTP request (default 30s)
```

### Options inherited from parent commands

```
  -C, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact typeinstance](capact_typeinstance.md)	 - This command consists of multiple subcommands to interact with target TypeInstances

