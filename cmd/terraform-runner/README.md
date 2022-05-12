# Terraform runner

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Terraform runner is a [runner](https://capact.io/docs/architecture/runner), which downloads and run Terraform modules. Runner is a wrapper for
the terraform binary. It downloads specified module, runs terraform init and depending on action: apply, destroy or plan.
After run, it collects the output and converts it into Capact required format.

## Prerequisites

- [Go](https://golang.org)
- [Terraform](https://www.terraform.io/downloads.html)

- Google Cloud Platform credentials JSON for Service Account with `Cloud SQL Admin` role
  This is not the Terraform Runner requirement but in example CloudSQL instance is created.

To get the GCP Service Account you can follow the documentation [here](https://cloud.google.com/iam/docs/creating-managing-service-accounts#creating). Then generate the JSON key and download it to get the credentials JSON.

## Usage

To run any of the commands, execute:

```bash
# Export JSON credentials filepath
export GCP_CREDS_FILEPATH={full-path-to-gcp-service-account-credentials-json} 

# Tar example module
cd ./cmd/terraform-runner/example-input/ && tar -zcvf /tmp/cloudsql.tgz ./main.tf && cd -
```

### Apply

To run example which provisions CloudSQL for PostgreSQL on GCP, execute:
```bash
GOOGLE_APPLICATION_CREDENTIALS=$GCP_CREDS_FILEPATH \
  RUNNER_ARGS_PATH=cmd/terraform-runner/example-input/apply-args.yml \
  RUNNER_CONTEXT_PATH=cmd/terraform-runner/example-input/context.yml \
  RUNNER_LOGGER_DEV_MODE=true \
  RUNNER_WORKDIR=/tmp/tf-runner-workspace \
  go run cmd/terraform-runner/main.go
```

### Destroy

To clean up resources created from the [Apply](#apply) section, run:

```bash
# Backup TFState file as the one from workspace will be overriden by Terraform Runner
cp /tmp/tf-runner-workspace/terraform.tfstate /tmp/tf-runner-workspace/terraform.tfstate.bak 

# Copy TypeInstance file
cp /tmp/terraform.tfstate /tmp/cloudsqlti 

GOOGLE_APPLICATION_CREDENTIALS=$GCP_CREDS_FILEPATH \
  RUNNER_STATE_TYPE_INSTANCE_FILEPATH=/tmp/cloudsqlti \
  RUNNER_ARGS_PATH=cmd/terraform-runner/example-input/destroy-args.yml \
  RUNNER_CONTEXT_PATH=cmd/terraform-runner/example-input/context.yml \
  RUNNER_LOGGER_DEV_MODE=true \
  RUNNER_WORKDIR=/tmp/tf-runner-workspace \
  go run cmd/terraform-runner/main.go
```

## Configuration

The following environment variables can be set:

| Name                                       | Required | Default                       | Description                                                                                                           |
|--------------------------------------------|----------|-------------------------------|-----------------------------------------------------------------------------------------------------------------------|
| RUNNER_CONTEXT_PATH                        | yes      |                               | Path to the YAML file with runner context                                                                             |
| RUNNER_ARGS_PATH                           | yes      |                               | Path to the YAML file with input arguments                                                                            |
| RUNNER_LOGGER_DEV_MODE                     | no       | `false`                       | Enable additional log messages                                                                                        |
| RUNNER_OUTPUT_TERRAFORM_RELEASE_FILE_PATH  | no       | `/tmp/terraform-release.yaml` | Defines path under which the Terraform artifacts is saved                                                             |
| RUNNER_OUTPUT_ADDITIONAL_FILE_PATH         | no       | `/tmp/additional.yaml`        | Defines path under which the additional output is saved                                                               |
| RUNNER_OUTPUT_TFSTATE_FILE_PATH            | no       | `/tmp/terraform.tfstate`      | Defines path under which the terraform.tfstate output is saved                                                        |
| RUNNER_STATE_TYPE_INSTANCE_FILEPATH        | no       |                               | Defines path to the input state TypeInstance file. If not set, then the runner will run apply with an empty state file|

## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
