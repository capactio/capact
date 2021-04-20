## ocftool upgrade

Upgrades Capact

### Synopsis

Use this command to upgrade the Capact version on a cluster.

```
ocftool upgrade [flags]
```

### Examples

```
# Upgrade Capact components to the newest available version
ocftool upgrade

# Upgrade Capact components to 0.1.0 version
ocftool upgrade --version 0.1.0
```

### Options

```
      --action-name-prefix string           Specifies Capact upgrade Action name prefix. (default "capact-upgrade-")
      --enable-test-setup                   Enables test setup for the Capact E2E validation scenarios.
      --helm-repo-url string                Capact Helm chart repository URL. Use @master tag to select repository which holds master Helm chart versions. (default "https://storage.googleapis.com/capactio-awesome-charts")
  -h, --help                                help for upgrade
      --increase-resource-limits            Enables higher resource requests and limits for components. (default true)
      --max-queue-time duration             Maximum waiting time for the completion of other, currently running upgrade tasks. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 10m0s)
      --override-capact-image-repo string   Allows you to override Docker image repository for Capact components. By default, Docker image repository from Helm chart is used.
      --override-capact-image-tag string    Allows you to override Docker image tag for Capact components. By default, Docker image tag from Helm chart is used.
      --timeout duration                    Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 10m0s)
      --version string                      Capact version. (default "@latest")
  -w, --wait                                Waits for the upgrade process until it finish or the defined "--timeout" occurs.
```

### SEE ALSO

* [ocftool](ocftool.md)	 - Collective Capability Manager CLI

