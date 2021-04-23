# Specification 0.0.1

The [schema](./schema) directory contains a representation of the OCF entities in the [JSON Schema draft-07](https://json-schema.org/draft-07/json-schema-release-notes.html) specification.

### RepoMetadata

RepoMetadata stores read-only information about the [Open Capability Hub](../../docs/e2e-architecture.md#och) (OCH), such as OCH version, supported OCF specification version, etc. This entity should be placed in the `core` directory in your OCH content. In the future, it will be embedded into the OCH server.

The RepoMetadata format is defined in [repo-metadata.json](./schema/repo-metadata.json).

> **NOTE:** Currently, the **spec.implementation** and **spec.ocfVersion.supported** properties are not supported by the OCH server.

### Attribute

Attribute provides an option to categorize [Implementations](#implementation), [Types](#type) and TypeInstances. For example, you can use **cap.core.attribute.workload.stateful** Attribute to find and filter Stateful Implementations.

The Attribute specification is defined in [attribute.json](./schema/attribute.json).

### Type

Type represents an object, such as database, application, but also a simple primitive data, like an IP address. The Type needs to be described using JSONSchema specification.

Type is used in [Interface](#interface) and [Implementation](#implementation) as a description and validation of possible input and output parameters. The object, which stores JSON value matching JSON schema from Type, is called TypeInstance.

The core Types are placed in the `core` directory. In the future, core Types will be embedded into the OCH server.

The [type-features.md](../../docs/type-features.md) describes all Type entity features.

The Type specification is defined in [type.json](./schema/type.json).

> **NOTE:** Currently, Type validation based on JSONSchema is not supported.

### Interface

Interface defines an action signature. It describes the action name, input, and output parameters. It is a similar concept that the one used in programming languages.

The Interface specification is defined in [interface.json](./schema/interface.json).

### InterfaceGroup

InterfaceGroup logically groups various Interfaces. This allows end-users to find all Interfaces, which for example operate on PostgreSQL instances. InterfaceGroup is required even if you have only one Interface.

The InterfaceGroup specification is defined in [interface-group.json](./schema/interface-group.json).

### Implementation

Implementation holds the definition of an [action](../../docs/terminology.md#action) and its prerequisites (dependencies). It allows you to define different ways on how a given action should be executed. For example, for **postgres.install** Interface, have **aws.postgresql.install** and **gcp.postgresql.install** Implementations. Implementation must implement at least one Interface.

The Implementation specification is defined in [implementation.json](./schema/implementation.json).

### Vendor

Vendor defines details of an external OCH server. This will be part of the OCH federation feature.

The Vendor specification is defined in [vendor.json](./schema/vendor.json).

> **NOTE:** Currently, it is not supported by the OCH server.
