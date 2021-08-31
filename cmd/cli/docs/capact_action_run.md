---
title: capact action run
---

## capact action run

Queues up a specified Action for processing by the workflow engine

```
capact action run ACTION [flags]
```

### Options

```
  -h, --help               help for run
  -n, --namespace string   Kubernetes namespace where the Action was created (default "default")
      --timeout duration   Timeout for HTTP request (default 30s)
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

