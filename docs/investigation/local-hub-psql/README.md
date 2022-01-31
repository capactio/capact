# Local Hub with PostgreSQL implementation

This document aggregates raw notes from a short investigation as a part of the [#626](https://github.com/capactio/capact/issues/626) issue.

## Graph solutions for PostgreSQL

In PostgreSQL it may be time-consuming to create graph relation for TypeInstances by our own. See:

  - https://stackoverflow.com/questions/20776718/how-to-model-graph-data-in-postgresql
  - https://patshaughnessy.net/2017/12/11/trying-to-represent-a-tree-structure-using-postgres
  - https://www.bustawin.com/dags-with-materialized-paths-using-postgres-ltree/

To avoid implementing and maintaining our own solution, we could save time and use some existing graph solution for PostgreSQL.

#### [AgensGraph](https://github.com/bitnine-oss/agensgraph)

- PostgreSQL fork with OpenCypher and SQL queries support
    - You can even mix them in a single query
- Missing support for apoc; we would still need to rewrite our queries
- [Golang driver](https://github.com/bitnine-oss/agensgraph-golang) seems to be abandoned - no commits from over 3 years
- [NodeJS driver]((https://github.com/bitnine-oss/agensgraph-nodejs)) seems to be at an early stage, without any commit for more than 9 months
- Last commit on master was around 4 months ago
- We cannot treat it as a replacement of Neo4j and simply use different DB with the [`neo4j-graphql`](https://github.com/neo4j/graphql) library, as it uses different driver with different protocol (bolt).

#### [Apache AGE](https://github.com/apache/incubator-age)

- PostgreSQL extension with OpenCypher implementation 
- Missing support for apoc; we would still need to rewrite our queries
- Inspired by AgensGraph
- PostgreSQL 11 only support (until 2022 Q1, as the docs say)
- Drivers:
  - [Official Golang driver](https://github.com/apache/incubator-age/tree/master/drivers/golang) seems to be at a very early stage
  - [Unofficial Golang driver](https://github.com/rhizome-ai/apache-age-go) is at a very early stage (`v.0.0.4`), created by a single person, and can't be considered right now 
- [NodeJS client](https://github.com/apache/incubator-age/tree/master/drivers/nodejs) is very simple, introduced around a year ago with [not many changes over time](https://github.com/apache/incubator-age/commits/master/drivers/nodejs)
- We cannot treat it as a replacement of Neo4j and simply use different DB with the [`neo4j-graphql`](https://github.com/neo4j/graphql) library, as it uses different driver with different protocol (bolt).

This is more actively developed than [AgensGraph](#agensgraphhttpsgithubcombitnine-ossagensgraph), however it is at a very early stage.


## GraphQL server implementation

Assuming that we know how we want to represent graph in PostgreSQL, we still need to find an effective way how to bootstrap GraphQL server on top of it. Here are some of the solutions I checked:

## Go

There are not many solutions when we would like to use Go as the language for our Local Hub:

### [GraphJin](https://github.com/dosco/graphjin)

- written in Go (but it doesn't help much - read on)
- You can use YAML config or Go struct config if you build your custom Docker image
- write custom query resolvers as [SQL functions](https://github.com/dosco/graphjin/wiki/Guide-to-GraphQL#custom-functions)
- generates GraphQL API which is not user friendly. For example insert/update/patch/delete operations is actually one mutation:

   ```graphql
   mutation {
     products(update: $data, where: { id: { eq: 12 } }) {
       id
       name
     }
   }     
   ```

  Read more in the [docs](https://github.com/dosco/graphjin/wiki/Guide-to-GraphQL)

- you can't customize generated resolvers
    - no way to customize names and behavior of the queries and mutations
- custom resolvers which call external REST APIs (missing documentation, the only mention is in [Readme](https://github.com/dosco/graphjin/wiki#features))
- adding business logic without external server with JS is possible, but it works not like we want to: [client must specify special directive to run a given JS script](https://github.com/dosco/graphjin/wiki/Guide-to-GraphQL#adding-business-logic-with-javascript)
- you can insert/update nested properties
  - this is not well documented (missing DB schemas)
  - it breaks GraphQL contract: introspection is not available, this is a custom logic beyond GraphQL
  - See [nested insert](https://github.com/dosco/graphjin/wiki/Guide-to-GraphQL#nested-insert) and [nested update](https://github.com/dosco/graphjin/wiki/Guide-to-GraphQL#nested-update)
  - See the PoC to see how counterintuitive it is
- high level of entry: missing documentation (e.g. custom resolvers, table column mapping)

### Node.JS

#### [PostGraphile](https://github.com/graphile/postgraphile)

- Quite user-friendly GraphQL API generated from DB schema - for example, see [CRUD mutations](https://www.graphile.org/postgraphile/crud-mutations/)
- custom [queries](https://www.graphile.org/postgraphile/custom-queries/) and [mutations](https://www.graphile.org/postgraphile/custom-mutations/) with SQL functions
- ability to customize fields mapping with ["smart comments"](https://www.graphile.org/postgraphile/smart-comments/)
- ability to have mutation to create related entities with a [separate plugin](https://github.com/mlipscombe/postgraphile-plugin-nested-mutations)

I didn't test it by myself, but it seems like the best option, comparing to [Hasura](#hasurahttpsgithubcomhasuragraphql-engine) and [Graphjin](#graphjinhttpsgithubcomdoscographjin) - see the [thread on hacker news](https://news.ycombinator.com/item?id=22433478).
However, I believe the GraphQL-first solution would be better for us than database-schema-first one.

#### [Hasura](https://github.com/hasura/graphql-engine)

- [GraphQL schema customization by schema API and UI](https://hasura.io/docs/latest/graphql/core/api-reference/schema-api/index.html)
- Custom resolvers - only with the following approach:
  - [remote services (webhooks)](https://hasura.io/docs/latest/graphql/core/actions/index.html)
  - [remote GraphQL schemas (GraphQL services)](https://hasura.io/docs/latest/graphql/core/remote-schemas/index.html)
- No plugins
- Works as a separate service

This solution doesn't really fit into our use case. We would like to have custom logic in our resolvers without any delegation to a separate service.

#### [JoinMonster](https://github.com/join-monster/join-monster)

- A library which translates GraphQL calls to SQL
- Basically it is a set of helpers to make DB queries easy. It is a layer on top of the `graphql` library, and you specify mapping between GraphQL and SQL tables.
- You can provide custom SQL queries
- If we want to have schema-first approach, we need to use a [`graphql-tools` adapter](https://github.com/join-monster/join-monster-graphql-tools-adapter)
- Ability to customize all resolvers in JavaScript

This solution looks promising, as it simplifies the boilerplate we need to write in order to translate GraphQL to SQL. It would definitely speed up the development. However, we would still need to implement our simple graph solution.

## Implementation ideas

### Own simple graph implementation in PostgreSQL

- Implement Local Hub in Go with [gqlgen](https://github.com/99designs/gqlgen), [gorm](https://github.com/go-gorm/gorm) and our own graph implementation
- Implement Local Hub in JavaScript with [JoinMonster](#joinmonsterhttpsgithubcomjoin-monsterjoin-monster) and our own graph implementation

We can use [ltree](https://www.postgresql.org/docs/current/ltree.html) extension. However, in general, implementing and maintaining our own graph solution doesn't seem like the most efficient approach. Also, IMHO it won't scale well when we would like to rewrite our Public Hub as well.

### Use existing graph implementation for PostgreSQL 

- Implement Local Hub in Go with AgensGraph, using official, abandoned Go client, [gorm](https://github.com/go-gorm/gorm) and our own custom OpenCypher queries

The abandoned Go client could be problematic in near future; we can't use Apache AGE, as it's on too early stage

### Others

- Continue the investigation and reevaluate [dgraph](https://github.com/dgraph-io/dgraph)

### Decision

We decided to reevaluate `dgraph` first. If `dgraph` suits our needs, we will replace the graph database in Local Hub.
If not, then we will keep the Local Hub implemented in Node.js with Neo4j database for the time being, and implement [Delegated Storage](../../proposal/20211207-delegated-storage.md) as an extension for the current codebase. Then, we will revisit the Local and Public Hub rewrite once again. 
