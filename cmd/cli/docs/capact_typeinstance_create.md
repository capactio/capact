---
title: capact typeinstance create
---

## capact typeinstance create

Creates a new TypeInstance(s)

### Synopsis

Create one or multiple TypeInstances from a given file.

Syntax:
	
	typeInstances:
	  - alias: parent # required when submitting more than one TypeInstance
	    attributes: # optional
	      - path: cap.attribute.cloud.provider.aws
	        revision: 0.1.0
	    typeRef: # required
	      path: cap.type.aws.auth.credentials
	      revision: 0.1.0
	    value: # required
	      accessKeyID: fake-123
	      secretAccessKey: fake-456
	
	usesRelations: # optional
	  - from: parent
	    to: 123-4313 # ID of already existing TypeInstance, or TypeInstance alias from a given request


NOTE: Supported syntax are YAML and JSON.


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

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact typeinstance](capact_typeinstance.md)	 - This command consists of multiple subcommands to interact with target TypeInstances

