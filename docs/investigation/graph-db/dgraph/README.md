# Dgraph as the OCH graph database

This proof of concept shows the OCH server implemented using [Dgraph](https://dgraph.io) v20.07.

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

## Quickstart
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
3.	The OCH content has additional properties and different names in manifests. By doing, so I didn't have to focus on mappers between OCF Entities and the data store model. In the normal scenario, we should read the OCF entity then convert it to the domain model, calculate edges and map to Dgraph data storage object.
4.	The OCH content is based on the mocked versions form [`hack/mock/graphql/public`](https://github.com/capactio/capact/tree/release-0.1/hack/mock/graphql/public).

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
    │  ├── interface            # sample Interfaces
    │  └── type                 # sample Types
    ├── public-och-schema.graphql  # The GraphQL schema with Dgraph directives
    ├── public-och-schema.rdf      # The RDF schema
    └── schema.graphql
```

### Conditional upsert

We need to support situations when edges should always point to the latest revision. To ensure that state we can use [conditional upserts](https://github.com/dgraph-io/dgraph-docs/blob/release/v20.07/content/mutations/conditional-upsert.md)

It is not possible to execute conditional upsert using GraphQL mutation. You need to use DQL. You can use RDF or JSON syntax. Furthermore, you cannot use your own filter functions. Currently, supported functions are: [`eq/le/lt/ge/gt`](https://discuss.dgraph.io/t/would-like-support-of-eq-le-lt-ge-gt-in-mutation-conditional-upsert-other-than-existing-len-function-only/8846).

Check the `loadInterfaceRevisions` function from the [client/internal/interface_populator.go](app/internal/interface_populator.go) file to see how the conditional upsert can be done using Dgo client.

## Pros
-	Dgraph support GraphQL schema and expose GraphQL API out-of-the-box.
-	GraphQL requires that the type repeats all the fields from the interface, Dgraph doesn’t need that repetition in the input schema and will generate the correct GraphQL.
-	For each query we have dedicated filters out of the box.
-	For each query we have pagination.

## Cons

I sorted the problems that I faced during the investigation. The **HARD** category means that problems listed there have a high priority and implementation can be time-consuming and the **LOW** means that issues listed there have a quick workaround which should be fine even for GA.

#### HARD:
-	Custom queries need to return the whole object. We are not able to query only those fields which were requested by the user. [I described that problem on their forum](https://discuss.dgraph.io/t/custom-dql-resolver-for-field-define-in-graphql-schema/11934/2).

#### MEDIUM
-	We cannot have the input type for the custom field queries:

	```graphql
	type InterfaceRevision {
	    id: ID!

	    # THIS input ImplementationFilter is not allowed
	    implementations(filter: ImplementationFilter): [Implementation!] @custom(http: {
	      url: "http://example.com"
	      method: POST
	     body: "{id:$id, filter:$filter}"
	    })
	}

	input ImplementationFilter {
	  requirementsSatisfiedBy: [Requirement]
	}
	```

	Applying this schema, results in such error:

	```
	cannot upload schema: response: {"errors":[{"message":"resolving updateGQLSchema failed because input:29: Type InterfaceRevision; Field customImplementations; @custom directive, body template must use fields defined within the type, found `input`.\n (Locations: [{Line: 3, Column: 4}])","extensions":{"code":"Error"}}]}
	```

	Additionally, the `filters` keyword is reserved for queries

	Probably we will need to use the String type and the user will be responsible for marshaling a JSON input.

-	There is no `Any` scalar, and we also cannot create our own scalar types. As a result, we need to use string type for **jsonSchema** property. Discussion about supporting JSON types is in progress but there is no information if and when this will be implemented. More info [here](https://discuss.dgraph.io/t/json-blob-as-a-scalar/11034/7).

-	There is no option to override generated queries and mutation names.

-	The DQL queries return entities in RDF format. This needs to be later mapped to GraphQL. For POC purposes I decided to use the regexp to find and remove those prefixes.

	**DQL response**

	```
	[
	  {
	    "Implementation.name": "install",
	    "Implementation.prefix": "cap.implementation.atlassian.jira",
	    "Implementation.path": "cap.implementation.atlassian.jira.install",
	    "Implementation.latestRevision": {
	      "ImplementationRevision.metadata": {
	        "path": "cap.implementation.atlassian.jira.install",
	        "displayName": "Install Jira",
	        "description": "Action which installs Jira via Helm chart",
	        "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
	        "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png",
	        "name": "install",
	        "prefix": "cap.implementation.atlassian.jira",
	        "maintainers": [
	          {
	            "Maintainer.name": "Capact Dev Team",
	            "Maintainer.email": "team-dev@capact.io",
	            "Maintainer.url": "https://capact.io"
	          }
	        ],
	        "supportURL": " https://mox.sh/helm"
	      },
	      // ...
	    }
	]
	```

	**GraphQL response**

	```
	[
	  {
	    "path": "cap.implementation.atlassian.jira.install",
	    "latestRevision": {
	      "metadata": {
	        "name": "install",
	        "prefix": "cap.implementation.atlassian.jira",
	        "path": "cap.implementation.atlassian.jira.install",
	        "description": "Action which installs Jira via Helm chart",
	        "documentationURL": "https://github.com/javimox/helm-charts/tree/master/charts/jira-software",
	        "supportURL": " https://mox.sh/helm",
	        "displayName": "Install Jira",
	        "maintainers": [
	          {
	            "name": "Capact Dev Team",
	            "email": "team-dev@capact.io",
	            "url": "https://capact.io"
	          }
	        ],
	        "iconURL": "https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png"
	      },
	    // ...
	  }
	]
	```

#### LOW:
-	The Ratel UI is simple, and it is not helpful during debugging.

-	There is only one [OGM](https://github.com/akshaydeo/dgogm) that has not been updated since 2017.

-	Using the @hasInverse filed in GraphQL Schema is not reflected in DQL

-	[Scalar types cannot be returned from custom queries](https://discuss.dgraph.io/t/a-scalar-type-was-returned-but-graphql-was-expecting-an-object/10908)

-	[There is no implementation for `for each` like statements](https://discuss.dgraph.io/t/foreach-func-in-dql-loops-in-bulk-upsert/5533/9). We need to do that programmatically as was done in the `getImplementedInterfaceIds` function from the [`app/internal/implementation_populator.go`](./app/internal/implementation_populator.go) file.

-	By default, dgraph generates a lot of boilerplate. Each entity has its own mutations/queries. We can disable that in the newest [version](https://dgraph.io/docs/master/graphql/schema/generate) which was not tested during this POC.

-	Dgraph supports RDF and GraphQL schemas but using only the RDF schema result in such error:

	```bash
	"Not resolving queryInterface. There's no GraphQL schema in Dgraph.  Use the /admin API to add a GraphQL schema"
	```

	It means that we need to always create the graphql schema if we want to use GraphQL API.

## Extras
-	[Compression](https://dgraph.io/docs/graphql/api/requests/#compression) is out of the box. Maybe we can use it at the beginning for cache sync.

-	Exclusive features like ACLs, binary backups, encryption at rest, and more: https://dgraph.io/docs/enterprise-features/

## Dgraph Lambda Research

Lambda is just a layer on top of custom resolvers. It has the same limitations as custom resolvers. The only advantage of it is that there is an already available lambda server that allows writing resolvers in JavaScript. A Go server could also be created for the same purpose.

## Needs investigation

-	Can facet help with a query for Implementations that fulfill specific requirements?

-	How Dgraph maps the RDF entity to GraphQL types? Can we reuse that logic in our custom resolvers?

## Summary

The Dgraph solution gives out of the box the GraphQL API and speeds up the development as we can reuse already available GraphQL schema. The Dgraph documentation is quite good but only describes the basic queries/mutations. More sophisticated queries/mutations are not documented and quite often they are just not supported. Dgraph is a good solution for the project which do not need a lot of customization with own business logic.
