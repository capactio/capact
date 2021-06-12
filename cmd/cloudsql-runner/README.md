# CloudSQL runner

> **NOTE**: This runner is deprecated in favor of [Terraform Runner](../terraform-runner).

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

CloudSQL runner is a [runner](https://capact.io/docs/architecture/runner), which creates and manages CloudSQL instances and databases on Google Cloud Platform.

## Prerequisites

- [Go](https://golang.org)
- Google Cloud Platform credentials JSON for Service Account with `Cloud SQL Admin` role

To get the GCP Service Account you can follow the documentation [here](https://cloud.google.com/iam/docs/creating-managing-service-accounts#creating). Then generate the JSON key and download it, to get the credentials JSON.

## Usage

To start the runner type:
```bash
RUNNER_GCP_SERVICE_ACCOUNT_FILEPATH={path-to-gcp-service-account-credentials-json} \
  RUNNER_CONTEXT_PATH=cmd/cloudsql-runner/example-context.yaml \
  RUNNER_ARGS_PATH=cmd/cloudsql-runner/example-args.yaml \
  RUNNER_LOGGER_DEV_MODE=true \
  go run cmd/cloudsql-runner/main.go
```

## Configuration

The following environment variables can be set:

| Name                                       | Required | Default                      | Description                                                           |
|--------------------------------------------|----------|------------------------------|-----------------------------------------------------------------------|
| RUNNER_CONTEXT_PATH                        | yes      |                              | Path to the YAML file with runner context                             |
| RUNNER_ARGS_PATH                           | yes      |                              | Path to the YAML file with input arguments                            |
| RUNNER_LOGGER_DEV_MODE                     | no       | `false`                      | Enable additional log messages                                        |
| RUNNER_GCP_SERVICE_ACCOUNT_FILEPATH        | no       | `/etc/gcp/sa.json`           | Path to the GCP Service Account credentials file                      |
| RUNNER_GCP_SERVICE_ACCOUNT_FORMAT          | no       | `json`                       | Format of the GCP Service Account credentials file - `yaml` or `json` |
| RUNNER_OUTPUT_CLOUD_SQL_INSTANCE_FILE_PATH | no       | `/tmp/cloudSQLInstance.yaml` | Defines path under which the Cloud SQL instance artifacts is saved    |
| RUNNER_OUTPUT_ADDITIONAL_FILE_PATH         | no       | `/tmp/additional.yaml`       | Defines path under which the additional output is saved               |
| KUBECONFIG                                 | no       | `~/.kube/config`             | Path to kubeconfig file                                               |

## Development

To read more about development, see the [Development guide](https://capact.io/docs/development/development-guide).
