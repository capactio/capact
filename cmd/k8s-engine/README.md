# Voltron engine

Voltron engine is a component responsible for handling Action custom resources. It implements the Kubernetes controller pattern

## Supported features:

- CRUD operations on Action CRs
- Kubernetes controller for Action CR

## How to build

```bash
make build-app-image-k8s-engine
```

## How to use

As the Voltron engine is an integral part of Voltron is it hard to run it without the whole Voltron deployment. For development you can either:
1. [Build new image and deploy on local KinD cluster](#build-new-image-and-deploy-on-local-kind-cluster)
2. [Use telepresence](#use-telepresence)

### Build new image and deploy on local KinD cluster

Running the following command will build new component images (including the Voltron engine) and deploy to the local KinD cluster
```bash
make dev-cluster-update
```

### Use telepresence

[Telepresence](https://www.telepresence.io/) is a tool to make it easier to develop applications, which are running on Kubernetes.

You can use the feature to replace a pod running on the cluster with a pod, which forward all traffic directed to this pod to your PC. In this way, you can run the process on your PC, like it would be in this pod.

```bash
# this will replace the pod with a telepresence proxy and open a new shell in your terminal
telepresence --swap-deployment voltron-engine

# run the engine
go run cmd/k8s-engine/main.go
```

## Hacking

Main source code is in:
- `cmd/k8s-engine/` - binary main
- `pkg/engine/` - public source code
- `internal/k8s-engine/` - private source code
