---
title: capact action wait
---

## capact action wait

Wait for a specific condition of a given Action

```
capact action wait ACTION [flags]
```

### Examples

```
# Wait for the Action "example" to contain the phase "READY_TO_RUN"
capact act wait --for=phase=READY_TO_RUN example

```

### Options

```
      --for string              The field condition to wait on. Currently, only the 'phase' field is supported: 'phase={phase-name}'.
  -h, --help                    help for wait
  -n, --namespace string        Kubernetes namespace where the Action was created. (default "default")
      --timeout duration        Timeout for HTTP request (default 30s)
      --wait-timeout duration   Maximum time to wait before giving up. "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 10m0s)
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

