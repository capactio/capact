# Voltron Open Capability Hub

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
  - [Local OCH](#local-och)
  - [Public OCH](#public-och)
  - [GraphQL Playground](#graphql-playground)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Open Capability Hub (OCH) is a component, which stores the OCF manifests and exposes API to access and manage them. It can work in two modes:
- Local OCH - in this mode it exposes GraphQL API for managing TypeInstances (create, read, delete  operations),
- Public OCH - in this mode it exposes read-only GraphQL API for querying all OCF manifests except TypeInstances.

The OCHs are accessed via a GraphQL API. OCH uses [Neo4j](https://neo4j.com) database.

## Prerequisites

- [Go](https://golang.org)
- Running Kubernetes cluster with Voltron installed

## Usage

### Local OCH

To run OCH in local mode, use the following command:
```bash
APP_OCH_MODE=local go run cmd/och/main.go
```

### Public OCH

To run OCH in local mode, use the following command:
```bash
APP_OCH_MODE=public go run cmd/och/main.go
```

### GraphQL Playground

You can now access the OCH's GraphQL Playground via http://localhost:8080/. For example to list all Interfaces on the public OCH make the following GraphQL query:
```graphql
query {
  interfaces {
    prefix,
    name
  }
}
```

## Configuration

The following environment variables can be set to configure OCH:

| Name                | Required | Default | Description                                            |
| ------------------- | -------- | ------- | ------------------------------------------------------ |
| APP_OCH_MODE        | yes      |         | Mode, in which OCH is run. Must be "public" or "local" |
| APP_GRAPHQL_ADDR    | no       | `:8080` | TCP address the GraphQL endpoint binds to              |
| APP_HEALTHZ_ADDR    | no       | `:8082` | TCP address the health probes endpoint binds to        |
| APP_LOGGER_DEV_MODE | no       | `false` | Enable development mode logging                        |
| APP_MOCK_GRAPH_QL   | no       | `false` | Use mocked data in GraphQL server                      |

## Development

#### Common development guides

To read more about development, see the [`development.md`](../../docs/development.md) document.
