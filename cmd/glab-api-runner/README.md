# Helm runner

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

GitLab REST API runner is a [runner](https://capact.io/docs/architecture/runner), which executes the REST calls against any GitLab instance.

## Prerequisites

- [Go](https://golang.org)
- Access to GitLab instance

## Usage

### Installation

To start the runner `install` command, run:
```bash
RUNNER_CONTEXT_PATH=cmd/glab-api-runner/example-input/context.yaml \
 RUNNER_ARGS_PATH=cmd/glab-api-runner/example-input/create-project-args.yaml \
 RUNNER_LOGGER_DEV_MODE=true \
 go run cmd/glab-api-runner/main.go
```

To check if the PostgreSQL Helm release was created, run:
```bash
helm list 
```

## Configuration

The following environment variables can be set:

| Name                                 | Required | Default                  | Description                                                                    |
|--------------------------------------|----------|--------------------------|--------------------------------------------------------------------------------|
| RUNNER_CONTEXT_PATH                  | yes      |                          | Path to the YAML file with runner context                                      |
| RUNNER_ARGS_PATH                     | yes      |                          | Path to the YAML file with input arguments                                     |
| RUNNER_COMMAND                       | yes      |                          | Selected Helm Runner's command (currently supported: `install`, `upgrade`)     |
| RUNNER_HELM_RELEASE_PATH             | no       |                          | Path to the YAML file with Helm Release. Applicable only for `upgrade` command |
| RUNNER_LOGGER_DEV_MODE               | no       | `false`                  | Enable additional log messages                                                 |
| RUNNER_HELM_DRIVER                   | no       | `secrets`                | Set Helm backend storage driver                                                |
| RUNNER_REPOSITORY_CACHE_PATH         | no       | `/tmp/helm`              | Set the path to the repository cache directory                                 |
| RUNNER_OUTPUT_HELM_RELEASE_FILE_PATH | no       | `/tmp/helm-release.yaml` | Defines path under which the Helm release artifacts is saved                   |
| RUNNER_OUTPUT_ADDITIONAL_FILE_PATH   | no       | `/tmp/additional.yaml`   | Defines path under which the additional output is saved                        |
| RUNNER_KUBECONFIG                    | no       |                          | Path to kubeconfig file used by Runner, if not set the value of KUBECONFIG will be used |
| KUBECONFIG                           | no       | `~/.kube/config`         | Path to kubeconfig file                                                        |



## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
