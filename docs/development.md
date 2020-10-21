# Development

Read this document to learn how to develop the project. Please also follow guidelines defined in [this](development-guidelines.md) document.

## Table of Contents

<!-- toc -->

- [Prerequisites](#prerequisites)
- [Dependency management](#dependency-management)
- [Testing](#testing)
  * [Unit tests](#unit-tests)
  * [Lint tests](#lint-tests)
  * [Integration tests](#integration-tests)
- [Development cluster](#development-cluster)
  * [Create a cluster and install components](#create-a-cluster-and-install-components)
  * [Rebuild Docker images and update cluster](#rebuild-docker-images-and-update-cluster)
  * [Delete cluster](#delete-cluster)
- [Build and push Docker images](#build-and-push-docker-images)
  * [All components](#all-components)
  * [Single component](#single-component)
- [Generators](#generators)
  * [Generate Go code from OCF JSON Schemas](#generate-go-code-from-ocf-json-schemas)
  * [Generate K8s resources](#generate-k8s-resources)

<!-- tocstop -->

## Prerequisites

* [Go](https://golang.org/dl/) at least 1.15
* [Docker](https://www.docker.com/)
* Make

Helper scripts may introduce additional dependencies. However, all helper scripts support the `SKIP_DEPS_INSTALLATION` environment variable flag.
**By default, flag is set to `true`**, so all scripts try to use tools installed on your local machine as this speed up the process.
If you do not want to install any additional tools, or you want to ensure reproducible scripts 
results export `SKIP_DEPS_INSTALLATION=false`, so the proper tool version will be automatically installed and used. 

## Dependency management

This project uses `go modules` as a dependency manager. To install all required dependencies, use the following command:

```bash
go mod download
```

## Testing

### Unit tests

To run all unit tests, execute:

```bash
make test-unit
```

To generate the unit test coverage HTML report, execute: 

```bash
make test-cover-html
```

> **NOTE:** The default browser with the generated report opens automatically.

### Lint tests

To run lint tests, execute:

```bash
make test-lint
```

To automatically fix lint issues, execute:

```bash
make fix-lint-issues
```

### Integration tests

We support the cross-functional integration tests that are defined in [test](../test) package and 
Kubernetes controller integration tests which are using fake K8s API Server and `etcd`. 

#### Cross-functional

The cross-functional tests are executed on [`kind`](https://sigs.k8s.io//kind) where all Voltron components are pre-installed.

```bash
make test-integration
```

### K8s controller

The Kubernetes controller tests are executed on your local machine by starting a local control plane - `etcd` and `kube-apiserver`. 
For that purpose we use the [envtest](https://sigs.k8s.io/controller-runtime/pkg/envtest) library. 

```bash
make test-k8s-controller
```

## Development cluster 

To run development cluster, we use [`kind`](https://sigs.k8s.io//kind).

### Create a cluster and install components

To create the development cluster and install all components, execute:

```bash
make dev-cluster
```

### Rebuild Docker images and update cluster

To rebuild Docker images and upgrade Helm chart on dev cluster with new images, execute:

```bash
make dev-cluster-update
```

### Delete cluster

To delete the development cluster, execute: 

```bash
kind delete cluster --name kind-dev-voltron
```

## Build and push Docker images

There are a Make targets dedicated to build and push Voltron Docker images.

We have images for:
- application defined under [cmd](../cmd) directory
- tests defined under [test](../test) directory
- infra tools defined under [/hack/images](../hack/images) directory 

The default build and push configuration can be change via environment variables. For example:

```bash
export DOCKER_PUSH_REPOSITORY=gcr.io/projectvoltron/
export DOCKER_TAG=latest
```

### All components

To build all Docker images with your changes and push them to a registry, follow these steps.

1. Build all Docker images:
    
    ```bash
    make build-all-images 
    ```

2. Push the Docker images to registry:

    ```bash
    make push-all-images
    ```

### Single component

If you want to build and push Docker image for a single component, follow these steps:

1. Build a specific Docker image:
    
    For application defined under [cmd](../cmd) package use it names, e.g. for `och`:
    ```bash
    make build-app-image-och
    ```

    For tests defined under [test](../test) package use it names, e.g. for `e2e`:
    ```bash
    make build-test-image-e2e
    ```

2. Push the built Docker image to a registry:

    For application defined under [cmd](../cmd) package use it names, e.g. for `och`:
    ```bash
    make push-app-image-och
    ```

    For tests defined under [test](../test) package use it names, e.g. for `e2e`:
    ```bash
    make push-test-image-e2e
    ```

## Generators

To execute all generators, execute:

```bash
make generate
```

Read below sections to execute only a specific generator.

### Generate Go code from the OCF JSON Schemas 

This project uses the [quicktype](https://github.com/quicktype/quicktype) library, which improves development by 
generating Go code from the [JSON Schemas](../ocf-spec/0.0.1/schema).

Each time the specification is changed you can regenerate the Go struct. To do this, execute:
```bash
make gen-go-api-from-ocf-spec
```

To generate the Go structs for a specific OCF version, execute:

```bash
OCF_VERSION={VERSION} make gen-go-api-from-ocf-spec
```

> **NOTE:** Go structs are generated in [`pkg/sdk/apis`](../pkg/sdk/apis) package.

### Generate K8s resources

This project uses [controller-gen](https://sigs.k8s.io/controller-tools/cmd/controller-gen) for generating utility code and Kubernetes YAML.
Code and manifests generation is controlled by the presence of special ["marker comments"](https://sigs.k8s.io/kubebuilder/docs/book/src/reference/markers.md) in Go code.

Each time the Go code related to K8s controller is change, you need to update generated resources. To do this, execute:

```bash
make gen-k8s-resources
```

### Generate code from GraphQL schema

This project uses the [GQLGen](https://github.com/99designs/gqlgen) library, which generates the Go struct and server from GraphQL schema definition.

In Voltron project we have three GraphQL schemas:
- [Engine](../pkg/engine/api/graphql/schema.graphql)
- [Local OCH](../pkg/och/api/local/schema.graphql)
- [Public OCH](../pkg/och/api/public/schema.graphql)

Each time the GraphQL schema is change, you need to update generated resources. To do this, execute:

```bash
make gen-graphql-resources
```
