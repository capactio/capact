# Voltron

## Overview

This repository contains the codebase for the Voltron project.

## Project structure

The repository has the following structure:

```
  .
  ├── cmd                     # Main application directory
  │
  ├── deploy                  # Deployment configurations and templates
  │
  ├── docs                    # Documentation related to the project
  │   ├── investigation       # Investigations and proof of concepts files
  │   ├── proposal            # Proposals for handling new features
  │   └── tutorial            # Tutorial on how to use Voltron
  │
  ├── hack                    # Scripts used by the Voltron developers
  │
  ├── internal                # Private component code
  │
  ├── ocf-spec                # Open Capability Format Specification
  │
  ├── och-content             # OCF Manifests for the Open Capability Hub
  │
  ├── och-js                  # Node.js implementation of Open Capability Hub
  │
  ├── pkg                     # Public component and SDK code
  │
  ├── test                    # Cross-functional test suites
  │
  ├── Dockerfile              # Dockerfile template to build applications and tests images
  │
  └── go.mod                  # Manages Go dependency. There is single dependency management across all components in this monorepo
```

## Components

The following Voltron components are in this repository
- [Argo runner](./cmd/argo-runner) - Runner, which executes Argo workflows.
- [CloudSQL runner](./cmd/cloudsql-runner) - Runner, which manages Google CloudSQL instances.
- [Gateway](./cmd/gateway) - GraphQL Gateway, which consolidates Voltron GraphQL APIs in one endpoint.
- [Helm runner](./cmd/helm-runner) - Runner, which manages Helm releases.
- [Engine](./cmd/k8s-engine) - Kubernetes Voltron Engine, which handles Action execution.
- [ocftool](./cmd/ocftool) - A CLI tool for working with OCF Manifests.
- [Open Capability Hub](./och-js) - Component, which stores OCF Manifests and exposes API to manage them.
- [Mocked Open Capability Hub](./cmd/och) - Mocked version of the Open Capability Hub. It does not use the database, but exposes mocks from `./hack/mock/graphql` on the API.
- [DB Populator](./cmd/populator) - Component, which populates OCF Manifests into database.

Check the README files in the component directories, for more details about how to use and develop them.

## Development

Read [this](./docs/development.md) document to learn how to develop the project. 
