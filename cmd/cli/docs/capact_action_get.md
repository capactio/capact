## capact action get

Lists the available Actions

```
capact action get [flags]
```

### Examples

```
# Show all Actions in table format
capact action get

# Show the Action "funny-stallman" in JSON format
capact action get funny-stallman -ojson

```

### Options

```
  -h, --help               help for get
  -n, --namespace string   Kubernetes namespace where the Action was created (default "default")
  -o, --output string      Output format. One of:
                           json | yaml | table (default "table")
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

