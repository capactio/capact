# Capact Hub

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
  - [Local Hub](#local-hub)
  - [Public Hub](#public-hub)
  - [GraphQL Playground](#graphql-playground)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Hub is a component, which stores the OCF manifests and exposes API to access, and manage them. It can work in two modes:

- Local Hub - in this mode it exposes GraphQL API for managing TypeInstances (create, read, delete operations),
- Public Hub - in this mode it exposes read-only GraphQL API for querying all OCF manifests except TypeInstances.

The Hubs are accessed via a GraphQL API.

## Prerequisites

- [Node v16](https://nodejs.org/)
- A running Neo4j database with APOC plugin

For the Neo4j database, you can run it locally using Docker:

```bash
docker run -d \
  -p 7687:7687 -p 7474:7474 \
  -e "NEO4J_AUTH=neo4j/okon" \
  -e "NEO4JLABS_PLUGINS=[\"apoc\"]" \
  --name hub-neo4j-instance \
  ghcr.io/capactio/neo4j:4.2.13-apoc
```

When you are done, remove the Docker container:

```bash
docker rm -f hub-neo4j-instance
```

## Usage

Download the NPM dependencies using:

```bash
npm install
```

### Local Hub

To run Hub in the local mode, use the following command:

```bash
APP_LOGGER_LEVEL="debug" APP_NEO4J_ENDPOINT=bolt://localhost:7687 APP_NEO4J_PASSWORD=okon APP_HUB_MODE=local npm run dev
```

### Public Hub

To run Hub in the public mode, use the following command:

```bash
APP_LOGGER_LEVEL="debug" APP_NEO4J_ENDPOINT=bolt://localhost:7687 APP_NEO4J_PASSWORD=okon APP_HUB_MODE=public npm run dev
```

### GraphQL Playground

Once you ran Hub locally, you can access the Hub GraphQL Playground under [http://localhost:8080/graphql](http://localhost:3000/graphql).

For example, to list all Interfaces on the public Hub make the following GraphQL query:

```graphql
query {
  Interface {
    prefix
    name
  }
}
```

## Configuration

The following environment variables can be set to configure Hub:

| Name                        | Required | Default                 | Description                                                                                                                                           |
| --------------------------- | -------- | ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| APP_HUB_MODE                | no       | `public`                | Mode, in which Hub is run. Must be "public" or "local"                                                                                                |
| APP_LOGGER_LEVEL            | no       | `info`                  | Specifies which log entries should be logged. Allowed values are described in [winston](https://github.com/winstonjs/winston#logging-levels) library. |
| APP_GRAPH_QL_ADDR           | no       | `:8080`                 | The address, where GraphQL endpoints binds to                                                                                                         |
| APP_NEO4J_ENDPOINT          | no       | `bolt://localhost:7687` | The Neo4j database Bolt protocol endpoint                                                                                                             |
| APP_NEO4J_USERNAME          | no       | `neo4j`                 | Neo4j database username                                                                                                                               |
| APP_NEO4J_PASSWORD          | yes      |                         | Neo4j database password                                                                                                                               |
| APP_EXPRESS_BODY_SIZE_LIMIT | no       | `32mb`                  | The limit of the maximum HTTP request body size                                                                                                       |

## Development

### Accessing Neo4j Browser

To access Neo4j Browser, follow the steps:

1. Run the following commands:

   - Capact cluster:
     ```bash
     kubectl -n capact-system port-forward svc/neo4j-neo4j 7474:7474
     kubectl -n capact-system port-forward svc/neo4j-neo4j 7687:7687
     ```
   - Docker:
     Make sure that `-p 7687:7687 -p 7474:7474` flags were specified for the `docker run` command.

2. Navigate to [http://localhost:7474](http://localhost:7474).
3. Change the connection URL to `neo4j://localhost:7687`.
4. Use `neo4j` configured user and password:
   - Capact cluster:
     See the default values in [`values.yaml`](../deploy/kubernetes/charts/neo4j/values.yaml) file.
   - Docker:
     Use values set by the `NEO4J_AUTH` environment variable.

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
