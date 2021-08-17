---
title: capact action create
---

## capact action create

Creates/renders a new Action with a specified Interface

```
capact action create INTERFACE [flags]
```

### Options

```
      --action-policy-from-file string    Path to the one-time Action policy file in YAML format
      --dry-run                           Specifies whether the Action performs server-side test without actually running the Action
  -h, --help                              help for create
  -i, --interactive                       Toggle interactive prompting in the terminal
      --name string                       The Action name. By default, a random name is generated.
  -n, --namespace string                  Kubernetes namespace where the Action is to be created
      --parameters-from-file string       Path to the Action input parameters file in YAML format
      --timeout duration                  Timeout for HTTP request (default 30s)
      --type-instances-from-file string   Path to the Action input TypeInstances file in YAML format. Example:
                                          typeInstances:
                                            - name: "config"
                                              id: "ABCD-1234-EFGH-4567"
      --validate                          Validate created Action before sending it to server (default true)
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

