---
title: capact install
---

## capact install

install Capact into a given environment

### Synopsis

Use this command to install the Capact version in the environment.

```
capact install [OPTIONS] [flags]
```

### Examples

```
# Install latest Capact version from main branch
capact install

# Install Capact 0.1.0 version
capact install --version 0.1.0

# Install Capact from local git repository. Needs to be run from the main directory
capact install --version @local
```

### Options

```
      --build-image strings                    Local images names that should not be build when using @local version. Takes comma-separated list. (default [argo-actions,argo-runner,e2e-test,gateway,hub-js,k8s-engine,populator])
      --capact-overrides strings               Overrides for Capact component.
      --cert-manager-overrides strings         Overrides for Cert Manager component.
      --environment string                     Capact environment. (default "kind")
      --helm-repo-url string                   Capact Helm chart repository URL. Use @latest tag to select repository which holds the latest Helm chart versions. (default "https://storage.googleapis.com/capactio-stable-charts")
  -h, --help                                   help for install
      --increase-resource-limits               Enables higher resource requests and limits for components. (default true)
      --ingress-controller-overrides strings   Overrides for Ingress controller component.
      --install-component strings              Components names that should not be installed. Takes comma-separated list. (default [neo4j,ingress-controller,argo,cert-manager,kubed,monitoring,capact])
      --name string                            Cluster name, overrides config. (default "kind-dev-capact")
      --namespace string                       Capact namespace. (default "capact-system")
      --timeout duration                       Maximum time during which the upgrade process is being watched, where "0" means "infinite". Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h". (default 10m0s)
      --update-hosts-file                      Updates /etc/hosts with entry for Capact GraphQL Gateway. (default true)
      --update-trusted-certs                   Add Capact GraphQL Gateway certificate. (default true)
      --verbose                                Prints more verbose output.
      --version string                         Capact version. Possible values @latest, @local, 0.3.0, ... (default "@latest")
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact](capact.md)	 - Collective Capability Manager CLI

