## capact action status

Displays the details of an Action's status

```
capact action status ACTION [flags]
```

### Examples

```
# Get the status of a specified Action's workflow execution:
capact action status ACTION

# Gets the status from a last-run Action's workflow execution:
capact action status @latest

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

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

