---
title: capact upgrade
---

## capact upgrade

Upgrades Capact

### Synopsis

Use this command to upgrade the Capact version on a cluster.

```
capact upgrade [flags]
```

### Examples

```
# Upgrade Capact components to the newest available version
capact upgrade

# Upgrade Capact components to 0.1.0 version
capact upgrade --version 0.1.0
```

### Options

```
      --action-name-prefix string           Specifies Capact upgrade Action name prefix. (default "capact-upgrade-")
      --crd string                          Capact Action CRD location. (default "https://raw.githubusercontent.com/capactio/capact/main/deploy/kubernetes/crds/core.capact.io_actions.yaml")
      --enable-test-setup                   Enables test setup for the Capact E2E validation scenarios.
      --helm-repo string                    Capact Helm chart repository URL. Use @latest tag to select repository which holds the latest Helm chart versions. (default "https://storage.googleapis.com/capactio-stable-charts")
  -h, --help                                help for upgrade
      --increase-resource-limits            Enables higher resource requests and limits for components. (default true)
      --max-queue-time duration             Maximum waiting time for the completion of other, currently running upgrade tasks. Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 10m0s)
      --override-capact-image-repo string   Allows you to override Docker image repository for Capact components. By default, Docker image repository from Helm chart is used.
      --override-capact-image-tag string    Allows you to override Docker image tag for Capact components. By default, Docker image tag from Helm chart is used.
      --print-insecure-helm-release-notes   Prints the base64-encoded Gateway password directly in Helm release notes.
      --timeout duration                    Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 10m0s)
      --version string                      Capact version. (default "@latest")
  -w, --wait                                Waits for the upgrade process until it's finished or the defined "--timeout" has occurred. (default true)
```

### Options inherited from parent commands

```
  -c, --config string                 Path to the YAML config file
  -v, --verbose int/string[=simple]   Prints more verbose output. Allowed values: 0 - disable, 1 - simple, 2 - trace (default 0 - disable)
```

### SEE ALSO

* [capact](capact.md)	 - Collective Capability Manager CLI

