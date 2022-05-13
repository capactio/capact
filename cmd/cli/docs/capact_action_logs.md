---
title: capact action logs
---

## capact action logs

Print the Action's logs

### Synopsis

Print the Action's logs

NOTE:   An action needs to be created and run in order to run this command.
        This command calls the Kubernetes API directly. As a result, KUBECONFIG has to be configured
        with the same cluster as the one which the Gateway points to.

```
capact action logs ACTION [POD] [flags]
```

### Examples

```
# Print the logs of an Action:
capact logs example

# Follow the logs of an Action:
capact logs example --follow

# Print the logs of single container in a pod
capact logs example step-pod -c step-pod-container

# Print the logs of an Action's step:
capact logs example step-pod

# Print the logs of the latest executed Action:
capact logs @latest

```

### Options

```
  -c, --container string    Print the logs of this container (default "main")
  -f, --follow              Specify if the logs should be streamed.
      --grep string         grep for lines
  -h, --help                help for logs
  -n, --namespace string    If present, the namespace scope for this CLI request
      --no-color            Disable colorized output
  -p, --previous            Specify if the previously terminated container logs should be returned.
      --since duration      Only return logs newer than a relative duration like 5s, 2m, or 3h. Defaults to all logs. Only one of since-time / since may be used.
      --since-time string   Only return logs after a specific date (RFC3339). Defaults to all logs. Only one of since-time / since may be used.
      --tail int            If set, the number of lines from the end of the logs to show. If not specified, logs are shown from the creation of the container or sinceSeconds or sinceTime (default -1)
      --timestamps          Include timestamps on each line in the log output
```

### Options inherited from parent commands

```
  -C, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

