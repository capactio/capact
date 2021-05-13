## capact hub implementations get

Lists the currently available Implementations on the Hub server

```
capact hub implementations get [flags]
```

### Examples

```
# Show all implementations in table format
capact hub implementations get cap.interface.database.postgresql.install

# Show all implementations in YAML format			
capact hub implementations get cap.interface.database.postgresql.install -oyaml

```

### Options

```
  -h, --help                        help for get
      --interface-revision string   Specific interface revision
  -o, --output string               Output format. One of:
                                    json | yaml | table (default "table")
```

### SEE ALSO

* [capact hub implementations](capact_hub_implementations.md)	 - This command consists of multiple subcommands to interact with Implementations stored on the Hub server

