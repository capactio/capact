# Terraform runner

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Terraform runner is a [runner](../../docs/runner.md), which downloads and run Terraform modules. Runner is a wrapper for
the terraform binary. It downloads specified module, runs terraform init and depending on action: apply, destroy or plan.
After run, it collects the output and converts it into Voltron required format.

## Prerequisites

- [Go](https://golang.org)
- [Terraform](https://www.terraform.io/downloads.html)

- Google Cloud Platform credentials JSON for Service Account with `Cloud SQL Admin` role
  This is not the Terraform Runner requirement but in example CloudSQL instance is created.

To get the GCP Service Account you can follow the documentation [here](https://cloud.google.com/iam/docs/creating-managing-service-accounts#creating). Then generate the JSON key and download it to get the credentials JSON.

## Usage

To start the runner type:
```bash
GOOGLE_APPLICATION_CREDENTIALS={full-path-to-gcp-service-account-credentials-json} \
  RUNNER_ARGS_PATH=cmd/terraform-runner/example-args.yml \
  RUNNER_CONTEXT_PATH=cmd/terraform-runner/example-context.yml \
  RUNNER_LOGGER_DEV_MODE=true \
  RUNNER_WORKDIR=/tmp/workspace \
  go run cmd/terraform-runner/main.go
```

## Configuration

The following environment variables can be set:

| Name                                       | Required | Default                       | Description                                                        |
|--------------------------------------------|----------|-------------------------------|--------------------------------------------------------------------|
| RUNNER_CONTEXT_PATH                        | yes      |                               | Path to the YAML file with runner context                          |
| RUNNER_ARGS_PATH                           | yes      |                               | Path to the YAML file with input arguments                         |
| RUNNER_LOGGER_DEV_MODE                     | no       | `false`                       | Enable additional log messages                                     |
| RUNNER_OUTPUT_TERRAFORM_RELEASE_FILE_PATH  | no       | `/tmp/terraform-release.yaml` | Defines path under which the Terraform artifacts is saved          |
| RUNNER_OUTPUT_ADDITIONAL_FILE_PATH         | no       | `/tmp/additional.yaml`        | Defines path under which the additional output is saved            |
| RUNNER_OUTPUT_TFSTATE_FILE_PATH            | no       | `/tmp/terraform.tfstate`      | Defines path under which the terraform.tfstate output is saved     |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
