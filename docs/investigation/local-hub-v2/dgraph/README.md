# Dgraph as the Local Hub

This is a second approach to Dgraph project after [the first investigation](../../graph-db/dgraph/README.md) done about a year ago.

## Motivation

We're looking for a good replacement to Neo4j, to resolve problems with high resource consumption and licensing. Ideally, we want to implement Local Hub in Go.

### Goal

The main goal is to check whether Dgraph can be used as a replacement for current Neo4j implementation to speed up Local Hub extensibility and reduce resource consumption.

Phases:
1. Check what was changed after a year.
2. Try to port existing Local Hub in minimal scope:
	- Query TypeInstances,
	- and TypInstances mutation with relation.
3. Use `@custom` or `@lambda` functionality to resolve delegated storage concern.
	- Instead of official [JavaScript](https://github.com/dgraph-io/dgraph-lambda) implementation, use [dgraph-lambda-go](https://github.com/Schartey/dgraph-lambda-go) server created by community.
4. If porting was possible, do the benchmarking between current and new solution.

### Non-goal

Even thought that, Public and Local Hubs share the same Neo4j instance, checking if Public Hub can be ported it's **not a goal** of this investigation.
Additionally, we don't want to use the Dgraph only as of the database and create our own GraphQL server with dedicated resolvers' implementation for each entity.

## Prerequisite

- Install Go
- Install Docker
- Install [Insomnia](https://insomnia.rest/download)

## What has changed?

In general, they mostly do a [bug fix releases](https://github.com/dgraph-io/dgraph/blob/12c3ef564cde11ecc3de96ec1516b3148e52d795/CHANGELOG.md).

## Overview

Dgraph doesn't allow adding custom logic directly in Dgraph server. To change the Dgraph behavior, you can:
- use `@custom` directive to redirect resolution of a given queries, mutations and fields. Resolution is done by calling the defined HTTP server.
- use `@lambda` directive to redirect resolution to lambda server. Dgraph provides lambda server only for JavaScript. In general, this is the same concept of executing HTTP calls against a registered Lambda server.

Dgraph is written in Go, but it cannot be directly extended in Go. Custom resolvers are always an HTTP call to your service.

It's due to the fact that Dgraph's GraphQL calls are not translated but directly executed on Dgraph server. In the official docs, you read:
> Dgraph is the world’s first, and at this time only, service to offer a specification-compliant GraphQL endpoint without the need for an additional translation layer in the tech stack. You will not find a hidden layer of GraphQL resolvers in Dgraph. GraphQL is natively executed within the core of Dgraph itself.

In comparison to current Neo4j implementation. We can do exactly the same, but we need to write delegation logic in JavaScript, as a result we can use any protocol we want to. Dgraph just gives a syntax sugar for extensions, but at the same time restrict it to HTTP only.

## Give it a try

1. Ensure that Docker is running on your localhost.
2. Run the Dgraph and lambda server and load GraphQL schema:
   ```bash
   docker compose up --build
   ```
   >**NOTE:** If the data was not populated successfully, run: `docker compose run populate`.
3. Use [Insomnia](https://insomnia.rest/download) to run all queries and mutations from [scenario.graphql](./assets/scenario.graphql) one by one.
   >**NOTE:** GraphQL address: http://localhost:8080

### Findings

After running the above example, you can notice that:

- The `@custom` or `@lambda` directives are only allowed on fields where the type definition has a field with type `ID!` or a field with `@id` directive.
- A type must have at least one field that is not of `ID!` type and doesn't have `@custom` or `@lambda` directive.
- In `@custom` and `@lambda` you have access only to primitive fields available on the root. In our case, we are not able to get the **backend** property when resolving `version`.
- If a field is marked with `@custom` or `@lambda` directive, it's not included in mutation.
- To prepare the demo scenario I had to check the Dgraph code directly as snippets described in Dgraph documentation are sometimes invalid.

At the beginning, I thought that lambdas on fields will be able to speed up our development. Due to, found issues, I don't think they will.

### Migration problems

I tried to migrate the current [Local Hub schema](../../../../hub-js/graphql/local/schema.graphql) to Dgraph and I spotted such issues:

- No option to add webhooks to protect the mutation. The current webhooks are just asynchronously triggered with a given event.

  **Why required?** This was needed to protect TypeInstance deletion when it's locked.

- No option to hide or rename generated query/mutation.

  **Why required?** Because of the above point, we will need to write a custom logic to protect the TypeInstance deletion when it's locked. To do so, I wanted to reuse automatically generated DQL instead of writing our own logic.

- [Custom DQL is still not supported on GraphQL mutation](https://discuss.dgraph.io/t/why-does-custom-dql-not-support-mutation/10366/12)

  **Why required?** We are not able to easily customize mutations as we can with Cypher in Neo4j.

- [We are not able to customize generated names for queries/mutations/filters.](https://discuss.dgraph.io/t/custom-names-for-crud-operations/13780/10)

  **Why required?** Without that, we are not able to keep backward compatibility. They suggest to use lambda for aliasing, but then GraphQL server will expose both variants. If we disable the default generation, then we need to write the DGL query by our own.

- [No option to define custom scalars.](https://discuss.dgraph.io/t/graphql-custom-scalar-types/16742)

  **Why required?** Without that, we won't have a readable API.

- [JSONSchema needs to be exposed as String.](https://discuss.dgraph.io/t/json-blob-as-a-scalar/11034/10)

  **Why required?** Without that, we are not able to keep backward compatibility.

- [No option to skip a given field on create](https://discuss.dgraph.io/t/graphql-error-non-nullable-field-was-not-present-in-result-from-dgraph/9503/6).

  **Why required?** We want to ignore the **lockedBy** property which is automatically generated for TypeInstance mutation. This field should be managed by other mutation but still visible in TypeInstance type for all queries. We could use `@lambda` resolver to overcome this issue. But this is more a hack than a proper solution.


Conclusion:
> Hmm… 😬 I started full-on with Dgraph’s GraphQL but have slowed down tremendously due to numerous feature requests/ bug fixes required for true production readiness. (...)

_source: https://discuss.dgraph.io/t/graphql-error-non-nullable-field-was-not-present-in-result-from-dgraph/9503/11_

## Summary

After a year, Dgraph didn't change too much. I failed to easily port current logic into Dgraph because of its limitations. In Dgraph we will need to implement almost all by our self. As a result, this won't speed up our current development.
Porting to Dgraph only makes sens when we want to reduce the resources and license issues. We need to be aware that there is no Cypher alternative in Dgraph.
