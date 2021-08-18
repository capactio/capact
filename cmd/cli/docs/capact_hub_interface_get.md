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
  -o, --output string      Output format. One of: json | table | yaml (default "table")
      --timeout duration   Timeout for HTTP request (default 30s)
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact hub interface](capact_hub_interface.md)	 - This command consists of multiple subcommands to interact with Interfaces stored on the Hub server

