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

> **CAUTION:**  In order to run DB populator manually, make sure the populator inside development cluster is disabled.
> To disable it, run :`ENABLE_POPULATOR=false make dev-cluster-update`

It requires one argument, which is a path to directory with `och-content` directory. Internally it uses
[go-getter](https://github.com/hashicorp/go-getter) so it can download manifests from different locations
and in different formats.

To build the binary run:

```shell
go build -ldflags "-s -w" -o populator ./cmd/populator/main.go
```

To be able to use it locally when Voltron is running in a Kubernetes cluster, two ports need to
be forwarded:

```shell
kubectl -n neo4j port-forward svc/neo4j-neo4j 7687:7687
kubectl -n neo4j port-forward svc/neo4j-neo4j 7474:7474
```


It will create a `populator` binary in a local dir.

To run it and use local manifests from Voltron repo:

```shell
./populator .
```

To use manifests from private git repo, private key, encoded in base64 format, is needed.
For example command to download manifests from Voltron repo would look like this:
```shell
expoort SSHKEY=`base64 -w0 ~/.ssh/id_rsa`
./populator git@github.com:Project-Voltron/go-voltron.git?sshkey=$SSHKEY
```

For better performance populator starts HTTP server to serve manifests converted to JSON files.
Neo4j needs access to this JSON files. `APP_JSON_PUBLISH_ADDR` environment variable should be set
so populator can send a correct link to a Neo4j:

```shell
APP_JSON_PUBLISH_ADDR=http://{HOST_IP} ./populator .
```
Replace `HOST_IP` with your computer IP

## Configuration

You can set the following environment variables to configure the Gateway:

| Name                                | Required | Default   | Description                                                                                                                                                           |
| ----------------------------------- | -------- | --------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| APP_NEO4J_ADDR                       | no       | `neo4j://localhost:7687` | Neo4j address                                                                                                                                         |
| APP_NEO4J_USER                       | no       | `neo4j`                  | Neo4j admin user                                                                                                                                      |
| APP_NEO4J_PASSWORD                   | yes      |                          | Neo4h admin password                                                                                                                                  |
| APP_JSON_PUBLISH_ADDR                | yes      |                          | Address on which populator will serve JSON files                                                                                                      |
| APP_JSON_PUBLISH_PORT                | no       | `8080`                   | Port number on which populator will be listening                                                                                                      |
| APP_MANIFESTS_PATH                   | no       | `och-content`            | Path to a directory in a repository where manifests are stored                                                                                        |
| APP_REFRESH_WHEN_HASH_CHANGES        | no       | `false`                  | Flag to make populator populate data only when there are new changes in a repository                                                                  |
| APP_LOGGER_DEV_MODE                  | no       | `false`                  | Enable development mode logging                                                                                                                       |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
