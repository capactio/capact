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
  -h, --help              help for login
  -p, --password string   Password
  -u, --username string   Username
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact](capact.md)	 - Collective Capability Manager CLI

