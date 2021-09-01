---
title: capact hub interface browse
---

## capact hub interface browse

Browse provides the ability to browse through the available OCF Interfaces in interactive mode. Optionally create a Target Action.

```
capact hub interface browse [flags]
```

### Examples

```
# Browse (and optionally create an Action) from the available OCF Interfaces.
<cli> hub interfaces browse

```

### Options

```
  -h, --help                  help for browse
      --path-pattern string   The pattern of the path of a given Interface, e.g. cap.interface.* (default "cap.interface.*")
      --timeout duration      Timeout for HTTP request (default 30s)
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - tracing (default 0 - disable)
```

### SEE ALSO

* [capact hub interface](capact_hub_interface.md)	 - This command consists of multiple subcommands to interact with Interfaces stored on the Hub server

