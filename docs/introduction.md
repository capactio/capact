---
sidebar_position: 1
---

# Introduction

**Capact** (pronounced: "cape-act", /ˈkeɪp.ækt/) is a simple way to manage applications, infrastructure and execute day-two operations.

## Overview

### In a nutshell

The key benefit which Capact brings is interchangeable dependencies. Cluster Admin may configure preferences for resolving the dependencies (e.g. to prefer cloud-based or on-premise solutions). As a result, the end-user is able to easily install applications with multiple dependencies without any knowledge of platform-specific configuration.

Apart from installing applications, Capact makes it easy to:
- execute day-two operations (such as upgrade, backup, and restore)
- run any workflow - to process data, configure the system, run serverless workloads, etc. The possibilities are virtually endless.

Capact aims to be a platform-agnostic solution. However, the very first Capact implementation is based on Kubernetes.

### Example

To explain Capact in action, let's focus on [Mattermost](https://mattermost.org/) installation. Mattermost requires PostgreSQL.

From User perspective, the flow is easy.

1. User navigates to the Capact Action Catalog.
2. Once User clicks Install button for Mattermost in the App Catalog, PostgreSQL is configured according to Cluster Admin and User preferences:
   
   - Cluster Admin can configure Capact to prefer cloud-based GCP solutions. In this case, if User Installs Mattermost on cluster, Capact will provision GCP CloudSQL for PostgreSQL database and use it.
   - If on-premise solutions are preferred, PostgreSQL will be installed on the same Kubernetes cluster with Helm.
   - If User provides an existing PostgreSQL database installation, deployed anywhere, Capact will use it for Mattermost installation.
   
3. Once the database is configured, Capact Engine runs the action that deploys Mattermost on the cluster.
4. After deploying Mattermost, the Capact Engine may run additional actions that install and configure other components, such as the identity provider and load balancer.

## Components

The following Capact components reside in this repository:

- [Argo runner](https://github.com/capactio/capact/tree/main/cmd/argo-runner) - Runner, which executes Argo workflows.
- [CloudSQL runner](https://github.com/capactio/capact/tree/main/cmd/cloudsql-runner) - Runner, which manages Google CloudSQL instances.
- [Gateway](https://github.com/capactio/capact/tree/main/cmd/gateway) - GraphQL Gateway, which consolidates Capact GraphQL APIs in one endpoint.
- [Helm runner](https://github.com/capactio/capact/tree/main/cmd/helm-runner) - Runner, which manages Helm releases.
- [Engine](https://github.com/capactio/capact/tree/main/cmd/k8s-engine) - Kubernetes Capact Engine, which handles Action execution.
- [CLI](https://github.com/capactio/capact/tree/main/cmd/cli) - A CLI tool for interacting with Capact.
- [Open Capability Hub](https://github.com/capactio/capact/tree/main/och-js) - Component, which stores OCF Manifests and exposes API to manage them.
- [Populator](https://github.com/capactio/capact/tree/main/cmd/populator) - A CLI tool, which populates resources such as OCF manifests into database.
- [Open Capability Format specification](https://github.com/capactio/capact/tree/main/ocf-spec) - Specification, which defines the shape of Capact entities.

Check the README files in the component directories, for more details about how to use and develop them.
