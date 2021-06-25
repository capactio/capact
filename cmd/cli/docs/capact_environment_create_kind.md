---
title: capact environment create kind
---

## capact environment create kind

Provision local kind cluster

```
capact environment create kind [flags]
```

### Options

```
      --cluster-config string   path to a kind config file
  -h, --help                    help for kind
      --image string            node docker image to use for booting the cluster (default "kindest/node:v1.19.1")
      --kubeconfig string       sets kubeconfig path instead of $KUBECONFIG or $HOME/.kube/config
      --name string             cluster name, overrides config (default "kind-dev-capact")
      --retain                  retain nodes for debugging when cluster creation fails
      --wait duration           wait for control plane node to be ready
```

### Options inherited from parent commands

```
  -c, --config string   Path to the YAML config file
```

### SEE ALSO

* [capact environment create](capact_environment_create.md)	 - This command consists of multiple subcommands to create a Capact environment

