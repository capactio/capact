# Voltron GraphQL gateway

## Overview

Voltron GraphQL gateway is a component, which aggregates GraphQL API from the Voltron Engine and Open Capability Hub.

## Prerequisites

- Running Kubernetes cluster with Voltron installed
- Go compiler 1.14+
- [Telepresence](https://www.telepresence.io/)

## Usage

See [this document](../../docs/development.md#replace-a-cluster-component-with-your-local-process) for how to setup a Telepresence session to the Gateway deployment on your development Kubernetes cluster.

After you have the Telepresence session created, you can run the gateway in the Telepresence shell:
```bash
go run cmd/gateway/main.go
```

### Access GraphQL playground

You can access the GraphQL playground on the Gateway by opening http://localhost:8080. As currently the gateway is secured using basic auth you need to provide the following headers:
```json
{
  "Authorization": "Basic Z3JhcGhxbDp0MHBfczNjcjN0"
}
```

Then you should be able to make queries to the gateway:
```graphql
query($implementationPath: NodePath!) {
  implementation(path: $implementationPath) {
    name,
    prefix,
    latestRevision {
      spec {
        action {
          runnerInterface
          args
        }
      }
    }
  }
}
```

## Configuration

You can set the following environment variables to configure the Gateway:

| Name                                | Default   | Description                                                                                                                                                           |
|-------------------------------------|-----------|-----------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| APP_GRAPHQL_ADDR                    | `:8080`   | TCP address the GraphQL endpoint binds to                                                                                                                             |
| APP_HEALTHZ_ADDR                    | `:8082`   | TCP address the health probes endpoint binds to                                                                                                                       |
| APP_LOGGER_DEV_MODE                 | `false`   | Enable development mode logging                                                                                                                                       |
| APP_INTROSPECTION_GRAPHQL_ENDPOINTS | `false`   | Comma separated list of GraphQL endpoint to introspect and merge into one unified GraphQL endpoint. Ex. `http://localhost:3000/graphql,http://localhost:3001/graphql` |
| APP_INTROSPECTION_ATTEMPTS          | `120`     | Number of attempts to introspect the remote GraphQL endpoints                                                                                                         |
| APP_INTROSPECTION_RETRY_DELAY       | `1s`      | Time delay between unsuccessful introspection attempts                                                                                                                |
| APP_AUTH_USERNAME                   | `graphql` | Basic auth username used to secure the GraphQL endpoint                                                                                                               |
| APP_AUTH_PASSWORD                   |           | Basic auth password used to secure the GraphQL endpoint                                                                                                               |