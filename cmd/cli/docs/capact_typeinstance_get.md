## capact typeinstance get

Displays one or multiple TypeInstances

```
capact typeinstance get [TYPE_INSTANCE_ID...] [flags]
```

### Examples

```
# Display TypeInstances with IDs c49b and 4793
capact typeinstance get c49b 4793

# Save TypeInstances with IDs c49b and 4793 to file in the update format which later can be submitted for update by: 
# capact typeinstance apply --from-file /tmp/typeinstances.yaml
capact typeinstance get c49b 4793 -oyaml --export > /tmp/typeinstances.yaml

```

### Options

```
      --export          Converts TypeInstance to update format.
  -h, --help            help for get
  -o, --output string   Output format. One of: json | table | yaml (default "table")
```

### SEE ALSO

* [capact typeinstance](capact_typeinstance.md)	 - This command consists of multiple subcommands to interact with target TypeInstances

