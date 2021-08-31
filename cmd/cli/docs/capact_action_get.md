---
title: capact action get
---

## capact action get

Displays one or multiple Actions

```
capact action get [flags]
```

### Examples

```
# Show all Actions in table format
capact action get

# Show the Action "funny-stallman" in JSON format
capact action get funny-stallman -ojson

```

### Options

```
  -h, --help               help for get
  -n, --namespace string   Kubernetes namespace where the Action was created (default "default")
  -o, --output string      Output format. One of: json | table | yaml (default "table")
      --timeout duration   Timeout for HTTP request (default 30s)
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

