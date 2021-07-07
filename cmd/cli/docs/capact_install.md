---
title: capact install
---

## capact install

install Capact into a given environment

```
capact install [OPTIONS] [SERVER] [flags]
```

### Options

```
      --enable-populator                    Enables Public Hub data populator (default true)
      --enable-test-setup                   Enables test setup for the Capact E2E validation scenarios.
      --environment string                  Capact environment. (default "kind")
      --focus-image strings                 Local images to build, all if not specified.
      --helm-repo-url string                Capact Helm chart repository URL. Use @latest tag to select repository which holds the latest Helm chart versions. (default "https://storage.googleapis.com/capactio-stable-charts")
  -h, --help                                help for install
      --increase-resource-limits            Enables higher resource requests and limits for components. (default true)
      --name string                         Cluster name, overrides config. (default "kind-dev-capact")
      --namespace string                    Capact namespace. (default "capact-system")
      --override-capact-image-repo string   Allows you to override Docker image repository for Capact components. By default, Docker image repository from Helm chart is used.
      --override-capact-image-tag string    Allows you to override Docker image tag for Capact components. By default, Docker image tag from Helm chart is used.
      --print-insecure-helm-release-notes   Prints the base64-encoded Gateway password directly in Helm release notes.
      --skip-component strings              Components names that should not be installed.
      --skip-image strings                  Local images names that should not be build when using local build.
      --timeout duration                    Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 10m0s)
      --version string                      Capact version. (default "@latest")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact](capact.md)	 - Collective Capability Manager CLI

