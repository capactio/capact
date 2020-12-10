# Voltron Open Capability Hub

# Overview

Voltron Open Capablity Hub (OCH) is a component, which stores the OCF manifests and TypeInstances. I can work in two modes:
- local mode - it stores TypeInstances for a Voltron deployment
- public mode - works as a public repository, which provides OCF manifests to local OCHs

## Prerequisites

- Running Kubernetes cluster with Voltron installed
- Go compiler 1.14+
- (optional) [Telepresence](https://www.telepresence.io/)

## Usage

As the Voltron OCH is an integral part of Voltron is it hard to run it without the whole Voltron deployment. For development you can either:
1. [Build new image and deploy on local KinD cluster](#build-new-image-and-deploy-on-local-kind-cluster)
2. [Use telepresence](#use-telepresence)

### Build new image and deploy on local KinD cluster

Running the following command will build new component images (including the Voltron OCH) and deploy to the local KinD cluster
```bash
make dev-cluster-update
```

### Use telepresence

[Telepresence](https://www.telepresence.io/) is a tool to make it easier to develop applications, which are running on Kubernetes.

You can use the feature to replace a pod running on the cluster with a pod, which forward all traffic directed to this pod to your PC. In this way, you can run the process on your PC, like it would be in this pod.

```bash
# this will replace the pod with a telepresence proxy and open a new shell in your terminal
telepresence --swap-deployment voltron-och-local
# or
telepresence --swap-deployment voltron-och-public

# run the engine
go run cmd/och/main.go
```
