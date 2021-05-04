## capact hub implementations search

Lists the currently available Implementations on the Hub server

```
capact hub implementations search [flags]
```

### Examples

```
# Show all implementations in table format
capact hub implementations search cap.interface.database.postgresql.install

# Show all implementations in YAML format			
capact hub implementations search cap.interface.database.postgresql.install -oyaml

```

### Options

```
  -h, --help                        help for search
      --interface-revision string   Specific interface revision
  -o, --output string               Output format. One of:
                                    json | yaml | table (default "table")
```

### SEE ALSO

* [capact hub implementations](capact_hub_implementations.md)	 - This command consists of multiple subcommands to interact with Implementations stored on the Hub server

