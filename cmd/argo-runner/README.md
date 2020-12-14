# Argo runner

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Argo runner is a [runner](../../docs/runner.md), which executes Argo workflows. It is used as a built-in runner for Voltron Kubernetes implementation.

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
cat <<EOF > /tmp/argo-runner-args.yml
context:
  name: argo-runner-job
  dryRun: false
  platform:
    namespace: default
    ownerRef:
      apiVersion: batch/v1
      kind: Job
      name: argo-runner-owner
      uid: $(kubectl get jobs argo-runner-owner -ojsonpath='{.metadata.uid}')
args:
  workflow:
    entrypoint: main
    templates:
      - name: main
        container:
          image: docker/whalesay
          command: ["/bin/bash", "-c"]
          args: ["sleep 2 && cowsay hello world"]
EOF
```

2. Start the runner type:
```bash
RUNNER_INPUT_PATH=/tmp/argo-runner-args.yml go run cmd/argo-runner/main.go
```

You can check the workflow status in Argo UI on http://localhost:2746 after setting port-forwarding:
```bash
kubectl port-forward -n argo svc/argo-server 2746
```

## Configuration

The following environment variables can be set:

| Name                                | Required | Default            | Description                           |
| ----------------------------------- | -------- | ------------------ | ------------------------------------- |
| RUNNER_INPUT_PATH                   | yes      |                    | Path to the runner YAML input file    |
| RUNNER_LOGGER_DEV_MODE              | no       | `false`            | Enable additional log messages        |
| RUNNER_GCP_SERVICE_ACCOUNT_FILEPATH | no       | `/etc/gcp/sa.json` | Path to the GCP JSON credentials file |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
