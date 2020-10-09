# Voltron

## Overview

This repository contains the Go codebase for the Voltron project.

## Project structure

The repository has the following structure:

```
  ├── cmd
  │ ├── gateway                 # GraphQL Gateway that consolidates all Voltron GraphQL APIs in one endpoint
  │ ├── k8s-engine              # Kubernetes Voltron engine
  │ └── och                     # OCH server
  │
  ├── deploy                    # Deployment configurations and templates
  │ └── kubernetes              # Kubernetes related deployment (Helm charts, CRDs etc.)
  │
  ├── docs                      # Documentation related to the project
  │ └── investigation           # Investigations and proof of concepts files
  │
  ├── hack                      # Scripts used by the Voltron developers
  │
  ├── ocf-spec                  # Open Capability Format Specification
  │
  ├── pkg                       # Component related logic.
  │ ├── db-populator            # Populates Voltron entities to graph database
  │ ├── engine                  # Voltron platform-agnostic engine
  │ ├── gateway                 # GraphQL Gateway
  │ ├── och                     # Open Capability Hub server 
  │ ├── runner                  # Voltron runners, e.g. Argo Workflow runner, Helm runner etc.
  │ └── sdk                     # SDK for Voltron eco-system
  │
  │── test                      # Cross-functional test suites
  │
  ├── Dockerfile                # Dockerfile template to build applications and tests images
  │
  └── go.mod                    # Manages Go dependency. There is single dependency management across all components in this monorepo
```

## Development

Read this document to learn how to develop the project.

### Prerequisites

* [Go](https://golang.org/dl/) at least 1.15
* [Docker](https://www.docker.com/)
* Make

### Install dependencies

This project uses `go modules` as a dependency manager. To install all required dependencies, use the following command:

```bash
go mod download
```

### Run tests

To run all unit and lint tests, execute the following command:

```bash
make test-unit
make test-lint
```

To run integration test on [`kind`](https://kind.sigs.k8s.io/docs/user/quick-start/), execute the following command:

```bash
make test-integration
```

> **NOTE:** The required dependency (kind, Helm) can be installed by the script when you export `SKIP_DEPS_INSTALLATION=false` environment variable. By default, dependency installation is skipped.  

### Verify the code

To check if the code is correct and you can push it, use the `make` command. It builds the application, runs tests, checks the status of the vendored libraries, runs the static code analysis, and checks if the formatting of the code is correct.

### Test coverage

To generate the unit test coverage HTML report, execute the following command: 

```bash
make test-cover-html
```

> **NOTE:** The default browser with the generated report opens automatically.

### Run development local cluster on `kind` 

To create a `kind` cluster with preinstalled Voltron components, execute the following command:
```
make dev-cluster
```

> **NOTE:** To delete the development cluster, run: `kind delete cluster --name kind-dev-voltron`

#### Update images

To build Voltron Docker images with your changes and update them on dev cluster, execute the following command:

```
make dev-cluster-update
```

### Build and push Docker images

If you want to build all Docker images with your changes and push them to a registry, follow these steps:

1. Build all Docker images:
    
    ```bash
    make build-all-images 
    ```

2. Configure environment variables pointing to your registry, for example:

    ```bash
    export DOCKER_PUSH_REPOSITORY=gcr.io/projectvoltron/
    export DOCKER_TAG=latest
    ```

3. Push the Docker images to registry:

    ```bash
    make push-all-images
    ```

If you want to build and push Docker image for a single component, follow these steps:

1. Build a specific Docker image:
    
    For application defined under [cmd](./cmd) package use it names, e.g. for `och`:
    ```bash
    make build-app-image-och
    ```

    For tests defined under [test](./test) package use it names, e.g. for `e2e`:
    ```bash
    make build-test-image-e2e
    ```

3. Push the built Docker image to a registry:

    For application defined under [cmd](./cmd) package use it names, e.g. for `och`:
    ```bash
    make push-app-image-och
    ```

    For tests defined under [test](./test) package use it names, e.g. for `e2e`:
    ```bash
    make push-test-image-e2e
    ```

> **NOTE:** Registry can be configured exactly in the same way as specified in the previous section.


### Generate code Go structs from OCF JSON Schemas manifests 

This project uses the [quicktype](https://github.com/quicktype/quicktype) library, which improves development by 
generating Go code from the [JSON Schemas](ocf-spec/0.0.1/schema).

Each time the specification is changed you can regenerate the Go struct. To do this, execute the following command:
```bash
make gen-go-api
```

To generate the Go structs for a specific OCF version, execute the following command:
```bash
OCF_VERSION={VERSION} make gen-go-api
```

> **NOTE:** Go structs are generated in [`pkg/sdk/apis`](./pkg/sdk/apis) package.
