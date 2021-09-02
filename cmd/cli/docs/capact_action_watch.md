---
title: capact action watch
---

## capact action watch

Watch an Action until it has completed execution

### Synopsis


    Watch an Action until it has completed execution

    NOTE:   An action needs to be created and run in order to run this command.
            This command calls the Kubernetes API directly. As a result, KUBECONFIG has to be configured
            with the same cluster as the one which the Gateway points to.
    

```
capact action watch ACTION [flags]
```

### Examples

```
# Watch an Action:
capact action watch ACTION

# Watch the Action which was created last:
capact action watch @latest

```

### Options

```
  -h, --help                         help for watch
  -n, --namespace string             If present, the namespace scope for this CLI request
      --node-field-selector string   selector of node to display, eg: --node-field-selector phase=abc
      --status string                Filter by status (Pending, Running, Succeeded, Skipped, Failed, Error)
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

