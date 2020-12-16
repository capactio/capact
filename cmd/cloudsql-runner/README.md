# CloudSQL runner

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

CloudSQL runner is a [runner](../../docs/runner.md), which creates and manages CloudSQL instances and databases on Google Cloud Platform.

## Prerequisites

- [Go](https://golang.org)
- Google Cloud Platform credentials JSON for Service Account with `Cloud SQL Admin` role

To get the GCP Service Account you can follow the documentation [here](https://cloud.google.com/iam/docs/creating-managing-service-accounts#creating). Then generate the JSON key and download it, to get the credentials JSON.

## Usage

To start the runner type:
```bash
RUNNER_GCP_SERVICE_ACCOUNT_FILEPATH={path-to-gcp-service-account-credentials-json} \
  RUNNER_INPUT_PATH=cmd/cloudsql-runner/example-input.yml \
  RUNNER_LOGGER_DEV_MODE=true \
  go run cmd/cloudsql-runner/main.go
```

## Configuration

The following environment variables can be set:

| Name                                | Required | Default            | Description                           |
| ----------------------------------- | -------- | ------------------ | ------------------------------------- |
| RUNNER_INPUT_PATH                   | yes      |                    | Path to the runner YAML input file    |
| RUNNER_LOGGER_DEV_MODE              | no       | `false`            | Enable additional log messages        |
| RUNNER_GCP_SERVICE_ACCOUNT_FILEPATH | no       | `/etc/gcp/sa.json` | Path to the GCP JSON credentials file |
| KUBECONFIG                          | no       | `~/.kube/config`   | Path to kubeconfig file               |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
