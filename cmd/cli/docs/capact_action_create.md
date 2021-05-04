## capact action create

Creates/renders a new Action with a specified Interface

```
capact action create INTERFACE [flags]
```

### Options

```
      --dry-run                           Specifies whether the Action performs server-side test without actually running the Action
  -h, --help                              help for create
  -i, --interactive                       Toggle interactive prompting in the terminal
      --name string                       The Action name. By default, a random name is generated.
  -n, --namespace string                  Kubernetes namespace where the Action is to be created
      --parameters-from-file string       The Action input parameters in YAML format
      --type-instances-from-file string   The Action input TypeInstances in YAML format. Example:
                                          typeInstances:
                                            - name: "config"
                                              id: "ABCD-1234-EFGH-4567"
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

