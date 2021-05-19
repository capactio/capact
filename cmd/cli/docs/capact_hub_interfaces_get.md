## capact hub interfaces get

Displays one or multiple Interfaces available on the Hub server

```
capact hub interfaces get [flags]
```

### Examples

```
# Show all Interfaces in table format:
capact hub interfaces get

# Show "cap.interface.database.postgresql.install" Interface in JSON format:
capact hub interfaces get cap.interface.database.postgresql.install -ojson

```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. One of:
                        json | yaml | table (default "table")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact hub interfaces](capact_hub_interfaces.md)	 - This command consists of multiple subcommands to interact with Interfaces stored on the Hub server

