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

Open Capability Hub (OCH) is a component, which stores the OCF manifests and exposes API to access, and manage them. It can work in two modes:
- Local OCH - in this mode it exposes GraphQL API for managing TypeInstances (create, read, delete  operations),
- Public OCH - in this mode it exposes read-only GraphQL API for querying all OCF manifests except TypeInstances.

The OCHs are accessed via a GraphQL API.

## Prerequisites

- [Node v15](https://nodejs.org/)
- A running Neo4j database with APOC plugin

For the Neo4j database, you can run it locally using Docker:
```
docker run -d \
  -p 7687:7687 -p 7474:7474 \
  -e "NEO4J_AUTH=neo4j/okon" \
  -e "NEO4JLABS_PLUGINS=[\"apoc\"]" \
  neo4j:4.1.3
```

## Usage

### Local OCH

To run OCH in local mode, use the following command:
```bash
APP_NEO4J_ENDPOINT=bolt://localhost:7687 \
  APP_NEO4J_PASSWORD=okon \
  APP_OCH_MODE=local \
  npm run dev
```

### Public OCH

To run OCH in local mode, use the following command:
```bash
APP_NEO4J_ENDPOINT=bolt://localhost:7687 \
  APP_NEO4J_PASSWORD=okon \
  APP_OCH_MODE=public \
  npm run dev
```

### GraphQL Playground

You can now access the OCH's GraphQL Playground via http://localhost:3000/graphql. For example, to list all Interfaces on the public OCH make the following GraphQL query:
```graphql
query {
  Interface {
    prefix,
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

To read more about development, see the [`development.md`](../../docs/development.md) document.
