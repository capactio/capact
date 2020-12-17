# Using Neo4j as OCH database with GoGM

This proof of concept shows GraphQL server with [GoGM library](https://github.com/mindstand/gogm) run against Neo4j database.
For GraphQL server implementation, the [gqlgen](https://gqlgen.com/) was used.

## Prerequisites
- Go
- Neo4J run on localhost, secured with `root` password for `neo4j` user

## Usage

To run the PoC, execute:

```bash
go run docs/investigation/graph-db/neo4j-gogm/main.go
```

Navigate to [http://localhost:3000/](http://localhost:3000/) and execute the following query:
```graphql
query {
    interfaceGroups {
        metadata {
            name
            description
            iconURL
        }
        interfaces {
            revisions {
                metadata {
                    name
                    displayName
                    description
                }
              
              	revision
            }
        }
    }
}
```

## Development

To regenerate the GraphQL schema, run:

```bash
# gqlgen
cd ./docs/investigation/graph-db/neo4j-gogm/graphql
gqlgen generate --verbose --config ./config.yaml
```

To regenerate functions to attach/detach references between types, run:

```bash
go get github.com/mindstand/gogm/cmd/gogmcli
gogmcli generate ./docs/investigation/graph-db/neo4j-gogm/graphql
```

## Summary

Because of the GoGM maturity issues, the PoC was ended earlier as we decided to not invest time in investigating it further. Read the subsections to learn more about the PoC outcome.

### Performance

As every nested struct is a different node in Graph DB, the GraphQL server for OCH would contain plenty of different nested resolvers. In worst case scenarios, a single complex query could run dozens of separate Database queries.

To solve the issue, we would need to use cache and build efficient queries. We could use [Dataloaders](https://gqlgen.com/reference/dataloaders/) or custom cache.

### Ease of use

GoGM use struct tags to map fields and relations. Every Go Type has to be updated with the struct tags for all fields. Every nested Go struct is represented as a separate node in graph. Even if a relation is directed, GoGM enforces developer to define pointers on both sides of the relations. For example, `InterfaceGroup` has to have pointer to `GenericMetadata`, and `GenericMetadata` has to have pointer to `InterfaceGroup`. See the [`models.go`](./graphql/models.go) file.

There is dedicated `gogmcli`, which generates methods to link nodes from both sides. In all examples the pointer for any relation from children to parent was always set (even if the direction of the relation was opposite). However, in this example it wasn't necessary to set pointers for relations on children, as the data was populated correctly.

### GoGM maturity

The GoGM library is not a mature project. It has 22 stars on GitHub, there is already 1.5.1 release, which could be considered as stable, but the reality is a bit different.

- There is almost no documentation apart from simple Create, Get By ID, List and Delete operations. See the [README.md](https://github.com/mindstand/gogm/blob/b8197657fba8056c48c53332c7bbe27b3a53958f/README.md). There is also [a separate repository with another example](https://github.com/mindstand/gogm-example/tree/b407f50d556ab752bb7b71b61c44565d14ad9a74), but it also doesn't showcase more than simple operations.
- GoGM returns enigmatic errors, which makes it hard to trace them. For example: `reflect.Value.Convert: value of type string cannot be converted to type *string` or ` errors found: (1) errors occurred: var name can not be empty -- total errors (1)`. There is [an issue](https://github.com/mindstand/gogm/issues/45) for that.
- GoGM doesn't handle pointer values properly while querying database. It can insert to database Go structs with pointer values, but it cannot query them. handle nullable GraphQL fields. These attempts fail with error such as `reflect.Value.Convert: value of type string cannot be converted to type *string`. Pointer values are very important for GraphQL server, as `gqlgen` use them to highlight that a given field is nullable in GraphQL.
- There is not a single example how to build more advanced operations, like listing nodes with filtering. For filter queries it uses a different library [go-cypherdsl](https://github.com/mindstand/go-cypherdsl), which contains **no documentation at all**. There are no examples and even test cases are not meaningful, which means it is hard understand how to use it properly. In this case it would be much easier to construct the Cypher queries by hand. See the [`schema.resolvers.go`](./graphql/schema.resolvers.go) file.

