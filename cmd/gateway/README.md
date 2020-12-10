# Voltron GraphQL gateway

## Overview

Voltron GraphQL gateway is a component, which aggregates GraphQL API from the Voltron Engine and Open Capability Hub

## Prerequisites

- Running Kubernetes cluster with Voltron installed
- Go compiler 1.14+
- (optional) [Telepresence](https://www.telepresence.io/)

## Usage

As the gateway is an integral part of Voltron is it hard to run it without the whole Voltron deployment. For development you can either:
1. [Build new image and deploy on local KinD cluster](#build-new-image-and-deploy-on-local-kind-cluster)
2. [Use telepresence](#use-telepresence)

### Build new image and deploy on local KinD cluster

To deploy the gateway to your dev cluster type:
```bash
make dev-cluster-update
```

This will build all apps (including gateway) and deploy them to the dev KinD cluster.

### Use telepresence

[Telepresence](https://www.telepresence.io/) is a tool to make it easier to develop applications, which are running on Kubernetes.

You can use the feature to replace a pod running on the cluster with a pod, which forward all traffic directed to this pod to your PC. In this way, you can run the process on your PC, like it would be in this pod.

```bash
# this will replace the pod with a telepresence proxy and open a new shell in your terminal
telepresence --swap-deployment voltron-gateway

# run the engine
go run cmd/k8s-engine/main.go
```

### Testing

During the local deployment an entry in `/etc/hosts` is added:
```properties
# /etc/hosts
[...]
127.0.0.1 gateway.voltron.local
```

You can access the GraphQL playground on the gateway by opening http://gateway.voltron.local/graphql. As currently the gateway is secured using basic auth you need to provide the following headers:
```json
{
  "Authorization": "Basic Z3JhcGhxbDp0MHBfczNjcjN0"
}
```

Then you should be able to make queries to the gateway:
```graphql
query($implementationPath: NodePath!) {
  implementation(path: $implementationPath) {
    name,
    prefix,
    latestRevision {
      spec {
        action {
          runnerInterface
          args
        }
      }
    }
  }
}
```
