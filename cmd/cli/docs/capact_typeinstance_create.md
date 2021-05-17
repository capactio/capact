## capact typeinstance create

Creates a new TypeInstance(s)

```
capact typeinstance create [flags]
```

### Examples

```
# Create TypeInstances defined in a given file
capact typeinstance create -f ./tmp/typeinstances.yaml

```

### Options

```
  -f, --from-file strings   The TypeInstances input in YAML format (can specify multiple)
  -h, --help                help for create
  -o, --output string       Output format. One of: json | table | yaml (default "table")
```

### SEE ALSO

* [capact typeinstance](capact_typeinstance.md)	 - This command consists of multiple subcommands to interact with target TypeInstances

