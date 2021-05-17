## capact typeinstance update

Updates a given TypeInstance(s)

### Synopsis

Updates a given TypeInstance(s).
CAUTION: Race updates may occur as TypeInstance locking is not used by CLI.


```
capact typeinstance update [-f file]... | TYPE_INSTANCE_ID [flags]
```

### Examples

```
# Apply TypeInstances from the given file
capact typeinstance update -f /tmp/typeinstances.yaml 

# Update TypeInstance in editor mode 
capact typeinstance update TYPE_INSTANCE_ID

```

### Options

```
  -f, --from-file strings   The TypeInstances input in YAML format (can specify multiple)
  -h, --help                help for update
  -o, --output string       Output format. One of: json | table | yaml (default "table")
```

### SEE ALSO

* [capact typeinstance](capact_typeinstance.md)	 - This command consists of multiple subcommands to interact with target TypeInstances

