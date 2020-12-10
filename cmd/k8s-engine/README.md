# Voltron Engine

## Overview

Voltron Engine is a component responsible for handling Action custom resources. It implements the Kubernetes controller pattern.

## Prerequisites

- Running Kubernetes cluster with Voltron installed
- Go compiler 1.14+
- (optional) [Telepresence](https://www.telepresence.io/)

## Usage

See [this document](../../docs/development.md#replace-a-cluster-component-with-your-local-process) for how to setup a Telepresence session to the Voltron Engine deployment on your development Kubernetes cluster.

After you have the Telepresence session created, you can run the Engine in the Telepresence shell:
```bash
go run cmd/k8s-engine/main.go
```

You can now access the Engine's GraphQL API via https://gateway.voltron.local/graphql. For example to list all actions make the following GraphQL query:
```graphql
query {
  actions {
    name
  }
}
```

## Configuration

| Name                          | Default                          | Description                                                                                                  |
|-------------------------------|----------------------------------|--------------------------------------------------------------------------------------------------------------|
| APP_ENABLE_LEADER_ELECTION    | false                            | Enable leader election for Kubernetes controller. This ensures only 1 controller is active at any time point |
| APP_LEADER_ELECTION_NAMESPACE |                                  | Set the Kubernetes namespace, in which the leader election ConfigMap is created                              |
| APP_GRAPHQL_ADDR              | `:8080`                          | TCP address the metrics endpoint binds to                                                                    |
| APP_GRAPHQL_ADDR              | `8081`                           | TCP address the metrics endpoint binds to                                                                    |
| APP_HEALTHZ_ADDR              | `:8082`                          | TCP address the health probes endpoint binds to                                                              |
| APP_LOGGER_DEV_MODE           | `false`                          | Enable development mode logging                                                                              |
| APP_MAX_CONCURRENT_RECONCILES | `1`                              | Maximum number of concurrent reconcile loops in the controller                                               |
| APP_MOCK_GRAPHQL              | `false`                          | Set mock responses on the GraphQL server                                                                     |
| APP_GRAPHQL_GATEWAY_ENDPOINT  | `http://voltron-gateway/graphql` | Endpoint of the Voltron Gateway                                                                              |
| APP_GRAPHQL_GATEWAY_USERNAME  |                                  | Basic auth username used to authenticate at the Voltron Gateway                                              |
| APP_GRAPHQL_GATEWAY_PASSWORD  |                                  | Basic auth password used to authenticate at the Voltron Gateway                                              |
| APP_BUILTIN_RUNNER_TIMEOUT    | `30m`                            | Set the timeout for the workflow execution of the builtin runners                                            |
| APP_BUILTIN_RUNNER_IMAGE      |                                  | Set the image of the builtin runner                                                                          |
