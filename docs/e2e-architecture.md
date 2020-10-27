# E2E Architecture

The document describes high-level Voltron architecture, all components in the system and interactions between them.

> **NOTE**: This document showcases architecture for Alpha phase. After Alpha stage the document will be updated to describe the target GA architecture. 

## Components

The following diagram visualizes all components in the system:

![Components](assets/components.svg)

### OCF

Open Capability Format is not a component per-se, but it's also an important part of the system. OCF is a specification of manifests for every entity in the system.
It is stored in a form of multiple JSON Schema files. From the JSON schemas all internal [SDK](#sdk) Go structs are generated.

OCF manifests are stored in [OCH](#och).

### UI

UI is the easy way to manage Actions and consume [OCH](#och) content.

It exposes the following functionalities:
- see available Implementations for a given system, grouped by InterfaceGroups and Interfaces,
- see the available TypeInstances, along with theirs status and metrics,
- render and execute Actions, with advanced rendering mode support.

UI does all HTTP requests to [Gateway](#gateway).

UI may be deployed outside the Kubernetes cluster with core Voltron components.

### CLI

CLI is command line tool which makes easier with working the [OCF](#ocf) manifests.

Currently, CLI exposes the following features:
- validates OCF manifests against the OCF JSON Schemas,
- signs OCF manifests that are loaded by [OCH](#och). 

CLI utilizes [SDK](#sdk).

### Gateway

Gateway is a GraphQL reverse proxy. It aggregates multiple remote GraphQL schemas into a single endpoint. It enables UI to have a single destination for all GraphQL operations.

Based on the GraphQL operation, it forwards the query or mutation to a corresponding service:
- [Engine](#engine) - for Action CRUD operations,
- Local [OCH](#och) - for TypeInstance CRUD operations,
- Public [OCH](#och) - for read operations for all other manifests apart except TypeInstance.

It also runs an additional GraphQL server, which exposes single mutation to change URL for Public [OCH](#och) API aggregated by Gateway.

### Engine

Engine is responsible for validating, rendering and executing Actions.

It composes of two modules:
- Kubernetes operator, which handles Action validation, rendering and execution based on Action Custom Resources,
- GraphQL API server, which exposes platform-agnostic API for managing Actions.

Engine consumes both local and public OCH APIs via single Gateway endpoint:
- to resolve Action prerequisites based on TypeInstances, it uses Local [OCH](#och) API,
- to resolve other manifests such as Interface or Implementation, it uses Public [OCH](#och) API.

Engine utilizes [SDK](#sdk). To execute Actions, it uses Kubernetes Jobs, that executes [Argo](https://github.com/argoproj/argo) workflows.

### OCH

Open Capability Hub stores [OCF](#ocf) manifests and exposes API to access and manage them. It uses graph database as a storage for the data.

OCH works in two modes:
- Local OCH exposes GraphQL API for managing TypeInstances (create, read, delete operations),
- Public OCH, which exposes read-only GraphQL API for querying all OCF manifests except TypeInstances.

Manifests for Public OCH are populated with DB Populator, which directly populates the graph database with manifests from a given directory structure.

OCH utilizes [SDK](#sdk).

### SDK

SDK is a Go library with low-level and high-level functions used by [Engine](#engine), [OCH](#och) and [CLI](#cli).

SDK can be used by Users to interact with Voltron components in a programmatic way.

## Detailed interaction

The section contains detailed interaction diagrams, to understand how the system works in a higher level of detail.

### Executing Action

On the following diagram, User executes the WordPress install Action using UI.

> **NOTE:** To make the diagram more readable, Gateway component was excluded. Every operation proxied by Gateway is described with __(via Gateway)__ phrase.

![Sequence diagram for WordPress install Action](assets/action-sequence-diagram.svg)
