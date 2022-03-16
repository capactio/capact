# Helm Storage Backend

## Overview

Helm Storage Backend is a service which handles Helm-related storage logic. It works in two modes. Each of them exposes the same Storage Backend gRPC server, but with a different set of features.

## Prerequisites

- [Go](https://golang.org)

## Usage

There are two separate modes of running the Helm storage backend.

### Helm Release storage backend

This mode exposes functionality which fetches metadata for a given Helm release from a Kubernetes cluster.

To run the server, execute:

 ```bash
 APP_LOGGER_DEV_MODE=true APP_MODE="release" go run ./cmd/helm-storage-backend/main.go
 ```

The server listens to gRPC calls according to the [Storage Backend Protocol Buffers schema](../../hub-js/proto/storage_backend.proto). To perform such calls, you can use e.g. [Insomnia](https://insomnia.rest/) tool.

### Helm Templating storage backend

This mode exposes functionality which renders a given Go template against an installed Helm release.

To run the server, execute:

 ```bash
 APP_LOGGER_DEV_MODE=true APP_MODE="template" go run ./cmd/helm-storage-backend/main.go
 ```

The server listens to gRPC calls according to the [Storage Backend Protocol Buffers schema](../../hub-js/proto/storage_backend.proto). To perform such calls, you can use e.g. [Insomnia](https://insomnia.rest/) tool.

## Configuration

| Name                | Required | Default          | Description                                        |
|---------------------|----------|------------------|----------------------------------------------------|
| APP_MODE            | yes      |                  | One of the service modes: `release` or `template`. |
| KUBECONFIG          | no       | `~/.kube/config` | Path to kubeconfig file                            |
| APP_GRPC_ADDR       | no       | `:50051`         | TCP address the gRPC server binds to.              |
| APP_HEALTHZ_ADDR    | no       | `:8082`          | TCP address the health probes endpoint binds to.   |
| APP_LOGGER_DEV_MODE | no       | `false`          | Enable development mode logging.                   |

## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
