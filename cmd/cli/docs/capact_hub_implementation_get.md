---
title: capact hub implementation get
---

## capact hub implementation get

Displays one or multiple Implementations available on the Hub server

```
capact hub implementation get [flags]
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
  -h, --help               help for get
  -o, --output string      Output format. One of: json | jsonpath | table | yaml (default "table")
  -t, --template string    JSON path output template
      --timeout duration   Timeout for HTTP request (default 30s)
```

### Options inherited from parent commands

```
  -C, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact hub implementation](capact_hub_implementation.md)	 - This command consists of multiple subcommands to interact with Implementations stored on the Hub server

