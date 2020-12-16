# Voltron Engine

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Voltron Engine is a component responsible for handling Action custom resources. It implements the Kubernetes controller pattern and exposes an GraphQL server for operations on Voltron Actions.

## Prerequisites

- [Go](https://golang.org)
- Running Kubernetes cluster with Voltron installed

## Usage

In order to run Engine, configure an access to a Kubernetes cluster. By default it is loaded from default location `.kube/config` in current user's home directory. To provide a different path, see the [Configuration](#configuration) section.

Another requirement is to provide a Voltron Gateway URL, so the Engine can fetch OCF TypeInstances, Interfaces and Implementations. You can use the Gateway running on local kind cluster, which is accessible under `https://gateway.voltron.local/graphql`.

To run the Engine, use:
```bash
APP_GRAPHQLGATEWAY_ENDPOINT=https://gateway.voltron.local/graphql \
  APP_GRAPHQLGATEWAY_USERNAME=graphql \
  APP_GRAPHQLGATEWAY_PASSWORD=t0p_s3cr3t \
  APP_BUILTIN_RUNNER_IMAGE='local/argo-runner:dev' \
  go run cmd/k8s-engine/main.go
```

You can now access the Engine's GraphQL API via http://localhost:8080/. For example to list all actions make the following GraphQL query:
```graphql
query {
  actions {
    name
  }
}
```

## Configuration

| Name                          | Required | Default                          | Description                                                                                                  |
| ----------------------------- | -------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| APP_ENABLE_LEADER_ELECTION    | no       | false                            | Enable leader election for Kubernetes controller. This ensures only 1 controller is active at any time point |
| APP_LEADER_ELECTION_NAMESPACE | no       |                                  | Set the Kubernetes namespace, in which the leader election ConfigMap is created                              |
| APP_GRAPHQL_ADDR              | no       | `:8080`                          | TCP address the metrics endpoint binds to                                                                    |
| APP_GRAPHQL_ADDR              | no       | `8081`                           | TCP address the metrics endpoint binds to                                                                    |
| APP_HEALTHZ_ADDR              | no       | `:8082`                          | TCP address the health probes endpoint binds to                                                              |
| APP_LOGGER_DEV_MODE           | no       | `false`                          | Enable development mode logging                                                                              |
| APP_MAX_CONCURRENT_RECONCILES | no       | `1`                              | Maximum number of concurrent reconcile loops in the controller                                               |
| APP_MOCK_GRAPHQL              | no       | `false`                          | Set mock responses on the GraphQL server                                                                     |
| APP_GRAPHQLGATEWAY_ENDPOINT   | no       | `http://voltron-gateway/graphql` | Endpoint of the Voltron Gateway                                                                              |
| APP_GRAPHQLGATEWAY_USERNAME   | yes      |                                  | Basic auth username used to authenticate at the Voltron Gateway                                              |
| APP_GRAPHQLGATEWAY_PASSWORD   | yes      |                                  | Basic auth password used to authenticate at the Voltron Gateway                                              |
| APP_BUILTIN_RUNNER_TIMEOUT    | no       | `30m`                            | Set the timeout for the workflow execution of the builtin runners                                            |
| APP_BUILTIN_RUNNER_IMAGE      | yes      |                                  | Set the image of the builtin runner                                                                          |
| KUBECONFIG                    | no       | `~/.kube/config`                 | Path to kubeconfig file                                                                                      |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
