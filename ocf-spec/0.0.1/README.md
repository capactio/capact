# Schema 0.0.1

The [schema](./schema) directory contains a representation of the OCF entities.

### Attribute

### Implementation

### Interface

### RepoMetadata

Stores read-only metadata about the OCH.

### Type

### Vendor

Vendor defines an external OCH repository. The name of a vendor object must be a valid [DNS subdomain name](https://kubernetes.io/docs/concepts/overview/working-with-objects/names#dns-subdomain-names). A strategy of resolving nested resources will be used initially by our OCF SDK/CLI. In the future, we may introduce caching of vendors.
