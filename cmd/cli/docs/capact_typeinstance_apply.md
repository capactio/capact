---
title: capact typeinstance apply
---

## capact typeinstance apply

Apply a given TypeInstance(s)

### Synopsis

Updates a given TypeInstance(s).
CAUTION: Race updates may occur as TypeInstance locking is not used by CLI.


```
capact typeinstance apply -f file... [flags]
```

### Examples

```
# Apply TypeInstances from the given file.
capact typeinstance apply -f /tmp/typeinstances.yaml

```

### Options

```
  -f, --from-file strings   The TypeInstances input in YAML format (can specify multiple)
  -h, --help                help for apply
  -o, --output string       Output format. One of: json | table | yaml (default "table")
      --timeout duration    Timeout for HTTP request (default 30s)
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact typeinstance](capact_typeinstance.md)	 - This command consists of multiple subcommands to interact with target TypeInstances

