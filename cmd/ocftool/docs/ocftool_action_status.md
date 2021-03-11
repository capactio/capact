## ocftool action status

Show Action status

```
ocftool action status ACTION [flags]
```

### Examples

```
# Get information about a workflow:
ocftool action status my-action

# Get the latest workflow:
ocftool action status @latest

```

### Options

```
  -h, --help                         help for status
  -n, --namespace string             If present, the namespace scope for this CLI request
      --no-color                     Disable colorized output
      --no-utf8                      Use plain 7-bits ascii characters
      --node-field-selector string   selector of node to display, eg: --node-field-selector phase=abc
  -o, --output string                Output format. One of: json|yaml|wide
      --status string                Filter by status (Pending, Running, Succeeded, Skipped, Failed, Error)
```

### SEE ALSO

* [ocftool action](ocftool_action.md)	 - This command consists of multiple subcommands to interact with Action.

