## ocftool hub interfaces search

Search provides the ability to list and search for OCH Interfaces

```
ocftool hub interfaces search [flags]
```

### Examples

```
# Show all interfaces in table format
ocftool hub interfaces search

# Show all interfaces in JSON format which are located under the "cap.interface.templating" prefix 
ocftool hub interfaces search -o json --path-pattern "cap.interface.*"

```

### Options

```
  -h, --help                  help for search
  -o, --output string         Output format. One of:
                              json | yaml | table (default "table")
      --path-pattern string   Pattern of the path for a given Interface, e.g. cap.interface.* (default "cap.interface.*")
```

### SEE ALSO

* [ocftool hub interfaces](ocftool_hub_interfaces.md)	 - This command consists of multiple subcommands to interact with Interfaces stored on the Hub server

