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
- [Registering Storage Backends](#registering-storage-backends)
- [Workflow syntax - Create](#workflow-syntax---create)
- [Workflow syntax - Update](#workflow-syntax---update)
- [Storage Backend service implementation](#storage-backend-service-implementation)
- [GraphQL API](#graphql-api)
- [Dynamic TypeInstance projections](#dynamic-typeinstance-projections)
- [Examples](#examples)
- [Rejected ideas](#rejected-ideas)
    + [Registering Storage Backends](#registering-storage-backends-1)
    + [Workflow syntax](#workflow-syntax)
    + [Backends implementation](#backends-implementation)
- [Consequences](#consequences)

<!-- tocstop -->

## Motivation

Capact stores the state in a form of TypeInstances with static data. That is problematic, as the data may quickly become outdated in case of any external change. For example, if you install Mattermost using Capact and delete Helm chart with Helm CLI, you will still be able to see Helm release TypeInstance of already not existing Helm release.

Also, we should be able to provide a way to store sensitive data, such as credentials, securely. Currently, we store them as plaintext inside our database.

The last reason, is that long-term we should replace Neo4j database for a more lightweight solution (like sqlite or PostgreSQL). Apart from being resource-hungry, Neo4j can be problematic also when it comes to licensing (GPL3). We believe that the pluggable back-ends concept could be a first step to abstract Neo4j and plug-in different storage backend.

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
    1. Write manifests without specifying a storage backend (use default one configured by Cluster Admin).
    1. Specify a specific storage backend as a part of a given Implementation.
1. There are two different cases when it comes CRUD operations on TypeInstances:
    1. CRUD operations on TypeInstance actually manages external resource (e.g. Vault) -> CRUD operations on TI in Local Hub actually creates, updates and deletes a given resource.
    1. CRUD operations on TypeInstance represents external resources managed in different way (e.g. by running Helm.install). CRUD operations on TI in Local Hub actually registers, unregisters and updates references for external state without changing them.

## Prerequisites

1. We implement these two Type features:
    - https://capact.io/docs/feature/type-features/#additional-references-to-parent-nodes
    - https://capact.io/docs/feature/type-features#find-types-based-on-prefix-of-parent-nodes

1. (Optional) We add optional `TypeInstance.metadata.name`, which is **unique across all TypeInstances** and immutable regardless resourceVersion. See similar issue: [#579](https://github.com/capactio/capact/issues/579)
1. We add `TypeInstance.spec.backend` field (string)

## Registering Storage Backends

1. For every Storage Backend, we create a dedicated Type:

    It should follow convention of having `url` and `additionalParametersSchema` fields.

    ```yaml
    ocfVersion: 0.0.1
    revision: 0.0.1
    kind: Type
    metadata:
      path: cap.type.helm.storage
    spec:
      additionalRefs:
        - "cap.core.hub.storage" 
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
          "url": { # url of hosted app, which implements `BackendStorage` ProtocolBuffers interface.
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
          }
        },
        "additionalProperties": false
      }
    ```

    **TODO:** How we can enforce such convention? Type composition?

1. To install new Storage Backend, Cluster Admin has two options:

    - use Capact Actions (e.g. `cap.interface.capactio.capact.hub.storage.helm-release.install`).
    - Register a storage backend by creating such TypeInstance.

    Regardless the option, at the end there is one TypeInstance produced:
      
      ```yaml
      id: 3ef2e4ac-9070-4093-a3ce-142139fd4a16
      metadata:
        name: helm-storage
      typeRef:
        path: cap.type.helm.storage
        revision: 0.1.0
      latestResourceVersion:
        metadata:
          attributes:
          - path: cap.core.attribute.hub.storage.backend # related to GrapHQL implementation
            revision: 0.1.0
        value:
          url: "helm-release.default:50051"
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
      backend: capact-redis # immutable - contains TypeInstance ID or unique alias
        # if not provided, fallback to default one 
      ```

1. In fresh Capact installation, there is one TypeInstance already preregistered:

    **TODO:** Would Redis be really the default storage? Probably PostgreSQL will be better, as we treat it as a "builtin" Backend as well (to store Attributes, relations, etc.)

    ```yaml
    id: a36ed738-dfe7-45ec-acd1-8e44e8db893b
    name: capact-redis
    typeRef:
        path: cap.core.type.hub.storage.redis
        revision: 0.1.0
    latestResourceVersion:
      metadata:
        attributes:
        - path: cap.core.attribute.default # if more such Typeinstances with default Attribute, select first one
          revision: 0.1.0 
        - path: cap.core.attribute.hub.storage.backend
          revision: 0.1.0
      value:
        url: "storagebackend-handlers.capact-system:50051"
        additionalParameters: null
    backend: builtin # the storage option which stores already all other metadata. It could be the same Redis instance
    ```

    Default storage backend should have `additionalParameters` empty, or, optional.

    **TODO:** How to enforce that default backend doesn't have any additionalParameters needed? Maybe we shouldn't enforce that at all, as there could be additionalParameters, but not required. We could enforce that if it is implemented on Policy level...

## Workflow syntax - Create

1. In workflow, Content Developer can specify requirements for a given backend:

    ```yaml
    requires:
      cap.core.type.hub.storage: # Content Dev specifies such requirement only if he/she wants to force use a given backend
        allOf:
          - typeRef:
              path: cap.type.helm.storage
              revision: 0.1.0
              alias: helm-storage
    ```

    - This workflow cannot be run unless there is a `helm-release` StorageBackend installed (where `helm-release` is only workflow alias).
    - If there are no specific storage backend requirements set, the default backend will be used.

1. Content Developer outputs the following Argo workflow artifact:

    > **NOTE:** Before this proposal, the whole Argo workflow artifact was treated as a value. Now we would need to change that.

    As value (if he/she uses default backend or backend without any required additional params):

    ```yaml
    value: foo # option 1: save a specific value on an external backend
    ```

    or

    ```yaml
    additionalParameters: # option 2: register something which already exist as external TypeInstance - based on additionalParameters
      # however, additionalParameters are backend-specific properties, which means Content Dev need to explicitly specify the backend as described later.
      name: release-name
      namespace: release-namespace
    ```

    or both:

    ```yaml
    value: foo # option 3: save a specific value on an external backend with some additional parameters
    additionalParameters:
      key: bar
      value: baz
    ```

    In that way, someday we will be able to extend such approach with additional properties:
    ```yaml    
    instrumentation: # someday - if we want to unify the approach
      health:
        endpoint: foo.bar/healthz
      # (...)
    ```

1. Then, Content Developer specifies the Argo workflow artifact as output TypeInstance with familiar syntax:

    ```yaml
    # default - static
    capact-outputTypeInstances:
      - name: mattermost-config
        from: additional
        # no backend definition -> used default (default storage backend (TypeInstance) is annotated with `cap.core.attribute.default`)

    # option 2 - specific backend (referred in requires)
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
      additionalParameters:
        name: release-name
        namespace: release-namespace 
      backend: 3ef2e4ac-9070-4093-a3ce-142139fd4a16 # helm-release backend - resolved UUID based on the injected TypeInstance
    usesRelations: []
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

    1. Hub resolves details (URL) about the service
    1. Calls the registered storage backend service `onCreate` hook:

        ```javascript
        // TODO: Pseudocode, change to actual HTTP requests / ProtoBuf definition
        onCreate(typeInstanceID, additionalParameters?, value?): (additionalParameters?, error)
        ```

        This hook can mutate `additionalParameters`.

    1. Validate `additionalParameters` against JSON schema saved in the Backend Storage TypeInstance.
    1. Saves TypeInstance metadata in the core Hub storage backend, which contains all metadata of the TypeInstances and  theirs relations.

      ``` yaml
      id: # generated UUID
      typeRef:
        path: cap.core.type.hub.storage.redis
        revision: 0.1.0
      latestResourceVersion:
        resourceVersion: 1
        additionalParameters: # additional parameters that might be modified via the service handling `onCreate` hook
          name: release-name
          namespace: release-namespace 
      backend: 3ef2e4ac-9070-4093-a3ce-142139fd4a16 # helm-release backend - resolved UUID based on the injected TypeInstance
      ```

## Workflow syntax - Update

Similarly as with create, Content Developer specifies in the workflow:

```yaml
capact-updateTypeInstances:
- name: testUpdate
  from: update
```

where the `update` Argo artifact can contain `value` and / or `additionalParameters`.

As backend is immutable, we don't need to provide any additional syntax around TypeInstance update

## Storage Backend service implementation

1. The registered service will be called by Capact Hub and needs to handle following methods:

    TODO: Pseudocode, change to actual HTTP requests / ProtoBuf definition

    ```go
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

1. The service could be implemented using one of the following solutions (but it is not limited to):

  - [Dapr secrets](https://docs.dapr.io/developing-applications/building-blocks/secrets/secrets-overview/)
  - [Kubernetes external secrets](https://github.com/external-secrets/kubernetes-external-secrets)
  - [vault-k8s](https://github.com/hashicorp/vault-k8s)
  - [db](https://upper.io/v4/getting-started/)

## GraphQL API

The new GraphQL API can be used both on CLI and UI.

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

To see the value for all TypeInstances, query:

```graphql
query ListTypeInstances {
  typeInstances {
    id
    typeRef {
      path
      revision
    }
    lockedBy # resolver which calls proper backend storage service to ask for lock status
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
        value # resolver which calls proper backend storage service to ask for a given ResourceVersion value
        additionalParameters
      }
    }
    backend
  }
}
```

## Dynamic TypeInstance projections

**TODO:** Figure out how to solve problem with static TypeInstances like `mattermost-config`:
  - Support multiple different TypeInstances for a given projection
  - TypeInstance composition?

## Examples

**TODO:** Add some examples (apart from the one described above)

## Rejected ideas

#### Registering Storage Backends

1. Using Global / Action Policy to specify the default Storage Backend

    **Reason:** The Policy is already too big.

1. Dedicated entity of BackendStorage

    Such resource could reside in Local Hub, but it wouldn't be an OCF manifest. Cluster Admin should be able to manage them via GraphQL API, CLI and UI.

    **Reasons:**
    - We would still need some kind of BackendStorage templates (with `additionalParametersSchema` JSON schema) in public Hub
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

#### Backends implementation

1. Using Actions as a way to do CRUD operations (separate Interface/Implementation per Create/Update/Get/Delete operation)
 
    **Reason:** While the idea may seem exciting, that would be really time consuming and ineffective. We are too far from the point at where we can think about such solution. 

## Consequences

**TODO: Write consequences**
