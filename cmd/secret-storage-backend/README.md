# Secret Storage Backend

## Overview

Secret Storage Backend is a service which handles multiple secret storages for TypeInstances.

## Prerequisites

- [Go](https://golang.org)
- (Optional - if AWS Secrets Manager provider should be used) an AWS account with **AdministratorAccess** permissions on it

## Usage

### AWS Secrets Manager provider

By default, the Secret Storage Backend has the `aws_secretsmanager` provider enabled.

1. Create AWS security credentials with `SecretsManagerReadWrite` policy.
2. Export environment variables:

   ```bash
    export AWS_ACCESS_KEY_ID="{accessKey}"
    export AWS_SECRET_ACCESS_KEY="{secretKey}"
    ```
3. Run the server:

    ```bash
    APP_LOGGER_DEV_MODE=true go run ./cmd/secret-storage-backend/main.go
    ```

The server listens to gRPC calls according to the [Storage Backend Protocol Buffers schema](../../hub-js/proto/storage_backend.proto).
To perform such calls, you can use e.g. [Insomnia](https://insomnia.rest/) tool.

### Dotenv provider

To run the server with `dotenv` provider enabled, which stores data in files, execute:

   ```bash
   APP_SUPPORTED_PROVIDERS=dotenv,aws_secretsmanager APP_LOGGER_DEV_MODE=true go run ./cmd/secret-storage-backend/main.go
   ```

> **NOTE:** You can enable multiple providers, separating them by comma, such as: `APP_SUPPORTED_PROVIDERS=aws_secretsmanager,dotenv`.

## Configuration

| Name                    | Required | Default              | Description                                                                                                                   |
|-------------------------|----------|----------------------|-------------------------------------------------------------------------------------------------------------------------------|
| APP_GRPC_ADDR           | no       | `:50051`             | TCP address the gRPC server binds to.                                                                                         |
| APP_HEALTHZ_ADDR        | no       | `:8082`              | TCP address the health probes endpoint binds to.                                                                              |
| APP_SUPPORTED_PROVIDERS | no       | `aws_secretsmanager` | Supported secret providers separated by `,`. A given provider must be passed in additional parameters of gRPC request inputs. |
| APP_LOGGER_DEV_MODE     | no       | `false`              | Enable development mode logging.                                                                                              |

To configure providers, use environmental variables described in the [Providers](https://github.com/SpectralOps/teller#providers) paragraph for Teller's Readme.

## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
