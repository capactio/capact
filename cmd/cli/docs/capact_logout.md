---
title: capact logout
---

## capact logout

Logout from the Hub (Gateway) server

```
capact logout [SERVER] [flags]
```

### Examples

```
# Select what server to log out of via a prompt			
capact logout

# Logout of a specified Hub server
capact logout localhost:8080

```

### Options

```
  -h, --help   help for logout
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact](capact.md)	 - Collective Capability Manager CLI

