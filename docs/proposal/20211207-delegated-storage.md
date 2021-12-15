# Delegated storage

Created on 2021-12-09 by Paweł Kosiec ([@pkosiec](https://github.com/pkosiec/))

## Overview

This document describes the way how we will approach dynamic, external data for TypeInstances.

<!-- toc -->

- [Motivation](#motivation)
  * [Goal](#goal)
  * [Non-goal](#non-goal)
- [Assumptions](#assumptions)
- [Prerequisites](#prerequisites)
- [Registering storage backends](#registering-storage-backends)
- [Workflow syntax - Create](#workflow-syntax---create)
- [Workflow syntax - Update](#workflow-syntax---update)
- [Storage backend service implementation](#storage-backend-service-implementation)
- [Uninstalling storage backends](#uninstalling-storage-backends)
- [GraphQL API](#graphql-api)
  * [List storage backends](#list-storage-backends)
  * [Get TypeInstance details](#get-typeinstance-details)
  * [TypeInstance create](#typeinstance-create)
  * [TypeInstance update](#typeinstance-update)
- [Dynamic TypeInstance projections](#dynamic-typeinstance-projections)
- [Rejected ideas](#rejected-ideas)
    + [Registering storage backends](#registering-storage-backends-1)
    + [Workflow syntax](#workflow-syntax)
    + [Storage backend service implementation](#storage-backend-service-implementation-1)
- [Consequences](#consequences)

<!-- tocstop -->

## Motivation

Capact stores the state in a form of TypeInstances with static data. That is problematic, as the data may quickly become outdated in case of any external change. For example, if you install Mattermost using Capact and delete Helm chart with Helm CLI, you will still be able to see Helm release TypeInstance of already not existing Helm release.

Also, we should be able to provide a way to store sensitive data, such as credentials, securely. Currently, we store them as plaintext inside our database.

The last reason, is that long-term we should replace Neo4j database for a more lightweight solution (like sqlite or PostgreSQL). Apart from being resource-hungry, Neo4j can be problematic also when it comes to licensing (GPL3). We believe that the pluggable back-ends concept could be a first step to abstract Neo4j and plug-in different storage backend. However, this should be treated as an additional, nice-to-have goal, or side effect of the Delegated Storage proposal.

### Goal

The main goal is to support the following use-cases:

- Store and retrieve secrets using external, secure solutions
    - Examples:
        - user credentials (e.g. for PostgreSQL, GitLab)
        - SSH key (e.g. for bastion)
        - Kubeconfig (for external clusters)

    - The following backends should be supported initially (sorted by priority):
        1. Vault
        1. Secrets encrypted with [SOPS](https://github.com/mozilla/sops) stored on Git repository
        pluggable

        Other backends:
        - AWS Secrets Manager
        - Google Key Management
    
- Store and retrieve dynamic data
    - Examples:
        - Kubernetes cluster (e.g. Flux's Helm releases or Kubernetes Secrets)
        - Apps configuration (e.g. Mattermost config)
        - External dependencies (e.g. S3 buckets)

    - The following backends should be supported initially (sorted by priority):
        1. Flux HelmRelease Custom Resources
        1. Git repositories (e.g. GitLab projects)
        1. S3 buckets (e.g. for external Terraform state)

- Ability to manage such TypeInstances manually (via CLI and maybe UI) and as a part of Action
    - Support such dynamic TypeInstance creation, update, and deletion
    - Define GraphQL API and Implementation workflow syntax

- Support for automatic TypeInstance creation and deletion
    - React on events to create/delete such TypeInstances (e.g. Kubernetes events for any change for Flux's HelmReleases).

- Support extensibility for upcoming backends

Also, the additional, nice-to-have goals are:

- Remove Neo4j dependency from Local Hub while preserving TypeInstance metadata, such as relations

### Non-goal

- Support external back-ends for Capact manifest storage (Public Hub)
- Remove Neo4j dependency from Public Hub

## Assumptions

1. Content Developer should be able to:
    1. Write manifests without specifying a storage backend (use default one configured by Cluster Admin). In this case, a static TypeInstance value is stored in the default storage backend.
    1. Specify a specific storage backend as a part of a given Implementation. This case supports both static and dynamic TypeInstance values.
1. Cluster Admin can configure default backend storage for static values.
1. There are two different cases when it comes CRUD operations on TypeInstances:
    1. CRUD operations on TypeInstance actually manages external resource (e.g. Vault). That is, CRUD operations on TypeInstances in Local Hub actually creates, updates and deletes a given resource.
    1. CRUD operations on TypeInstance represents external resources managed in different way (e.g. by running Helm.install). That is, CRUD operations on TypeInstances in Local Hub actually registers, unregisters and updates references for external state without changing them.

## Prerequisites

1. We implement these two Type features:
    - https://capact.io/docs/feature/type-features/#additional-references-to-parent-nodes
    - https://capact.io/docs/feature/type-features#find-types-based-on-prefix-of-parent-nodes
1. We add `TypeInstance.spec.backend` field (string)
1. **Optional:** [Add TypeInstance `alias`/`description` metadata field](https://github.com/capactio/capact/issues/579)

## Registering storage backends

1. For every storage backend, we create a dedicated Type:

    ```yaml
    ocfVersion: 0.0.1
    revision: 0.0.1
    kind: Type
    metadata:
      path: cap.type.helm.storage
    spec:
      additionalRefs:
        - "cap.core.hub.storage" 
      jsonSchema:
        value: # JSON schema with:
          {
          "$schema": "http://json-schema.org/draft-07/schema",
          "$id": "http://example.com/example.json",
          "type": "object",
          "title": "The root schema",
          "required": [
            "url",
            "additionalParametersSchema"
          ],
          "properties": {
            "url": { # url of hosted app, which implements storage backend ProtocolBuffers interface.
              "$id": "#/properties/url",
              "type": "string",
              "format": "uri"
            },
            "additionalParametersSchema": { # JSON schema which describes additional properties passed in Capact workflow
              "const": { # see http://json-schema.org/draft/2019-09/json-schema-validation.html#rfc.section.6.1.3
                "$id": "#/properties/additionalParameters",
                "type": "object",
                "required": [
                  "name",
                  "namespace"
                ],
                "properties": {
                  "name": {
                    "$id": "#/properties/additionalParameters/properties/name",
                    "type": "string"
                  },
                  "namespace": {
                    "$id": "#/properties/additionalParameters/properties/namespace",
                    "type": "string"
                  }
                },
                "additionalProperties": false
              }
            },
            "acceptValue": { # specifies if a given storage backend (app) accepts TypeInstance value while creating/updating TypeInstance, or just additionalParameters
              "$id": "#/properties/acceptValue",
              "type": "boolean",
              "const": false # in this case - no
            },
          },
          "additionalProperties": false
        }
    ```

    It should follow the convention of having `url`, `acceptValue`, and `additionalParametersSchema` fields. As `additionalParameters` are optional, the `additionalParametersSchema` field is nullable.

    We can validate such convention using custom logic for Type validation. In case of the `cap.core.hub.storage` additional reference, we could prevent uploading such Type if the JSON schema don't meet our conditions.

    > **NOTE:** See also the [Rejected ideas](#rejected-ideas) section to learn why a generic validation idea was rejected.

1. To install new storage backend, Cluster Admin has two options:

    - use Capact Actions (e.g. `cap.interface.capactio.capact.hub.storage.helm-release.install`).
    - Register a storage backend by creating such TypeInstance.

    Regardless the option, at the end there is one TypeInstance produced:
      
      ```yaml
      id: 3ef2e4ac-9070-4093-a3ce-142139fd4a16
      typeRef:
        path: cap.type.helm.storage
        revision: 0.1.0
      latestResourceVersion:
        metadata:
          alias: helm-storage # new field, more user-friendly description of such TypeInstance
          attributes:
          - path: "cap.core.attribute.hub.storage.backend" # related to GraphQL implementation
            revision: 0.1.0
        value:
          url: "helm-release.default:50051"
          acceptValue: false
          additionalParametersSchema: {
            "$id": "#/properties/additionalParametersSchema",
            "type": "object",
            "required": [
              "name",
              "namespace"
            ],
            "properties": {
              "name": {
                "$id": "#/properties/additionalParametersSchema/properties/name",
                "type": "string"
              },
              "namespace": {
                "$id": "#/properties/additionalParametersSchema/properties/namespace",
                "type": "string"
              }
            },
            "additionalProperties": false
          }
      backend:
        id: "a36ed738-dfe7-45ec-acd1-8e44e8db893b" # new immutable property - contains TypeInstance ID
          # if not provided during TypeInstance creation, fallback to default one (get TypeInstance with proper Attribute and write its ID in this property)
      ```

1. In fresh Capact installation, there is one TypeInstance already preregistered:

    ```yaml
    id: a36ed738-dfe7-45ec-acd1-8e44e8db893b
    typeRef:
        path: cap.core.type.hub.storage.postgresql
        revision: 0.1.0
    latestResourceVersion:
      metadata:
        alias: capact-postgresql
        attributes:
        - path: cap.core.attribute.hub.storage.default # if more such Typeinstances with default Attribute, select first one
          revision: 0.1.0 
        - path: cap.core.attribute.hub.storage.backend
          revision: 0.1.0
      value:
        url: "storagebackend-handlers.capact-system:50051"
        acceptValue: true
        additionalParametersSchema: null
    backend: 
      abstract: true # Special keyword which specifies the built-in storage option which stores already all other metadata. Effectively, it would be the same database as the PostgreSQL accessed via Capact storage backend service (`storagebackend-handlers.capact-system:50051`), but accessed directly.
    ```

    - The one preregistered storage backend is Capact PostgreSQL. It uses special `backend` property: `abstract: true`. This is a reserved system keyword and user cannot create a TypeInstance with such. To use Capact PostgreSQL, the `capact-postgresql` TypeInstance ID has to be referenced as a backend.
    - Default storage backend should have `additionalParameters` empty (`null`) or optional, in order to work properly.
      
      We can validate that using custom logic in TypeInstance validation. When User wants to add the `cap.core.attribute.hub.storage.default` Attribute we can see if the `additionalParameters` meet our conditions and if not, prevent such change.

## Workflow syntax - Create

1. In workflow, Content Developer can specify requirements for a given backend:

    ```yaml
    requires:
      cap.core.type.hub.storage: # Optional - Content Dev specifies such requirement to force use a given backend
        allOf:
          - typeRef:
              path: cap.type.helm.storage
              revision: 0.1.0
              alias: helm-storage
    ```

    - This workflow cannot be run unless there is a `helm-release` storage backend installed (where `helm-release` is only workflow alias).
    - If there are no specific storage backend requirements set, the default backend will be used.

1. Content Developer outputs one of the following Argo workflow artifacts:

    > **NOTE:** Before this proposal, the whole Argo workflow artifact was treated as a value. Now we would need to change that.

    1. To store a given value on default backend or backend without any required additional parameters, which also accepts TypeInstance value:

        ```yaml
        # option 1: save a specific value on a storage backend
        # a given backend
        value: foo
        ```

    1. To point to some external data for a given storage backend:

        ```yaml
        # option 2: register something which already exist as external TypeInstance - based on `backend.additionalParameters`
        backend:
          additionalParameters:
            name: release-name
            namespace: release-namespace
        ```

        However, the `additionalParameters` are backend-specific properties, which means Content Developer need to explicitly specify the backend as described later.

    1. To save a specific value with additional parameters:
    
        For example, for an implementation of Kubernetes secrets storage backend, which actually creates and updates these secrets during TypeInstance creation:

        ```yaml
        # option 3: save a specific value on an external backend with some additional parameters
        value: foo
        backend:
          additionalParameters:
            key: bar
            value: baz
        ```

        The storage backend has to have `additionalParametersSchema` specified, as well as the `acceptValue` property set to `true`.

    In that way, someday we will be able to extend such approach with additional properties:
    
    ```yaml    
    instrumentation: # someday - if we want to unify the approach
      health:
        endpoint: foo.bar/healthz
      # (...)
    ```

    Such `instrumentation` data would be also stored in the same storage backend as the `value`. If Content Developer wants to store it somewhere else, then an additional Argo artifact to produce is needed.

1. Then, Content Developer specifies the Argo workflow artifact as output TypeInstance with familiar syntax:

    ```yaml
    # default - static
    capact-outputTypeInstances:
      - name: mattermost-config
        from: additional
        # no backend definition -> use default (default storage backend (TypeInstance) is annotated with `cap.core.attribute.hub.storage.default`)

    # option 2 - specific backend (defined in `Implementation.spec.requires` property)
    capact-outputTypeInstances:
      - name: helm-release
        from: helm-release
        backend: helm-storage # new property -> alias defined in the `requires` section
    ```

1. The automatically injected TypeInstance upload step, receives the following payload:

    ```yaml
    typeInstances:
    - alias: helm-release
      attributes: []
      createdBy: default/act-l49vh-30c7a078-6a77-475c-94dd-7466f56447ce
      typeRef:
        path: cap.type.helm.chart.release
        revision: 0.1.0
      value: null 
      backend:
        # Fields which are a part of TypeInstance
        id: 3ef2e4ac-9070-4093-a3ce-142139fd4a16 # helm-release backend - resolved UUID based on the injected TypeInstance

        # Fields which are a part of TypeInstanceResourceVersion (can be changed later via TypeInstance Update):
        additionalParameters:
          name: release-name
          namespace: release-namespace
    usesRelations: # automatically create relation between TypeInstance using a given backend
    - from: helm-release
      to: 3ef2e4ac-9070-4093-a3ce-142139fd4a16
    ```

1. Hub receives the following GraphQL mutation based on the payload fields from point above:

    ```graphql
    mutation CreateTypeInstances {
      createTypeInstances(
        in: {
          # ... payload
        }
      ) {
        id
        alias
      }
    }
    ```

1. Based on the `backend` data:

    1. Hub resolves details about the service (TypeInstance details)
    1. Hub validates whether the storage backend accepts TypeInstance value (`acceptValue` property). If not, and the value has been provided, it returns error.
    1. Hub calls the registered storage backend service `onCreate` hook:

        ```javascript
        // TODO: Pseudocode, change to actual HTTP requests / ProtoBuf definition
        onCreate(typeInstanceID, additionalParameters?, value?): (additionalParameters?, error)
        ```

        This hook can mutate `additionalParameters`.

    1. Hub validates `additionalParameters` against JSON schema saved in the storage backend TypeInstance.
    1. Saves TypeInstance metadata in the core Hub storage backend, which contains all metadata of the TypeInstances and  theirs relations.

      ``` yaml
      id: # generated UUID
      typeRef:
        path: cap.core.type.hub.storage.redis
        revision: 0.1.0
      latestResourceVersion:
        resourceVersion: 1
        backend:
          additionalParameters: # additional parameters that might be modified via the service handling `onCreate` hook
            name: release-name
            namespace: release-namespace 
      backend: 
        id: 3ef2e4ac-9070-4093-a3ce-142139fd4a16 # helm-release backend - resolved UUID based on the injected TypeInstance
      ```

## Workflow syntax - Update

Similarly as with create, Content Developer specifies in the workflow:

```yaml
capact-updateTypeInstances:
- name: testUpdate
  from: update
```

where the `update` Argo artifact can contain `value` and / or `additionalParameters`.

For additions in GraphQL API, see the [GraphQL API](#graphql-api) section.

## Storage backend service implementation

Capact Local Hub calls proper storage backend service while accessing the TypeInstance value or lock state.

1.  The registered storage backend service needs to implement the following gRPC + Protocol Buffers API:

    TODO: Pseudocode, change to actual HTTP requests / ProtoBuf definition

    ```proto
    syntax = "proto3";

    service StorageBackendService {
    rpc Search(SearchRequest) returns (SearchResponse);
    }

    // TypeInstance ResourceVersion value
    getValue(typeInstanceID, additionalParameters?, resourceVersion): (value, error)

    onCreate(typeInstanceID, additionalParameters?, value?): (additionalParameters?, error)

    onUpdate(
      typeInstanceID,
      old: {additionalParameters?, resourceVersion},
      new: {additionalParameters?, resourceVersion},
      ): (additionalParameters?, error)

    onDelete(typeInstanceID, additionalParameters?) error

    // TypeInstance locking 
    lockedBy(typeInstanceID, additionalParameters?) (string, error)
    onLock(typeInstanceID, additionalParameters?) error
    onUnlock(typeInstanceID, additionalParameters?) error
    ```

  An implementation of such service may vary between two use cases:

  1. CRUD operations on output TypeInstance actually manages external resource (e.g. Vault) -> onCreate, onUpdate, and onDelete actually creates, updates and deletes a given resource.
  1. output TypeInstance represents external resources managed in different way (e.g. via Capact actions - like Helm Runner). IMO we shouldn't move actual Helm release installation to TypeInstance "constructor").

      - The service can also implement watch for external resources (e.g. Kubernetes secrets) and call `createTypeInstances` and `deleteTypeInstances` Hub mutations. We may provide Go framework to speed up such development, similarly as we have with Runner concept.

1. The service could be implemented using one of the following solutions, or other alternatives:

  - [Dapr secrets](https://docs.dapr.io/developing-applications/building-blocks/secrets/secrets-overview/)
  - [Kubernetes external secrets](https://github.com/external-secrets/kubernetes-external-secrets)
  - [vault-k8s](https://github.com/hashicorp/vault-k8s)
  - [db](https://upper.io/v4/getting-started/)
  - [go-cloud](https://github.com/google/go-cloud)
  - [stow](https://github.com/graymeta/stow)

## Uninstalling storage backends

As described in the [Workflow syntax - Create](#workflow-syntax---create) section, every TypeInstance that uses a given storage backend, will use the `uses` property set:

```yaml
usesRelations: # automatically create relation between TypeInstance using a given backend
  - from: helm-release
    to: 3ef2e4ac-9070-4093-a3ce-142139fd4a16 # Helm storage backend
```

In that way, a given storage backend will contain `usedBy` relations.

According to the accepted [Rollback](./20201209-action-rollback.md) proposal:
- User won't be able to delete TypeInstance manually, but will run Rollback procedure instead.
- A given TypeInstance which contain any `usedBy` reference, cannot be deleted unless all related TypeInstances are deleted.

In that way, we prevent removal of any storage backend that is used.

## GraphQL API

The new GraphQL API can be used both on CLI and UI.

### List storage backends

To list all available StorageBackends in Hub:

```graphql
query {
   types(filter: { pathPattern: "cap.core.hub.storage.*" }) {
        name
        prefix
        path
    }
}
```

To list all configured StorageBackends in Capact:

```graphql
# Ideally, but it could be too complicated:
query ListTypeInstancesWithTypeRefFilter {
  typeInstances(
    filter: { typeRef: { path: "cap.core.hub.storage.*" } } # queries public Hub to fetch all Types attached to `cap.core.hub.storage` and return all TypeInstances which are of one of these TypeRefs
  ) {
    ...TypeInstance
  }
}

# Alternatively: introduce `cap.core.attribute.hub.storage.backend` Attribute and simply do:

query ListTypeInstancesWithAttributesAndTypeRefFilter {
  typeInstances(
    filter: {
      attributes: [
        { path: "cap.core.attribute.hub.storage.backend", rule: INCLUDE }
      ]
    }
  ) {
    ...TypeInstance
  }
}
```

### Get TypeInstance details

To see the value for all TypeInstances, we can use the following query:

```graphql
query ListTypeInstances {
  typeInstances {
    id
    typeRef {
      path
      revision
    }
    lockedBy # resolver which calls proper storage backend service to ask for lock status
    latestResourceVersion {
      resourceVersion
      createdBy
      metadata {
        attributes {
          path
          revision
        }
      }
      spec {
        value # resolver which calls proper storage backend service to ask for a given ResourceVersion value
      }
      backend { # new property
        additionalParameters
      }
    }
    backend { # new property
      # Initially, we can return only TypeInstance ID
      """TypeInstance ID"""
      id

      # Later, we can resolve full details here based on the ID
      latestResourceVersion {
        metadata {
          alias # new field
        }
        value # url + additionalParametersSchema
      }
      
    }
  }
}
```

### TypeInstance create

```graphql
input CreateTypeInstanceBackendInput { # New input
  id: ID # storage backend TypeInstance ID. Optional, as it will fallback to default one if not provided

  additionalParameters: Any # Properties which will be populated into the first Resource Version of the newly created TypeInstance
}

input CreateTypeInstanceInput {
  # (...)
  alias: String
  value: Any

  backend: CreateTypeInstanceBackendInput # new property
}

input CreateTypeInstancesInput {
  typeInstances: [CreateTypeInstanceInput!]!
  usesRelations: [TypeInstanceUsesRelationInput!]!
}

type Mutation {
  createTypeInstances(
      in: CreateTypeInstancesInput!
    ): [CreateTypeInstanceOutput!]!
}
```

### TypeInstance update

To properly handle TypeInstance update, the following additions to the API need to be made:

```graphql
input UpdateTypeInstanceBackendInput { # New input
  additionalParameters: Any
}

"""
At least one property needs to be specified.
"""
input UpdateTypeInstanceInput {
  # (...)
  value: Any

  backend: UpdateTypeInstanceBackendInput # New property
}

input UpdateTypeInstancesInput {
  # ...

  id: ID!
  typeInstance: UpdateTypeInstanceInput!
}

type Mutation {
  updateTypeInstances(in: [UpdateTypeInstancesInput]!): [TypeInstance!]!
}
```

## Dynamic TypeInstance projections

**TODO:** Figure out how to solve problem with static TypeInstances like `mattermost-config`:
  - Support multiple different TypeInstances for a given projection
  - TypeInstance composition?

<!-- In the first approach IMO we can skip support for multiple different TypeInstances

Maybe we will need to have sth like Porter has? Mixins for each "backend", e.g. option to https://github.com/MChorfa/porter-helm3#outputs. I saw that they have this pattern also for other "platforms" https://porter.sh/mixins/, maybe we could even reuse it without implementing that on our side.

Unfortunately, it will be on Content Developer side to create a proper pattern for resolution. -->

## Rejected ideas

#### Registering storage backends

1. Enforcing convention of having storage backend defined as Type with `uri` and `additionalParametersSchema`.

    Initially, we can't enforce such convention. That could be possible if we implement an ability to define validating JSON schema for Type nodes, and use such schemas to validate Type values which define `additionalRefs`. For example, the `cap.core.hub.storage` node could have JSON Schema defined, which validates Type values (JSON schema) attached to such node. In the end, that would be JSON schema validating another JSON schema.
    
    **Reason:** It is possible, but it's complex and brings too little benefits for now to implement it.

1. Using Global / Action Policy to specify the default storage backend.

    The benefit is that we could enforce empty or optional `additionalParameters` for such default storage backend.

    **Reason:** The Policy is already too big.

1. Adding optional `TypeInstance.metadata.name` or `alias`, which is unique across all TypeInstances and immutable regardless resourceVersion. It would allow easier referencing storage backends in the `TypeInstance.spec.backend` field:

    ```yaml
    id: 3ef2e4ac-9070-4093-a3ce-142139fd4a16
    metadata:
      name: helm-storage
    typeRef:
      path: cap.type.helm.storage
      revision: 0.1.0
    latestResourceVersion:
       #...
    backend: capact-postgresql # immutable - contains TypeInstance ID or unique alias
      # if not provided, fallback to default one during TypeInstance creation
    ```

    **Reason:** It is not really needed as we can use unique IDs to reference such backends. Also, we can expose GraphQL API which resolves details of a given storage backend based on the ID.

1. Dedicated entity of StorageBackend

    Such resource could reside in Local Hub, but it wouldn't be an OCF manifest. Cluster Admin should be able to manage them via GraphQL API, CLI and UI.

    **Reasons:**
    - We would still need some kind of StorageBackend templates (with `additionalParametersSchema` JSON schema) in public Hub
    - How we would be able to output such as a result of an Action? It could be done in a hacky way to output it as a side effect of running Action (not explicitly), but that would be definitely not elegant
    - We would need to have additional API

#### Workflow syntax

1. Keep the Argo artifact value as it is, and add additional syntax:

    ```yaml
    # default - static
    capact-outputTypeInstances:
      - name: mattermost-config # still static
        from: additional
        # no backend definition -> used default (default storage backend (TypeInstance) is annotated with `cap.core.attribute.default`)

    # option 2 - create TypeInstance on external storage
    capact-outputTypeInstances:
      - name: helm-release
        from: helm-release # values
        additionalParameters: "{steps.foo.output.artifacts.foo}" 
        backend: vault

    # option 3 - register something which already exist as external TypeInstance - based on additionalParameters
    capact-outputTypeInstances:
      - name: helm-release
        backend: helm-storage
        additionalParameters: "{steps.foo.output.artifacts.foo}"
    ```

    **Reason:** More complex usage in the workflow, and more complex implementation as well

#### Storage backend service implementation

1. Using Actions as a way to do CRUD operations (separate Interface/Implementation per Create/Update/Get/Delete operation)
 
    **Reason:** While the idea may seem exciting, that would be really time consuming and ineffective. We are too far from the point at where we can think about such solution. 

## Consequences

**TODO: Write consequences**
