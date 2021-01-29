# OCH GraphQL API

Open Capability Hub can be run in two modes: public and local. In a result, GraphQL API for Open Capability Hub consists of two separate GraphQL schemas:
- [Public OCH](../och-js/graphql/public/schema.graphql)
- [Local OCH](../och-js/graphql/local/schema.graphql)

## Public API

Public OCH API contains GraphQL operations for the following entities:
- RepoMetadata
- InterfaceGroup
- Interface
- Type
- Implementation
- Attribute

Currently, there are no GraphQL mutations or subscriptions available. Once populated with DB populator, all resources are read-only.

To see full GraphQL schema, open the [`schema.graphql`](../och-js/graphql/public/schema.graphql) file.
 
## Local API

Local OCH API contains GraphQL operations for managing TypeInstances.

To see full GraphQL schema, open the [`schema.graphql`](../och-js/graphql/local/schema.graphql) file.

## Examples

To run sample GraphQL queries and mutations for Public or Local OCH, follow the steps:

1. Open the Voltron Gateway GraphQL Playground.

   To see how to access the Gateway on development cluster, read the [Access Gateway GraphQL Playground](./development.md#access-gateway-graphql-playground) section in development guide.

1. Navigate to a proper directory with GraphQL schema and examples.
   
   For Public OCH examples, navigate to [`och-js/graphql/public`](../och-js/graphql/public) directory.
   For Local OCH examples, navigate to [`och-js/graphql/local`](../och-js/graphql/local) directory.

1. Copy and paste the `examples.graphql` file content to the GraphQL Playground IDE.
1. Click on the "Query Variables" tab.
1. Copy and paste the `examples.variables.json` file content to the Query Variables section of the GraphQL Playground IDE.
1. Run any query or mutation from the list.

## Common flows

The following section showcases a few common usage flows for OCH GraphQL API.

### Querying different revisions

For all entities which has revision support, a unified API is available.
There are three nested resolvers:
- accessing the latest revision
- getting a specific revision
- getting all revisions

The following example shows the possibilities on Attribute entity:

```graphql
query {
    attributes {
        # name, path and prefix are immutable across all revisions of a given node.
        name # equal metadata.name, e.g. stateless
        path # equal to metadata.path, e.g. cap.core.attribute.workload.stateless
        prefix # equal to metadata.prefix, e.g. cap.core.attribute.workload

        # latest revision
        latestRevision {
            metadata {
                name
                displayName
            }
            revision
        }

        # given revision
        revision(revision: "1.0.0") {
            metadata {
                name
                displayName
            }
            revision
        }
        
        # all revisions
        revisions {
            metadata {
                name
                displayName
            }
            revision
        }
    }
}
```

The same unified API is available for all entities.

### List InterfaceGroups, that contains Interfaces, which contains Implementations for a given system 

To list all InterfaceGroups with nested Interfaces and Implementations for a given system that satisfy the Interfaces, you may use a single GraphQL query.

See the [`InterfaceGroupsWithInterfacesAndImplementations` sample query](../och-js/graphql/public/examples.graphql).

### Get Type details along the corresponding TypeInstances

Because of the OCH local and public API separation, currently a single GraphQL query for Types and corresponding TypeInstances is not possible.
To achieve that, the following queries have to be executed: 

1. Get Type details with `type` query to public OCH

   See the [`Type` sample query](../och-js/graphql/public/examples.graphql).
   
1. Get TypeInstances for a given Type with `typeInstances`

   See the [`ListTypeInstancesWithTypeRefFilter` sample query](../och-js/graphql/local/examples.graphql).

## Limitations

- For alpha release, to filter Implementations with the ones that are supported on a given system, UI always send TypeInstances list. In the future, there will be a dedicated query, where the available TypeInstances will be detected automatically and all Implementations will be filtered based on them.
- For alpha and GA release we don't support revision ranges.
