---
title: capact policy apply
---

## capact policy apply

Updates current Policy with new value

```
capact policy apply -f {path} [flags]
```

### Examples

```
# Updates the Policy using content from file
capact policy apply -f /tmp/policy.yaml

```

### Options

```
  -f, --from-file string   The path to new Policy in YAML format
  -h, --help               help for apply
      --timeout duration   Timeout for HTTP request (default 30s)
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact policy](capact_policy.md)	 - This command consists of multiple subcommands to interact with Policy

