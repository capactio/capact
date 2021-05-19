## capact hub implementations get

Displays one or multiple Implementations available on the Hub server

```
capact hub implementations get [flags]
```

### Examples

```
# Show all Implementation Revisions in table format
capact hub implementations get

# Show "cap.implementation.gcp.cloudsql.postgresql.install" Implementation Revisions in YAML format			
capact hub implementations get cap.interface.database.postgresql.install -oyaml

```

### Options

```
  -h, --help            help for get
  -o, --output string   Output format. One of: json | table | yaml (default "table")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact hub implementations](capact_hub_implementations.md)	 - This command consists of multiple subcommands to interact with Implementations stored on the Hub server

