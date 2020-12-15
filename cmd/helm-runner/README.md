# Helm runner

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Helm runner is a [runner](../../docs/runner.md), which creates and manages Helm releases on Kubernetes.

## Prerequisites

- [Go](https://golang.org)
- [Helm 3](https://helm.sh/docs/intro/install/)
- Running Kubernetes cluster

## Usage

To start the runner execute:
```bash
RUNNER_INPUT_PATH=cmd/helm-runner/example-input.yml RUNNER_LOGGER_DEV_MODE=true go run cmd/helm-runner/main.go
```

You can check, if the PostgreSQL Helm release was created:
```bash
$ helm list
NAME                    NAMESPACE       REVISION        UPDATED                                 STATUS       CHART                    APP VERSION
postgresql-1607608471   default         1               2020-12-10 14:54:34.882358554 +0100 CET deployed     postgresql-10.1.3        11.10.0 
```

## Configuration

The following environment variables can be set:

| Name                         | Required | Default          | Description                                    |
| ---------------------------- | -------- | ---------------- | ---------------------------------------------- |
| RUNNER_INPUT_PATH            | yes      |                  | Path to the runner YAML input file             |
| RUNNER_LOGGER_DEV_MODE       | no       | `false`          | Enable additional log messages                 |
| RUNNER_HELM_DRIVER           | no       | `secrets`        | Set Helm backend storage driver                |
| RUNNER_REPOSITORY_CACHE_PATH | no       | `/tmp/helm`      | Set the path to the repository cache directory |
| KUBECONFIG                   | no       | `~/.kube/config` | Path to kubeconfig file                        |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
