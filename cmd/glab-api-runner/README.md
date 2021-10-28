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

1. Update **baseURL** and **auth** properties in [`create-project-args.yaml`](./example-input/create-project-args.yaml).

2. Start the runner:

    ```bash
    RUNNER_CONTEXT_PATH=cmd/glab-api-runner/example-input/context.yaml \
     RUNNER_ARGS_PATH=cmd/glab-api-runner/example-input/create-project-args.yaml \
     RUNNER_LOGGER_DEV_MODE=true \
     go run cmd/glab-api-runner/main.go
    ```

3. Get connections details:

    ```bash
    cat /tmp/additional.yaml
    ```

## Configuration

The following environment variables can be set:

| Name                                 | Required | Default                  | Description                                                                    |
|--------------------------------------|----------|--------------------------|--------------------------------------------------------------------------------|
| RUNNER_CONTEXT_PATH                  | yes      |                          | Path to the YAML file with runner context                                      |
| RUNNER_ARGS_PATH                     | yes      |                          | Path to the YAML file with input arguments                                     |
| RUNNER_LOGGER_DEV_MODE               | no       | `false`                  | Enable additional log messages                                                 |
| RUNNER_OUTPUT_ADDITIONAL_FILE_PATH   | no       | `/tmp/additional.yaml`   | Defines path under which the additional output is saved                        |

## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
