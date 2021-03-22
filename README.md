# Voltron

A simple way to manage applications and execute day-two operations.

## Overview

### In a nutshell

The key benefit which Voltron brings is interchangeable dependencies. Cluster Admin may configure preferences for resolving the dependencies (e.g. to prefer cloud-based or on-premise solutions). As a result, the end-user is able to easily install applications with multiple dependencies without any knowledge of platform-specific configuration.

Apart from installing applications, Voltron makes it easy to:
- execute day-two operations (such as upgrade, backup, and restore)
- run any workflow - to process data, configure the system, run serverless workloads, etc. The possibilities are virtually endless.

Voltron aims to be a platform-agnostic solution. However, the very first Voltron implementation is based on Kubernetes.

### Example

To explain Voltron in action, let's focus on Jira installation. Jira requires PostgreSQL.

From User perspective, the flow is easy.

1. User navigates to the Voltron App Catalog.
2. Once User clicks Install button for Jira in the App Catalog, PostgreSQL is configured according to Cluster Admin and User preferences:
   - Cluster Admin can configure Voltron to prefer cloud-based GCP solutions. In this case, if User Installs Jira on cluster, Voltron will provision GCP CloudSQL for PostgreSQL database and use it.
   - If on-premise solutions are preferred, PostgreSQL will be installed on the same Kubernetes cluster with Helm.
   - If User provides an existing PostgreSQL database installation, deployed anywhere, Voltron will use it for Jira installation.
3. Once the database is configured, Voltron Engine runs the action that deploys Jira on the cluster.
4. After deploying Jira, the Voltron Engine may run additional actions that install and configure other components, such as the identity provider and load balancer.

## Get started

The section contains useful links for getting started with Voltron.

- **Tutorials:** To learn how to install, use Voltron and develop content for it, follow our [tutorials](./docs/tutorial).
- **Development:** To run Voltron on your local machine and start contributing to Voltron, read [`development.md`](./docs/development.md) document.

To read full Voltron documentation, see the [`docs`](./docs) directory.

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
  │   └── tutorial            # Tutorials on how to use Voltron
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

The following Voltron components are in this repository:
- [Argo runner](./cmd/argo-runner) - Runner, which executes Argo workflows.
- [CloudSQL runner](./cmd/cloudsql-runner) - Runner, which manages Google CloudSQL instances.
- [Gateway](./cmd/gateway) - GraphQL Gateway, which consolidates Voltron GraphQL APIs in one endpoint.
- [Helm runner](./cmd/helm-runner) - Runner, which manages Helm releases.
- [Engine](./cmd/k8s-engine) - Kubernetes Voltron Engine, which handles Action execution.
- [CLI](./cmd/ocftool) - A CLI tool for working with OCF Manifests.
- [Open Capability Hub](./och-js) - Component, which stores OCF Manifests and exposes API to manage them.
- [DB Populator](./cmd/populator) - Component, which populates OCF Manifests into database.
- [Open Capability Format specification](./ocf-spec) - Specification, which defines the shape of Voltron entities

Check the README files in the component directories, for more details about how to use and develop them.

