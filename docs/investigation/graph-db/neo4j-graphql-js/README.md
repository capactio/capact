# Using Neo4j as OCH database with `neo4j-graphql.js`

This proof of concept presents GraphQL server implemented
with [`neo4j-graphql.js`](https://github.com/neo4j-graphql/neo4j-graphql-js) that uses Neo4j database.

## Requirements

- Node 14.15.2
- Neo4J 4.1.3 database
    - run on `localhost:7687`
    - with [APOC](https://neo4j.com/labs/apoc/4.1/installation) plugin installed
    - secured with `neo4j` user and `root` password

## Usage

Install dependencies:

```bash
npm install
```

Run the server:

```bash
npm start
```

Navigate to [http://localhost:3000/](http://localhost:3000/)

### Upload sample data

Run the mutation from the [`seed.graphql`](seed/seed.graphql) file, using variables from
the [`variables.json`](seed/variables.json) file.

### Run queries

Run the queries from the [`sample-query.graphql`](graphql/sample-query.graphql) file.

### Pros

The `neo4j-graphl-js` contains multiple features which makes it easier to develop a GraphQL server based on Neo4j
database.

#### Resolver generation

- It generates queries and mutations automatically based on GraphQL schema with custom directives. The `@relation`
  directive is required to specify relations between different nodes.

- If you have a query or mutation already specified for a given GraphQL type, then library detects it, and, instead of
  generating additional query/mutation, it injects filter input parameters to the existing query or mutation.

    - You can configure the behavior:
        - turn off query and/or mutation generation
        - exclude specific types from query or mutation generation

  See more [here](https://grandstack.io/docs/neo4j-graphql-js-api#config).

- It generates automatic filters for queries for every property of a given type, and even type one depth level below.
  Read more [here](https://grandstack.io/docs/graphql-filtering).

  There is also
  an [Experimental API](https://grandstack.io/docs/graphql-schema-generation-augmentation/#experimental-api) for
  enabling advanced queries during create/update/upsert/delete/relationship operations.

- For every query that returns multiple items, the library generates pagination and ordering input arguments.

- It handles scalars properly, as well as spatial types, Unions, Interfaces.

#### Custom resolvers

You can prepare custom resolvers in two ways:

- by using custom database queries, defined straight in GraphQL schema with `@cypher` directive. They are still
  optimized and merged into a single database query:

  From [docs](https://grandstack.io/docs/graphql-custom-logic#computed-scalar-fields):
  > The generated Cypher query includes the annotated Cypher query as a sub-query, preserving the single database call to resolve the GraphQL request.

  The queries handles input arguments (even when they are complex input types)

- by defining custom JavaScript resolvers.

  For root resolvers, that is queries and mutations, you can use
  the [`neo4jgraphql`](https://grandstack.io/docs/neo4j-graphql-js-api#neo4jgraphqlobject-params-context-resolveinfo-debug-executionresult)
  helper, which generate a single database query for all data the user asked for. In that way you can e.g. process input
  data beforehand, or easily rename generated queries or mutations, still using all the benefits from the library.

#### Other features

- JWT scope-based authorization is
  available [out of the box](https://grandstack.io/docs/neo4j-graphql-js-middleware-authorization).
- [Support for multiple Neo4j databases](https://grandstack.io/docs/neo4j-multiple-database-graphql)
- [Support of Apollo Federation and Gateway](https://grandstack.io/docs/apollo-federation-gateway-with-neo4j)

#### Maturity

The `neo4j-graphql-js` is a mature project from the Neo4j GraphQL organization, which gathers all GraphQL integrations
and tools for Neo4j. It is actively developed as a part of [GRANDstack](https://grandstack.io/).
Its [documentation](https://grandstack.io/docs/neo4j-graphql-js-quickstart) is well-written and it contains multiple
examples.

When it comes to the database, the Neo4j database has the biggest community in graph databases ecosystem. The tooling
around the database, such as Desktop client with IDE, are polished and very helpful in development. Cypher query
language is SQL-like language to build all databases queries. After some short tutorial it seems to be intuitive, if
someone is already familiar with SQL.

### Cons

#### Custom JavaScript nested resolvers limitations

Even though `neo4j-graphql-js` is a flexible solution, during this PoC a single limitation has been observed. It is
related to custom nested resolvers.

As it was previously stated in [Pros](#pros) section, there is an ability to define custom Cypher queries for some
fields, queries or mutations. Also, there is an option to create own JavaScript resolvers.

If you don't want to use only Cypher for resolving logic, you have to specify JavaScript resolver. If this is a nested
resolver, you cannot use
the [`neo4jgraphql` helper function](https://grandstack.io/docs/neo4j-graphql-js#translate-graphql-to-cypher). This
helper function can be only used for root queries or
mutations [See the issue #390](https://github.com/neo4j-graphql/neo4j-graphql-js/issues/390).

If you still want to query database in custom JavaScript resolvers, then you need to query the database manually - see
the [`resolvers.mjs`](./resolvers.mjs) file for implementation reference. If the nested JavaScript resolver is a leave (
it doesn't contain any properties with related entities), then there is no issue. The issue is if you want to query some
nested properties for related objects. To handle them properly, the custom JavaScript resolver needs to return parent
object with all nested (related) objects.

See the `InterfaceRevision.implementationRevisionsForRequirementsCustom` nested resolver implementation
in [`resolvers.mjs`](./resolvers.mjs) and [`schema.graphql`](./graphql/schema.graphql). If you try to query nested
fields of the `ImplementationRevision`, the GraphQL server will return error, because the result of nested resolvers
are `undefined`.

The reason of that is the `neo4j-graphql-js` library works by not using custom resolvers at all. It generates the
database query on root resolver (for mutation or query) during server start. In a result, if you write a custom
JavaScript resolver for  `InterfaceRevision.implementationRevisionsForRequirementsCustom`, you would need to return all
the data for related object, which results in overfetching.

To avoid overfetching, there is one solution. We access `selectionSet` in `resolveInfo` in a resolver to get queried
fields and then build a final query, similarly to what the library does. However, in most cases, using `@cypher`
directive should be enough to implement some custom database-related logic.

#### Limited filtering functionality for scalar inputs

To have filtering parameters based on string value, for example, for `path`: `path_in`, `path_not`, `path_starts_with`
etc., the field has to be `ID` or `String` type. For scalars there is only strict equal filtering out of the box.

#### Technologies

- The `neo4j-graphql-js` library would enforce us to write OCH in TypeScript/JavaScript.

- Neo4j is implemented in Java, which consumes more resources than, for example, `dgraph`. During the PoC development, a
  single database instance consumed about 450MB of RAM memory on development machine.

### Other observations

- Custom DB queries, defined with `@cypher` directive directly on GraphQL schema,
  requires [APOC](https://github.com/neo4j-contrib/neo4j-apoc-procedures) module installed on the database. It brings
  many [additional features](https://neo4j.com/labs/apoc/4.1/overview/) of the database queries, which we could also
  utilize in some cases.

- In our use case, TypeInstances could be saved as separate nodes, and then read with recursive query such as:
  ```
  MATCH (i:TypeInstance {id: "...")-[:COMPOSES_OF *0..]->(i2:TypeInstance) RETURN i, i2
  ```

  This query will return all TypeInstance nodes that the TypeInstance parent composes of.

  Splitting TypeInstances during save or merging them during read has to be done manually, on client side.

  Additional resources:
    - [Neo4j documentation](https://neo4j.com/docs/cypher-manual/current/clauses/match/#relationships-in-depth)
    - [Getting all child nodes recursively](https://stackoverflow.com/questions/45359581/get-all-child-nodes-recursive)
    - [Simple recursive Cypher query](https://stackoverflow.com/questions/31079881/simple-recursive-cypher-query)

## Summary

The `neo4j-graphql-js` gives us ability to implement OCH fast, without custom logic apart from custom database queries.
Considering Neo4j with the library, the solution seems to be flexible enough to cover all our cases.

The only real downside is the limited support for custom JavaScript resolvers. However, most cases could be handled with
the `@cypher` directive, as the Cypher language is mature and contains multiple features. The APOC library extend them
further with additional capabilities. 
