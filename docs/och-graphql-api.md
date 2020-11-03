# OCH GraphQL API

Open Capability Hub can be run in two modes: public and local. In a result, GraphQL API for Open Capability Hub consists of two separate GraphQL schemas:
- [Public OCH](../pkg/och/api/public/schema.graphql)
- [Local OCH](../pkg/och/api/local/schema.graphql)

## Public API

Public OCH API contains GraphQL operations for the following entities:
- RepoMetadata
- InterfaceGroup
- Interface
- Type
- Implementation
- Tag

Currently, there are no GraphQL mutations or subscriptions available. Once populated with DB populator, all resources are read-only.

To see full GraphQL schema, open the [`schema.graphql`](../pkg/och/api/public/schema.graphql) file.
 
## Local API

Local OCH API contains GraphQL operations for managing TypeInstances.

To see full GraphQL schema, open the [`schema.graphql`](../pkg/och/api/local/schema.graphql) file.

## Examples

The following section showcases a few examples.

### Querying different revisions

For every entity which has revision support, three queries are available:
- getting the latest revision content
- getting a specific revision content
- getting all revisions content

The following example shows the possibilities on Tag entity:

```graphql
query {
    tags {
        # name, path and prefix are immutable across all revisions of a given node.
        name # equal metadata.name, e.g. stateless
        path # equal to metadata.path, e.g. cap.core.tag.workload.stateless
        prefix # equal to metadata.prefix, e.g. cap.core.tag.workload

        # latest revision
        latestRevision {
            metadata {
                name
                displayName
            }
            revision
            spec {
                additionalRefs
            }
        }

        # given revision
        revision(revision: "1.0.0") {
            metadata {
                name
                displayName
            }
            revision
            spec {
                additionalRefs
            }
        }
        
        # all revisions
        revisions {
            metadata {
                name
                displayName
            }
            revision
            spec {
                additionalRefs
            }
        }
    }
}
```

### List InterfaceGroups, that contains Interfaces, which contains Implementations for a given system 

```graphql
query {
    interfaceGroups {
        metadata {
            name
            description
            iconURL
        }
        interfaces {
            latestRevision {
                metadata {
                    name
                    displayName
                    description
                }
                spec {
                    input
                    output
                }
                implementations(filter: {
                    tags: [{path: "cap.tag.foo.bar", rule: INCLUDE}],
                    requirementsSatisfiedBy: [
                        {
                            typeRef: {
                                path: "cap.core.type.platform.kubernetes", revision: "1.0.1",},
                            value: "{\"version\": \"1.18.9\"}"
                        },
                        {
                            typeRef: {
                                path: "cap.type.database.mysql.config", revision: "1.0.1",},
                            # if value is not provided, all value-related constraints are treated as satisfied.
                        }
                    ]
                }) {
                    name
                    latestRevision {
                        metadata {
                            displayName
                        }
                        spec {
                            action
                            appVersion
                        }
                    }
                }
            }
        }
    }
}
```

### Get Type details along the corresponding TypeInstances

Because of the OCH local and public API separation, currently a single GraphQL query for Types and corresponding TypeInstances is not possible.
To achieve that, the following queries have to be executed: 

1. Get Type details

    ```graphql
    query {
        type(path: "cap.core.type.platform.kubernetes") {
            name 
            path
            prefix
            latestRevision {
                metadata {
                    name
                    displayName
                    description
                }
                revision
                spec {
                    additionalRefs
                    jsonSchema
                }
            }
        }
    }
    ```

1. Get TypeInstances for a given Type

    ```graphql
    query {
        typeInstances(filter: {
            typeRef: {path: "core.type.platform.kubernetes" } # no revision means that the latest revision is picked
        }) {
            metadata {
                id
            }
            resourceVersion
            spec {
                value
                instrumentation {
                    health {
                        status
                    }
                    metrics {
                        dashboards {
                            url
                        }
                        endpoint
                        regex
                    }
                }}
        }
    }
    ```

## Limitations

- For alpha release, to filter implementations with the ones that are supported on a given system, UI always send TypeInstances list. In future, there will be a dedicated query, where the available TypeInstances will be detected automatically and all Implementations will be filtered based on them.
- For alpha and GA release we don't support revision ranges.
