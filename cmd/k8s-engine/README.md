# Voltron Engine

## Overview

Voltron Engine is a component responsible for handling Action custom resources. It implements the Kubernetes controller pattern

## Prerequisites

- Running Kubernetes cluster
- Go compiler 1.14+
- (optional) [Telepresence](https://www.telepresence.io/)

## Usage

As the Voltron Engine is an integral part of Voltron is it hard to run it without the whole Voltron deployment. For development and testing you can either:
1. [Build new image and deploy on local KinD cluster](#build-new-image-and-deploy-on-local-kind-cluster)
2. [Use telepresence](#use-telepresence)

### Build new image and deploy on local KinD cluster

Running the following command will build new component images (including the Voltron Engine) and deploy to the local KinD cluster
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
