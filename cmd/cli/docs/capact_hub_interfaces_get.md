## capact hub interfaces get

Get provides the ability to list and search for OCH Interfaces

```
capact hub interfaces get [flags]
```

### Examples

```
# Show all interfaces in table format:
capact hub interfaces get

# Show "cap.interface.database.postgresql.install" interface in JSON format:
capact hub interfaces get -o json cap.interface.database.postgresql.install

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

