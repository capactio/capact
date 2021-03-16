## ocftool hub implementations search

Search provides the ability to search for OCH Interfaces

```
ocftool hub implementations search [flags]
```

### Examples

```
# Show all implementations in table format
ocftool hub implementations search cap.interface.database.postgresql.install

# Show all implementations in YAML format			
ocftool hub implementations search cap.interface.database.postgresql.install -oyaml

```

### Options

```
  -h, --help                        help for search
      --interface-revision string   Specific interface revision
  -o, --output string               Output format. One of:
                                    json|yaml|table (default "table")
```

### SEE ALSO

* [ocftool hub implementations](ocftool_hub_implementations.md)	 - This command consists of multiple subcommands to interact with Hub server.

