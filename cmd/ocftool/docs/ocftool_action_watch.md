## ocftool action watch

Watch an Action until it completes

```
ocftool action watch ACTION [flags]
```

### Examples

```
# Watch an Action:
ocftool action watch my-action

# Watch the latest Action:
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

* [ocftool action](ocftool_action.md)	 - This command consists of multiple subcommands to interact with Action.

