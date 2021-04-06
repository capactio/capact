## ocftool upgrade

Upgrades Capact

### Synopsis

Use this command to upgrade the Capact version on a cluster.

```
ocftool upgrade [flags]
```

### Examples

```
# Upgrade Capact components to newest available version
ocftool upgrade

# Upgrade Capact components to 0.1.0 version
ocftool upgrade --version 0.1.0
```

### Options

```
  -h, --help                       help for upgrade
      --increase-resource-limits   Enables higher resource requests and limits for components. (default true)
      --timeout duration           Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 5m0s)
      --version string             Capact version (default "@latest")
  -w, --wait                       Waits for the upgrade process until it finish or the defined "--timeout" occurs.
```

### SEE ALSO

* [ocftool](ocftool.md)	 - Collective Capability Manager CLI

