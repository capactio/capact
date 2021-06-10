---
title: capact action delete
---

## capact action delete

Deletes the Action

```
capact action delete [ACTION_NAME...] [flags]
```

### Examples

```
# Deletes the foo Action in the default namespace
capact action delete foo

# Deletes all Actions with 'upgrade-' prefix in the foo namespace
capact action delete --name-regex='upgrade-*' --namespace=foo

```

### Options

```
  -h, --help                help for delete
      --name-regex string   Deletes all Actions whose names are matched by the given regular expression. To check the regex syntax, read: https://golang.org/s/re2syntax
  -n, --namespace string    Kubernetes namespace where the Action was created (default "default")
      --phase string        Deletes Actions only in the given phase. Supported only when the --name-regex flag is used. Allowed values: INITIAL, BEING_RENDERED, ADVANCED_MODE_RENDERING_ITERATION, READY_TO_RUN, RUNNING, BEING_CANCELED, CANCELED, SUCCEEDED, FAILED
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact action](capact_action.md)	 - This command consists of multiple subcommands to interact with target Actions

