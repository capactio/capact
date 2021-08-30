---
title: capact config set-context
---

## capact config set-context

Updates the active hub configuration context

```
capact config set-context [flags]
```

### Examples

```
# Selects which Hub/Gateway server to use of via a prompt
capact config set-context

# Sets the specified Hub/Gateway server
capact config set-context localhost:8080

```

### Options

```
  -h, --help   help for set-context
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact config](capact_config.md)	 - Manage configuration

