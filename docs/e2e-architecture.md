# E2E Architecture

The document describes high-level Voltron architecture, all components in the system and interactions between them.

> **NOTE**: This document showcases architecture for Alpha phase. After Alpha stage the document will be updated to describe the target GA architecture. 

## Components

The following diagram visualizes all components in the system:

![Components](assets/components.svg)

To learn about responsibilities for every components, read the subsections.

### UI

UI is the easy way to manage Actions, TypeInstances and use the public OCH content.

It exposes the following functionalities:
- See available Implementations for a given system, grouped by InterfaceGroups and Interfaces,
- See the available TypeInstances, along with theirs status and metrics,
- Render & execute Actions, with advanced rendering mode support.

Underneath, UI does all HTTP requests to [Gateway](#gateway).

UI may be deployed outside Kubernetes.

### CLI

CLI is command line tool which makes easier with working the OCF manifests.

Currently, CLI exposes the following features:
- validates OCF manifests against the OCF JSON Schemas
- signs OCF manifests that are loaded by OCH. 

CLI utilizes [SDK](#sdk).

### Gateway

Gateway is a GraphQL reverse proxy. It aggregates multiple remote GraphQL schemas into a single endpoint. It enables UI to have a single destination for all GraphQL operations.

Based on the GraphQL operation, it forwards the query or mutation to a corresponding service:
- Engine - for Action CRUD operations,
- Local OCH - for TypeInstance CRUD operations,
- Public OCH - for read operations for all other manifests apart except TypeInstance.

It also runs an additional GraphQL server, which exposes single mutation to change URL for Public OCH API aggregated by Gateway.

### Engine

Engine utilizes [SDK](#sdk).

#### GraphQL API

#### Kubernetes Controller

### OCH

OCH utilizes [SDK](#sdk).

#### Local

#### Public

### SDK

SDK is a Go library with low-level and high-level functions used by Engine, OCH and CLI.

SDK can be used by Users to interact with Voltron components in a programmatic way.

## Detailed interaction

### Executing Action

On the following diagram, User executes the WordPress install Action using UI.

![Sequence diagram for WordPress install Action](assets/action-sequence-diagram.svg)

## TODO
- Describe all components
- Update detailed sequence diagram
