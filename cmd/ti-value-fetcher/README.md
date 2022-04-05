# TypeInstance Value Fetcher

## Overview

TypeInstance Value Fetcher fetches the static value of a TypeInstance artifact based on context.
This helper app is useful inside a Capact Implementation workflow, where a given Argo artifact contains context, based on which the [ContextStorageBackend](../../hub-js/proto/storage_backend.proto) returns a static value.

However, the static value might be needed inside the same workflow, before uploading such artifact as a TypeInstance. This app calls external storage backend and executes `GetPreCreateValue` method to get such value.

## Prerequisites

- [Go](https://golang.org)
- Running Kubernetes cluster

## Usage

1. Run Kubernetes cluster
2. Install PostgreSQL chart according to the [Helm runner installation](../helm-runner/README.md#installation) section.
3. Run the Helm Release storage backend according to the [Helm Release usage](../helm-storage-backend/README.md#helm-release-storage-backend) instruction
4. Run the app:

   ```bash
   APP_LOGGER_DEV_MODE=true \
    APP_INPUT_TI_FILE_PATH="cmd/ti-value-fetcher/example-input/input-ti.yaml" \
    APP_INPUT_BACKEND_TI_FILE_PATH="cmd/ti-value-fetcher/example-input/storage-backend.yaml" \
    go run cmd/ti-value-fetcher/main.go
   ```

5. See the output:

   ```bash
   cat /tmp/output.yaml
   ```

## Configuration

| Name                           | Required | Default                  | Description                                                                                            |
|--------------------------------|----------|--------------------------|--------------------------------------------------------------------------------------------------------|
| APP_INPUT_TI_FILE_PATH         | no       | `/tmp/typeinstance.yaml` | The path to the file with TypeInstance artifact, which should have `value` property resolved.          |
| APP_INPUT_BACKEND_TI_FILE_PATH | no       | `/tmp/backend.yaml`      | The path to the Storage Backend TypeInstance artifact with connection details to the external service. |
| APP_OUTPUT_FILE_PATH           | no       | `/tmp/output.yaml`       | The path where the output file is saved.                                                               |
| APP_LOGGER_DEV_MODE            | no       | `false`                  | Enable development mode logging.                                                                       |

To configure providers, use environmental variables described in
the [Providers](https://github.com/SpectralOps/teller#providers) paragraph for Teller's Readme.

## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
