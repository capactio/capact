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

It requires one argument, which is a path to directory with `och-content` directory. Internally it uses
[go-getter](https://github.com/hashicorp/go-getter) so it can download manifests from different locations
and in different formats.

To run it locally and use manifests from Voltron repo:

```shell
./populator .
```

## Configuration

You can set the following environment variables to configure the Gateway:

| Name                                | Required | Default   | Description                                                                                                                                                           |
| ----------------------------------- | -------- | --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| APP_NEO4JADDR                       | no       | `neo4j://localhost:7687` | Neo4j address                                                                                                                                          |
| APP_NEO4JUSER                       | no       | `neo4j`   | Neo4j admin user                                                                                                                                                      |
| APP_NEO4JPASSWORD                   | yes      |           | Neo4h admin password                                                                                                                                                  |
| APP_JSONPUBLISHADDR                 | yes      |           | Address on which populator will serve JSON files                                                                                                                      |
| APP_JSONPUBLISHPORT                 | no       | `8080`    | Port number on which populator will be listening                                                                                                                      |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
