# Voltron DB populator

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
  - [Local OCH](#local-och)
  - [Public OCH](#public-och)
  - [GraphQL Playground](#graphql-playground)
- [Configuration](#configuration)
- [Development](#development)

## Overview 

DB populator is a component, which populates the OCF manifests into database. It reads manifest from remote or local path,
converts them into JSON and uploads to database.

## Prerequisites

- [Go](https://golang.org)
- Running Kubernetes cluster with Voltron installed

## Usage

DB populator requires few environment variables set:
 * APP_NEO4JADDR - is the TCP address the GraphQL endpoint binds to. Defualts to neo4j://localhost:7687
 * APP_NEO4JUSER - is the Neo4j admin user. Defaults to neo4j
 * APP_NEO4JPASSWORD - is the Neo4j admin password.
 * APP_JSONPUBLISHADDR -is the address on which populator will serve
   converted YAML files. It can be k8s service or for example local IP address

It requires one argument, which is a path to directory with `och-content` directory. Internally it uses
[go-getter](https://github.com/hashicorp/go-getter) so it can download manifests from different locations
and in different formats.

To run it locally and use manifests from Voltron repo:

```shell
./populator .
```

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
