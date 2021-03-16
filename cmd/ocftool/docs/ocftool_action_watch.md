## ocftool action watch

Watch an Action until it has completed execution

### Synopsis


    Watch an Action until it has completed execution

    NOTE:   An action needs to be created and run in order to run this command.
            Furthermore, 'kubectl' has to be configured with the context and default
            namespace set to be the same as the one which the Gateway points to. 
    

```
ocftool action watch ACTION [flags]
```

### Examples

```
# Watch an Action:
ocftool action watch ACTION

# Watch the Action which was created last:
ocftool action watch @latest

```

### Options

```
  -h, --help                         help for watch
  -n, --namespace string             If present, the namespace scope for this CLI request
      --node-field-selector string   selector of node to display, eg: --node-field-selector phase=abc
      --status string                Filter by status (Pending, Running, Succeeded, Skipped, Failed, Error)
```

### SEE ALSO

* [ocftool action](ocftool_action.md)	 - This command consists of multiple subcommands to interact with target Actions

