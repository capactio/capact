## capact policy apply

Updates current Policy with new value

```
capact policy apply -f {path} [flags]
```

### Examples

```
# Updates the Policy using content from file
capact policy apply -f /tmp/policy.yaml

```

### Options

```
  -f, --from-file string   The path to new Policy in YAML format
  -h, --help               help for apply
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact policy](capact_policy.md)	 - This command consists of multiple subcommands to interact with Policy

