# Dgraph as the OCH graph database

This proof of concept shows OCH server with implemented using [Dgraph v20.07](https://dgraph.io/docs/v20.07/).

## Motivation

We need to select a graph database for the OCH server which stores the OCF content and relations between each entity using edges.

### Goal

The main goal for this POC is to check whether we can use Dgraph seamlessly to expose the GraphQL client API without creating dedicated resolvers.

### Non-goal

Use the Dgraph only as of the database and create our own GraphQL server with dedicated resolvers' implementation for each entity.

## Prerequisite
-	Install Go
-	Install Docker
-	Install [Insomnia](https://insomnia.rest/download)

## Quick start
1.	Start Dgraph GraphQL server:

	```bash
	docker run --rm -it -p 8080:8080 -p 8000:8000 -p 9080:9080 dgraph/standalone:v20.07.2
	```

2.	Run database loader:

	```bash
	go run cmd/db-loader/main.go
	```

3.	Run custom resolver server:

	```bash
	go run cmd/resolver-svr/main.go
	```

4.	Import [Insomnia_localhost_OCH.json](./assets/Insomnia_localhost_OCH.json) into Insomnia and you are ready to execute sample queries.

### Simplifications
1.	The `interface.implementations` field uses resolver which is able to return Implementations only for the `latestResvision` property.
2.	The input filters for `interface.implementations` were not implemented as currently not possible.  
3.	The OCH content has additional properties and different names in manifests. By doing, so I didn't have to focus on mappers between OCF Entities and data store model. In normal scenario we should read the OCF entity then convert it to domain model, calculate edges and map to Dgraph data storage object.
4.	The OCH content is based on the mocked versions form [`hack/mock/graphql/public`](../../../../hack/mock/graphql/public).

### Behind the scene

The PoC has the following structure:

```
.
├── app 
│  └── cmd
│    ├── db-loader           # loader that is able to load GraphQL schema and entities from och-content
│    └── resolver-svr
└── assets
    ├── Insomnia_localhost_OCH.json
    ├── och-content                # simplified OCH content
    │  ├── implementation       # sample Implementations
    │  ├── interface            # sample Implementations
    │  └── type                 # sample Implementations
    ├── public-och-schema.graphql  # The GraphQL schema with Dgraph directives
    ├── public-och-schema.rdf      # The RDF schema
    └── schema.graphql
```

### Conditional upsert

We need to support situation when edges should always point to the latest revision. To ensure that state we can use [conditional upserts](https://dgraph.io/docs/v20.07/mutations/conditional-upsert/)

It is not possible to execute conditional upsert using GraphQL mutation. You need to use DQL. You can use RDF or JSON syntax. You cannot use own filter functions. Currently, supported functions are: [`eq/le/lt/ge/gt`](https://discuss.dgraph.io/t/would-like-support-of-eq-le-lt-ge-gt-in-mutation-conditional-upsert-other-than-existing-len-function-only/8846).

Check the `loadInterfaceRevisions` function from the [client/internal/interface_populator.go](client/internal/interface_populator.go) file to see how the conditional upsert can be done using Dgo client.

## Pros
-	Dgraph support GraphQL schema and expose GraphQL API out-of-the-box.
-	GraphQL requires that the type repeats all the fields from the interface, Dgraph doesn’t need that repetition in the input schema and will generate the correct GraphQL.

## Cons
-	The Ratel UI is simple, and it is not helpful during debugging.

-	There is only one [OGM](https://github.com/akshaydeo/dgogm) that has not been updated since 2017.

-	Custom queries needs to return whole object. We are not able to query only those fields which were requested.

-	We cannot have the input type for custom field queries:

	```
	cannot upload schema: response: {"errors":[{"message":"resolving updateGQLSchema failed because input:29: Type InterfaceRevision; Field customImplementations; @custom directive, body template must use fields defined within the type, found `input`.\n (Locations: [{Line: 3, Column: 4}])","extensions":{"code":"Error"}}]}
	```

	The `filters` keyword is reserved for queries

-	Using the @hasInverse filed in GraphQL Schema is not reflected in DQL

-	[Scalars do not work on custom queries](https://discuss.dgraph.io/t/a-scalar-type-was-returned-but-graphql-was-expecting-an-object/10908)

-	[There is no implementation for `for each` like statements](https://discuss.dgraph.io/t/foreach-func-in-dql-loops-in-bulk-upsert/5533/9)

-	By default, dgraph generates a lot of boilerplate. Each entity has own mutations/queries. We can disable that in the newest [version](https://dgraph.io/docs/master/graphql/schema/generate) which was not tested during this POC.

-	There is no `Any` scalar. As a result we need to use string type for **jsonSchema** property. More info [here](https://discuss.dgraph.io/t/json-blob-as-a-scalar/11034/7).

-	Dgraph supports RDF and GraphQL schemas but using only the RDF schema result in such error:

	```bash
	"Not resolving queryInterface. There's no GraphQL schema in Dgraph.  Use the /admin API to add a GraphQL schema"
	```

	It means that we need to always create the graphql schema if we want to use GraphQL API.

## Extras
-	[Compression](https://dgraph.io/docs/graphql/api/requests/#compression) is out of the box. Maybe we can use it at the beginning for cache sync.

-	Exclusive features like ACLs, binary backups, encryption at rest, and more: https://dgraph.io/docs/enterprise-features/

## Needs investigation
-	In the newest version they introduced Lambda fields which can help with writing custom resolvers. Unfortunately, it supports only JavaScript Lambdas and we will need to host our own lambda server for that.

-	Can facet help with query for Implementations that fulfil specific requirements?

-	How Dgraph maps the RDF entity to GraphQL types? Can we reuse that logic in our custom resolvers?

## Summary

TBD
