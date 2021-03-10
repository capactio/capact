## ocftool och list interfaces

List OCH Interfaces

```
ocftool och list interfaces [flags]
```

### Examples

```
#  List all interfaces in table format
ocftool och list interfaces

# Print path for the first entry in returned response 
ocftool och list interfaces -o=jsonpath="{.interfaces[0]['path']}"

# Print paths
ocftool och list interfaces -o=jsonpath="{range .interfaces[*]}{.path}{'\n'}{end}"

# Start interactive mode
ocftool och list interfaces -i

```

### Options

```
  -h, --help                 help for interfaces
  -i, --interactive          Start interactive mode
  -o, --output string        Output format. One of:
                             json|yaml|table|jsonpath=... (default "table")
      --path-prefix string   Pattern of the path of a given Interface, e.g. cap.interface.* (default "cap.interface.*")
```

### SEE ALSO

* [ocftool och list](ocftool_och_list.md)	 - This command consists of multiple subcommands to interact with OCH server.

