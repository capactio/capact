# Dgraph as OCH graph database

### Criteria

Checklist:
-	[ ] Can we use just one graphql schema for storing and querying object
-	[ ] How to expose the GraphQL Playground
-	[ ] Is there an output schema that can be consumed by GraphQL playground
-	[ ] What is under the hood, is it also using [gql-gen](https://github.com/99designs/gqlgen) lib for generating types
-	[ ] How to write custom queries, mutations, scalars, types
-	[ ] Is there some directive for validation (e.g. checking SemVer syntax)
-	[ ] Resources consumption (run 1k queries/mutations)
-	[ ] Check transactions system
-	[ ] Check [type system](https://dgraph.io/docs/query-language/type-system/)
-	[ ] Check go client, execute some queries using Go client and some using GraphQL Playground
-	[ ] Check subscription


OCF model -> domain model do the resolving logic -> DSO dgraph with all necessary edges etc.


## Quick start
-	Start Dgraph GraphQL server

	```bash
	$ docker run -it -p 8080:8080 dgraph/standalone:master
	...
	```
1.	Add a GraphQL Schema

	```bash
	$ curl -X POST localhost:8080/admin/schema --data-binary '@assets/public-och-schema.graphql'
	{"data":{"code":"Success","message":"Done"}}
	```

## Conditional upsert

It is not possible to execute using GraphQL mutation. You need to use DQL. You can use RDF or JSON syntax. You cannot use own filter functions. Currently, supported [`eq/le/lt/ge/gt`](https://discuss.dgraph.io/t/would-like-support-of-eq-le-lt-ge-gt-in-mutation-conditional-upsert-other-than-existing-len-function-only/8846).

#### Using string query:

```
upsert {
  query {
    u1 as inter(func: uid("0x5")) @filter(type(Interface)) @cascade {
        Interface.latestRevision @filter( lt(InterfaceRevision.revision, "0.1.3")){
         uid
        }
      }
  }

  mutation @if(not(eq(val(existingLines), "${lines}"))) {
    set {
      uid(fileId) <lines> "${lines}" .
    }
  }
}
```

- show all nodes 
```


{
  showallnodes(func: has(dgraph.type)){
    dgraph.type
    expand(_all_) {
      expand(_all_)
    }
  }
}
```

#### Using Dgo client

Check the `loadInterfaceRevisions` function from the [client/internal/interface_populator.go](client/internal/interface_populator.go) file.

## Pros
-	GraphQL requires that the type repeats all the fields from the interface, Dgraph doesn’t need that repetition in the input schema and will generate the correct GraphQL.

## Cons

- probably cannot have the input type for custom field queries: 
```
 cannot upload schema: response: {"errors":[{"message":"resolving updateGQLSchema failed because input:29: Type InterfaceRevision; Field customImplementations; @custom directive, body template must use fields defined within the type, found `input`.\n (Locations: [{Line: 3, Column: 4}])","extensions":{"code":"Error"}}]}

```
- cannot return error from custom queries (https://discuss.dgraph.io/t/remote-mutations-queries-how-to-return-errors/8696), it is possible (https://github.com/dgraph-io/dgraph/pull/6604)
-	Using the @hasInverse filed in GraphQL Schema is not reflected in DQL 

-	scalars do not work on custom queries: https://discuss.dgraph.io/t/a-scalar-type-was-returned-but-graphql-was-expecting-an-object/10908

-	No for each yet: https://discuss.dgraph.io/t/foreach-func-in-dql-loops-in-bulk-upsert/5533/9

-	`implementations(filter: CustomImplementationFilter): [Implementation!]!` transformed to `implementations: [Implementation!]`. Filters are added based on directive and it cannot be required as we will not be able to add interfaces.

-	some sections are not finished ![docs](assets/dgraph-missing-docs.png)

-	generates a lot of boilerplate. Each type has own mutations/queries. We can disable that but it is available on master: https://dgraph.io/docs/master/graphql/schema/generate

-	There is no `Any` scalar. How to replace that for **jsonSchema** property? https://discuss.dgraph.io/t/json-blob-as-a-scalar/11034/7

-	Support RDF and GraphQL Schemas but using only the RDF schema result in such error:

	```
	"Not resolving queryInterface. There's no GraphQL schema in Dgraph.  Use the /admin API to add a GraphQL schema"
	```

	It means that we need to always create the graphql schema

	## Extras

-	Populate database using RDF or JSON. Dgraph as of now supports mutation for two kinds of data: RDF and JSON.

-	Compression is out of the box: https://dgraph.io/docs/graphql/api/requests/#compression. Maybe we can use it at the beginning for cache sync?

-	Maybe use cascade to get implementation: https://dgraph.io/docs/graphql/queries/cascade. Use it to do not return interfaces without implementation.

-	To run with locally saved data:

	```bash
	docker run --rm -it -p 8080:8080 -p 9080:9080 -p 8000:8000 -v ~/dgraph:/dgraph dgraph/standalone:v20.03.0
	```

-	Exclusive features like ACLs, binary backups, encryption at rest, and more: https://dgraph.io/docs/enterprise-features/


- Additionally, there is no support for slice input: https://discuss.dgraph.io/t/support-lists-in-query-variables-dgraphs-graphql-variable/8758
I didnt change the schema too much but adding some new edges when injecting the data will help a lot e.g. adding relations between `implementations.requires.typeRef` and `types`. 

IMO we should flat the Metadata and be able to index by path
