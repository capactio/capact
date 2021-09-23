# Argo runner

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Argo runner is a [runner](https://capact.io/docs/architecture/runner), which executes Argo workflows. It is used as a built-in runner for Capact Kubernetes implementation.

## Prerequisites

- [Go](https://golang.org)
- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- running Kubernetes cluster with Argo installed

Before starting the runner you need to:
1. Ensure the Service Account used by the Argo workflow has proper RBAC permissions. Here is an example command to add permissions for using default service account in default namespace:
```bash
kubectl create clusterrolebinding default-default-admin --clusterrole admin --serviceaccount default:default
```
2. Create a job to use for ownerReference for the workflow and a secret for status reporting:
```bash
kubectl apply -f cmd/argo-runner/setup.yml
```
## Usage

1. Create the runner input YAML file:
```bash
cat <<EOF > /tmp/argo-runner-context.yaml
name: argo-runner-job
dryRun: false
platform:
  namespace: default
  ownerRef:
    apiVersion: batch/v1
    kind: Job
    name: argo-runner-owner
    uid: $(kubectl get jobs argo-runner-owner -ojsonpath='{.metadata.uid}')
EOF
```

2. Start the runner type:
```bash
RUNNER_CONTEXT_PATH=/tmp/argo-runner-context.yaml RUNNER_ARGS_PATH=cmd/argo-runner/example-args.yaml RUNNER_LOGGER_DEV_MODE=true go run cmd/argo-runner/main.go
```

You can check the workflow status in Argo UI on [http://localhost:2746](http://localhost:2746) after setting port-forwarding:
```bash
kubectl port-forward -n capact-system svc/argo-server 2746
```

## Configuration

The following environment variables can be set:

| Name                   | Required | Default          | Description                                |
| ---------------------- | -------- | ---------------- | ------------------------------------------ |
| RUNNER_CONTEXT_PATH    | yes      |                  | Path to the YAML file with runner context  |
| RUNNER_ARGS_PATH       | yes      |                  | Path to the YAML file with input arguments |
| RUNNER_LOGGER_DEV_MODE | no       | `false`          | Enable additional log messages             |
| KUBECONFIG             | no       | `~/.kube/config` | Path to kubeconfig file                    |

## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
