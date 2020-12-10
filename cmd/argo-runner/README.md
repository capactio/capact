# Argo runner

## Overview

Argo runner is a [Voltron workflow runner](../../docs/runner.md), which executes Argo workflows. It is used as the main Voltron workflow runner.

## Prerequisites

- running Kubernetes cluster with Argo installed
- kubectl
- Go compiler 1.14+

Before starting the runner you need to:
1. Ensure the Service Account used by the Argo workflow has proper RBAC permissions. Here is an example command to add permissions for using default service account in default namespace:
```bash
kubectl create clusterrolebinding default-default-admin --clusterrole admin --serviceaccount default:default
```
2. Create a job to use for ownerReference for the workflow and a secret for status reporting:
```bash
kubectl apply -f cmd/argo-runner/dummy-job.yml
```
## Usage

To start the runner type:
```bash
RUNNER_INPUT_PATH=cmd/argo-runner/example-input.yml go run cmd/argo-runner/main.go
```

You can check the workflow status in Argo UI on http://localhost:2746 after setup port-forwarding:
```bash
kubectl port-forward -n argo svc/argo-server 2746
```

## Configuration

The following environment variables can be set:

| Name                   | Default | Description                        |
|------------------------|---------|------------------------------------|
| RUNNER_INPUT_PATH      |         | Path of the runner YAML input file |
| RUNNER_LOGGER_DEV_MODE | `false` | Enable additional log messages     |
