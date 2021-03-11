## ocftool hub interfaces search

Search provides the ability to search for OCH Interfaces

```
ocftool hub interfaces search [flags]
```

### Examples

```
#  Show all interfaces in table format
ocftool hub interfaces search

# Print path for the first entry in returned response 
ocftool hub interfaces search -oyaml

```

### Options

```
  -h, --help                  help for search
  -o, --output string         Output format. One of:
                              json|yaml|table (default "table")
      --path-pattern string   Pattern of the path of a given Interface, e.g. cap.interface.* (default "cap.interface.*")
```

### SEE ALSO

* [ocftool hub interfaces](ocftool_hub_interfaces.md)	 - This command consists of multiple subcommands to interact with Hub server.
