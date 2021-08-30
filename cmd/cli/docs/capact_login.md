---
title: capact login
---

## capact login

Login to a Hub (Gateway) server

```
capact login [OPTIONS] [SERVER] [flags]
```

### Examples

```
# start interactive setup
capact login

# Specify server name and specify the user
capact login localhost:8080 -u user

```

### Options

```
  -h, --help               help for login
  -p, --password string    Password
      --timeout duration   Timeout for HTTP request (default 30s)
  -u, --username string    Username
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact](capact.md)	 - Collective Capability Manager CLI

