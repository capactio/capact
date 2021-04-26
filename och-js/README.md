# Capact Open Capability Hub

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
  - [Local OCH](#local-och)
  - [Public OCH](#public-och)
  - [GraphQL Playground](#graphql-playground)
- [Configuration](#configuration)
- [Development](#development)

## Overview

Open Capability Hub (OCH) is a component, which stores the OCF manifests and exposes API to access, and manage them. It can work in two modes:

- Local OCH - in this mode it exposes GraphQL API for managing TypeInstances (create, read, delete operations),
- Public OCH - in this mode it exposes read-only GraphQL API for querying all OCF manifests except TypeInstances.

The OCHs are accessed via a GraphQL API.

## Prerequisites

- [Node v15](https://nodejs.org/)
- A running Neo4j database with APOC plugin

For the Neo4j database, you can run it locally using Docker:

```bash
docker run -d \
  -p 7687:7687 -p 7474:7474 \
  -e "NEO4J_AUTH=neo4j/okon" \
  -e "NEO4JLABS_PLUGINS=[\"apoc\"]" \
  --name och-neo4j-instance \
  neo4j:4.2.3
```

When you are done, remove the Docker container:

```bash
docker rm -f och-neo4j-instance
```

## Usage

Download the NPM dependencies using:

```bash
npm install
```

### Local OCH

To run OCH in local mode, use the following command:

```bash
APP_NEO4J_ENDPOINT=bolt://localhost:7687 APP_NEO4J_PASSWORD=okon APP_OCH_MODE=local npm run dev
```

### Public OCH

To run OCH in public mode, use the following command:

```bash
APP_NEO4J_ENDPOINT=bolt://localhost:7687 APP_NEO4J_PASSWORD=okon APP_OCH_MODE=public npm run dev
```

### GraphQL Playground

Once you ran OCH locally, you can access the OCH GraphQL Playground under [http://localhost:3000/graphql](http://localhost:3000/graphql).

For example, to list all Interfaces on the public OCH make the following GraphQL query:

```graphql
query {
  Interface {
    prefix
    name
  }
}
```

## Configuration

The following environment variables can be set to configure OCH:

| Name               | Required | Default                 | Description                                            |
| ------------------ | -------- | ----------------------- | ------------------------------------------------------ |
| APP_OCH_MODE       | no       | `public`                | Mode, in which OCH is run. Must be "public" or "local" |
| APP_GRAPH_QL_ADDR  | no       | `:8080`                 | The address, where GraphQL endpoins binds to           |
| APP_NEO4J_ENDPOINT | no       | `bolt://localhost:7687` | The Neo4j database Bolt protocol endpoint              |
| APP_NEO4J_USERNAME | no       | `neo4j`                 | Neo4j database username                                |
| APP_NEO4J_PASSWORD | yes      |                         | Neo4j database password                                |

## Development

### Accessing Neo4j Browser

To access Neo4j Browser, follow the steps:

1. Run the following commands:

```bash
kubectl -n capact-system port-forward svc/neo4j-neo4j 7474:7474
kubectl -n capact-system port-forward svc/neo4j-neo4j 7687:7687
```

1. Navigate to [http://localhost:7474](http://localhost:7474).
1. Change the connection URL to `neo4j://localhost:7687`.
1. Use `neo4j` user and password configured during Helm chart installation. See the default values in [`values.yaml`](../../deploy/kubernetes/charts/neo4j/values.yaml) file.

To read more about development, see the [`development.md`](../../docs/development.md) document.
