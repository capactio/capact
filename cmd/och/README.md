# Voltron Open Capability Hub

# Overview

Voltron Open Capability Hub (OCH) is a component, which stores the OCF manifests and TypeInstances. I can work in two modes:
- local mode - In this mode it stores TypeInstances for a Voltron deployment.
- public mode - In this mode it works as a public repository, which provides OCF manifests to local OCHs.

## Prerequisites

- Running Kubernetes cluster with Voltron installed
- Go compiler 1.14+
- [Telepresence](https://www.telepresence.io/)

## Usage

See [this document](../../docs/development.md#replace-a-cluster-component-with-your-local-process) for how to setup a Telepresence session to the OCH deployment on your development cluster.

After you have the Telepresence session created, you can run the OCH in the Telepresence shell:
```bash
go run cmd/k8s-engine/main.go
```

## Configuration

The following environment variables can be set to configure OCH:

| Name                | Default | Description                                            |
|---------------------|---------|--------------------------------------------------------|
| APP_OCH_MODE        |         | Mode, in which OCH is run. Must be "public" or "local" |
| APP_GRAPHQL_ADDR    | `:8080` | TCP address the GraphQL endpoint binds to              |
| APP_HEALTHZ_ADDR    | `:8082` | TCP address the health probes endpoint binds to        |
| APP_LOGGER_DEV_MODE | `false` | Enable development mode logging                        |
| APP_MOCK_GRAPHQL    | `false` | Use mocked data in GraphQL server                      |
