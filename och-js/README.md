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
- [yarn](https://yarnpkg.com/)
- A running Neo4j database with APOC plugin

You can install `yarn` with NPM using:
```
npm install -g yarn
```

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
NEO4J_ENDPOINT=bolt://localhost:7687 \
  NEO4J_PASSWORD=okon \
  OCH_MODE=public \
  yarn dev
```

### Public OCH

To run OCH in local mode, use the following command:
```bash
NEO4J_ENDPOINT=bolt://localhost:7687 \
  NEO4J_PASSWORD=okon \
  OCH_MODE=local \
  yarn dev
```

### GraphQL Playground

You can now access the OCH's GraphQL Playground via http://localhost:3000/graphql. For example to list all Interfaces on the public OCH make the following GraphQL query:
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

| Name              | Required | Default | Description                                            |
| ----------------- | -------- | ------- | ------------------------------------------------------ |
| OCH_MODE          | yes      |         | Mode, in which OCH is run. Must be "public" or "local" |
| GRAPHQL_BIND_PORT | no       | `3000`  | TCP port the GraphQL endpoint binds to                 |
| NEO4J_ENDPOINT    | yes      |         | The Neo4j database Bolt protocol endpoint              |
| NEO4J_USERNAME    | no       | `false` | Neo4j database username                                |
| NEO4J_PASSWORD    | yes      |         | Neo4j database password                                |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
