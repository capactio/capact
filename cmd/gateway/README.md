# Capact GraphQL gateway

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
  - [Access GraphQL playground](#access-graphql-playground)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Capact GraphQL gateway is a component, which aggregates GraphQL APIs from the Capact Engine and Open Capability Hub.

## Prerequisites

- [Go](https://golang.org)
- Running Kubernetes cluster with Capact installed

## Usage

As Gateway aggregates multiple GraphQL endpoints for Capact components, an existing Capact installation is needed. You can use `kubectl port-forward` to setup port forwarding to GraphQL endpoints on the OCH and Capact Engine:
```
kubectl port-forward svc/capact-engine-graphql 3000:80 -n capact-system
kubectl port-forward svc/capact-och-public 3001:80 -n capact-system
kubectl port-forward svc/capact-och-local 3002:80 -n capact-system
```

To run the Gateway, execute:
```bash
APP_INTROSPECTION_GRAPH_QL_ENDPOINTS=http://localhost:3000/graphql,http://localhost:3001/graphql,http://localhost:3002/graphql \
  APP_AUTH_PASSWORD=t0p_s3cr3t \
  go run cmd/gateway/main.go
```

### Access GraphQL playground

You can access the GraphQL playground on the Gateway by opening [http://localhost:8080](http://localhost:8080). As the Gateway is secured using basic auth, you need to provide the following headers:
```json
{
  "Authorization": "Basic Z3JhcGhxbDp0MHBfczNjcjN0"
}
```

Then you should be able to make queries to the gateway:
```graphql
query {
  implementations {
    prefix,
    name
  }
}
```

## Configuration

You can set the following environment variables to configure the Gateway:

| Name                                | Required | Default   | Description                                                                                                                                                           |
| ----------------------------------- | -------- | --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| APP_GRAPHQL_ADDR                    | no       | `:8080`   | TCP address the GraphQL endpoint binds to                                                                                                                             |
| APP_HEALTHZ_ADDR                    | no       | `:8082`   | TCP address the health probes endpoint binds to                                                                                                                       |
| APP_LOGGER_DEV_MODE                 | no       | `false`   | Enable development mode logging                                                                                                                                       |
| APP_INTROSPECTION_GRAPHQL_ENDPOINTS | yes      |           | Comma separated list of GraphQL endpoint to introspect and merge into one unified GraphQL endpoint. Ex. `http://localhost:3000/graphql,http://localhost:3001/graphql` |
| APP_INTROSPECTION_ATTEMPTS          | no       | `120`     | Number of attempts to introspect the remote GraphQL endpoints                                                                                                         |
| APP_INTROSPECTION_RETRY_DELAY       | no       | `1s`      | Time delay between unsuccessful introspection attempts                                                                                                                |
| APP_AUTH_USERNAME                   | no       | `graphql` | Basic auth username used to secure the GraphQL endpoint                                                                                                               |
| APP_AUTH_PASSWORD                   | yes      |           | Basic auth password used to secure the GraphQL endpoint                                                                                                               |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
