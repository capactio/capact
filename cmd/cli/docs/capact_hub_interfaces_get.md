## capact hub interfaces get

Get provides the ability to list and search for OCH Interfaces

```
capact hub interfaces get [flags]
```

### Examples

```
# Show all interfaces in table format
capact hub interfaces get

# Show all interfaces in JSON format which are located under the "cap.interface.templating" prefix 
capact hub interfaces get -o json --path-pattern "cap.interface.*"

```

### Options

```
  -h, --help                  help for get
  -o, --output string         Output format. One of:
                              json | yaml | table (default "table")
      --path-pattern string   Pattern of the path for a given Interface, e.g. cap.interface.* (default "cap.interface.*")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact hub interfaces](capact_hub_interfaces.md)	 - This command consists of multiple subcommands to interact with Interfaces stored on the Hub server

