# Using Grakn database in OCH with Go client generated from protobuf definitions.

This proof of concept shows [Grakn](https://grakn.ai/) server as Graphql backend for OCH.
For GraphQL server implementation, the [gqlgen](https://gqlgen.com/) was used.

## Prerequisites
- Go
- Grakn server run on localhost. 

## Usage

To run the PoC, load data into grakn:

```bash
grakn console --keyspace och --file data/schema.gql
grakn console --keyspace och --file data/data.gql
```

Before building binary you need to generate go client. See description in [go-grakn/README.md](./go-grakn/README.md)

```bash
go build
```

and run it:

```bash
./grakn
```

Navigate to [http://localhost:3000/](http://localhost:3000/) and execute the following query:
```graphql
{
  interfaceGroups {
    interfaces {
      revisions {
        implementations {
          name
        }
      }
    }
  }
}

```

## Development

To regenerate the GraphQL schema, run:

```bash
# gqlgen
cd ./docs/investigation/graph-db/grakn/graphql
gqlgen generate --verbose --config ./config.yaml
```

## Summary

Because of the lack of high level go client, no ogm, no graphql support we decided to stop researching this database.

Grakn is using grpc for communication with database. There is no go-client available. We can generate one but it's very low level.
We would need to maintain higher level go client. This is time consuming. There will be also soon released a new Grakn version which has incompatible API so we would need to start from the beginning.
With generated client there is no easy way to map graphql schema to Grakn types. This could also have performance impact but it was not tested.
