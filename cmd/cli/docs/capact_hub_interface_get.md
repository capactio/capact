---
title: capact hub interface get
---

## capact hub interface get

Displays one or multiple Interfaces available on the Hub server

```
capact hub interface get [flags]
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

* [capact hub interface](capact_hub_interface.md)	 - This command consists of multiple subcommands to interact with Interfaces stored on the Hub server

