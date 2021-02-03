# OCH Content Lifecycle

Created on 2021-01-14 by Paweł Kosiec ([@pkosiec](https://github.com/pkosiec)).

## Overview

This document describes way, how to achieve zero-downtime content synchronization between Git repository, where the OCF
manifests are stored, and the OCH graph database.

<!-- toc -->

  * [Motivation](#motivation)
  * [Goal](#goal)
  * [Assumptions](#assumptions)
  * [Proposal](#proposal)
    + [Suggested solution: Node Labels](#suggested-solution-node-labels)
      - [Algorithm](#algorithm)
      - [Summary](#summary)
    + [Alternative: Entrypoint node](#alternative-entrypoint-node)
      - [Algorithm](#algorithm-1)
      - [Summary](#summary-1)
    + [Other alternatives](#other-alternatives)
- [Consequences](#consequences)

<!-- tocstop -->

## Motivation

The OCH component stores OCF manifests in graph database. To populate the database, we introduced DB Populator, which
fetches OCF manifests from a given Git repository, and populates the OCH database.

As the content on Git repository changes over time, we need a way to update the content in a way, that OCH doesn't have
any downtime.

## Goal

Prepare zero-downtime OCH content update strategy.

## Assumptions

- During database population by DB Populator, DB Populator generates random ID for every node and puts it in the field
  with `@id` directive in GraphQL schema.

## Proposal

### Suggested solution: Node Labels

Node labels
are [recommended as multi-tenancy workaround for Neo4j 3.x](https://community.neo4j.com/t/proper-way-to-implement-multi-tenancy-on-neo4j/625/6)
, which didn't support multiple databases. This is similar case to ours, as we want to achieve graph separation inside
the same database.

The `graphql-neo4j-js` has
an [`@additionalLabels` directive](https://grandstack.io/docs/neo4j-graphql-js-middleware-authorization/#additionallabels)
, which allows us to provide an additional label for Cypher query based on context.

In a result, all "entrypoint" GraphQL Types, that are used for GraphQL queries, would be defined as:

```graphql
type InterfaceGroup @additionalLabels(labels: ["published"]) { # new directive
    id: ID!
    # (...) other fields
}

type Interface @additionalLabels(labels: ["published"]) { # new directive
    id: ID!
    # (...) other fields
}

type Implementation @additionalLabels(labels: ["published"]) { # new directive
    id: ID!
    # (...) other fields
}
```

Note if a `@cypher` directive is used for a given type, the label has to be put manually. For example:

```graphql
type Query {
    # add "published" label inside custom Cypher query - without it it won't filter nodes properly
    interfaceGroupCustom(path: NodePath!): InterfaceGroup @cypher(statement: "MATCH (g:InterfaceGroup:published)-[:DESCRIBED_BY]->(m:GenericMetadata {path: $path}) return g")
}
```

#### Algorithm

The following section describes the DB Populator algorithm for populating the OCH content.

1. Populate initial content into database
    - Generate random Node IDs
    - Set `published` label for every node, e.g.:

      ```
      MATCH (n:InterfaceGroup)
      CALL apoc.create.addLabels( n, [ "published" ] )
      YIELD node
      Return count(n)
      ```

      ```
      MATCH (n:Interface)
      CALL apoc.create.addLabels( n, [ "published" ] )
      YIELD node
      Return count(n)
      ```

      ```
      MATCH (n:Implementation)
      CALL apoc.create.addLabels( n, [ "published" ] )
      YIELD node
      Return count(n)
      ```

    - Create node with populated content details:

       ```
       CREATE (n:ContentMetadata:published { repository: 'git@github.com:Project-Voltron/go-voltron.git', commit: 'f2cd6a3' })
       ```

1. Every some period of time, check if OCF manifest changed in Git repository.

    - Use `commit` from `ContentMetadata:published` node to check the diff between new commit and commit from which the
      content has been populated.
    - If there is any change in `och-content` directory, continue.

1. Populate full new content into database.

    - Generate random Node IDs
    - Set `unpublished` label for every node
    - Create node with populated content details:

       ```
       CREATE (n:ContentMetadata:unpublished { repository: 'git@github.com:Project-Voltron/go-voltron.git', commit: 'b259e25' })
       ```

1. In one transaction:

    - Get nodes with "published" label, add them label "to_remove" and remove label "published"

   > **NOTE:** On production we can use `apoc.cypher.runMany` to run many APOC statements in one block.

    ```
    MATCH (n:published)
    CALL apoc.create.addLabels( n, [ "to_remove" ] ) YIELD node

    Return count(node)
    ```

    ```
    MATCH (n:published)
    CALL apoc.create.removeLabels( n, [ "published" ] ) YIELD node

    Return count(node)
    ```

    - Get nodes with "unpublished" label, add them label "published" and remove label "unpublished"

1. Remove nodes with label `to_remove` using `apoc.periodic.iterate`

    ```
    call apoc.periodic.iterate("MATCH (n:to_remove) return n", "DETACH DELETE n", {batchSize:1000})
    yield batches, total return batches, total
    ```

#### Summary

- Simple set up in GraphQL schema: Use directive on every GraphQL type, which is exposed as GraphQL query.

  > **NOTE**: Any GraphQL mutation related to these types will also include the label. This shouldn't be a problem as the generated GraphQL mutations will be disabled.

- If there are any custom Cypher query for GraphQL queries, it needs to be adjusted as well.
- It is not optimal solution regarding performance, as we need to update all nodes in one transaction (add and delete
  labels).
- All `neo4j-graphql-js` features are still supported after these adjustments.
- This solution is also applicable for synchronizing content of OCH once we implement federation support. The only change is that we will replace OCH vendor subgraph, instead of whole OCH graph.  
- In the future, we may expose `ContentMetadata` node details as a part of `repoMetadata` GraphQL query.

### Alternative: Entrypoint node

Assume that there is a "Pointer" node, which points to every node, which has dedicated GraphQL query (Type, Attribute,
InterfaceGroup, Interface, Implementation, RepoMetadata). Using custom `@cypher` directives in GraphQL schema, we can
filter out all nodes, which doesn't have relation with Pointer:

```graphql
type Query {
    repoMetadata: RepoMetadata @cypher(statement: "MATCH (r:RepoMetadata)<-[:POINTS_TO]-(p:Pointer:published) RETURN r")

    interfaceGroups: [InterfaceGroup!]! @cypher(statement: "MATCH (i:InterfaceGroup)<-[:POINTS_TO]-(p:Pointer:published) RETURN i")
    interfaceGroup(path: NodePath!): InterfaceGroup @cypher(statement: "MATCH (p:Pointer:published)-[:POINTS_TO]->(g:InterfaceGroup)-[:DESCRIBED_BY]->(m:GenericMetadata {path: $path}) return g")

    # (...)
}
```

#### Algorithm

The following section describes the DB Populator algorithm for populating the OCH content.

1. Populate initial content into database
    - Generate random Node IDs
    - Create Pointer Node with "published" label.
      ```
      CREATE (n:Pointer:published { repository: 'git@github.com:Project-Voltron/go-voltron.git', commit: 'f2cd6a3' })
      ```
    - Point the pointer for every node, which has dedicated GraphQL query (Type, Attribute, InterfaceGroup, Interface,
      Implementation, RepoMetadata)

      For example:

      ```
      MATCH (p:Pointer:published), (i: Implementation)
      MERGE (p)-[r:POINTS_TO]->(i)
      RETURN r
      ``` 

      ```
      MATCH (p:Pointer:published), (i: InterfaceGroup)
      MERGE (p)-[r:POINTS_TO]->(i)
      RETURN r
      ``` 
1. Every some period of time, check if OCF manifest changed in Git repository.
    - Use `commit` from Pointer to check the diff between new commit and commit from which the content has been
      populated.
    - If there is any change in `och-content` directory, continue.

1. Populate full new content into database.

    - Generate random Node IDs
    - Create "pointer" Node with "unpublished" label

      ```
      CREATE (n:Pointer:unpublished { repository: 'git@github.com:Project-Voltron/go-voltron.git', commit: 'b259e25' })
      ```

    - Point the pointer for every new node, which has dedicated GraphQL query (Type, Attribute, InterfaceGroup,
      Interface, Implementation, RepoMetadata)

1. In one transaction:

    - Set Pointer node with "published" label to "to_remove"

      ```
      MATCH (n:Pointer:published)
      CALL apoc.create.setLabels( n, [ "Pointer", "to_remove" ] ) YIELD node

      Return count(node)
      ```

    - Set Pointer node with "unpublished" label to "published"

      ```
      MATCH (n:Pointer:unpublished)
      CALL apoc.create.setLabels( n, [ "Pointer", "published" ] ) YIELD node

      Return count(node)
      ```

1. Remove Pointer node with label `to_remove` along with all related nodes recursively

    ```
    MATCH (n:Pointer:to_remove)-[r*]-(e)
    FOREACH (rel IN r| DELETE rel)
    DELETE e
    ```

#### Summary

- Easy setup: Custom `@cypher` directives on all queries.
- Limitation: Losing ability to use [built-in filtering capabilities](https://grandstack.io/docs/graphql-filtering) for
  generated queries.
- Most performant solution: publishing content is a matter of unlabeling one node and labeling another.
- This solution is also applicable for synchronizing content of OCH once we implement federation support. The only change is that we will replace OCH vendor subgraph, instead of whole OCH graph.  
- In the future, we may expose `Pointer` node details as a part of `repoMetadata` GraphQL query.

Because of serious limitation of this solution, it is not suggested.

### Other alternatives

- Detect difference between Git repository and Neo4j database and prepare MERGE Cypher queries to update data.

  While it may be considered as optimal solution, it is too complex to implement it for Beta.

- Swapping Neo4j deployment.

  Two running Neo4j at the same time would consume too much resources. Also, the solution would depend on Kubernetes, which would be an issue at some point, when we will go platform-agnostic.

- Creating new database, populating it and then switching connection for OCH.

  While using multiple databases is
  a [recommended solution for Neo4j 4.x multi-tenancy implementation]([https://neo4j.com/developer/multi-tenancy-worked-example](https://neo4j.com/developer/multi-tenancy-worked-example/)
  , it is only supported in Enterprise version. Also, we would need to update OCH component configuration to point to
  different database in runtime, which doesn't seem as efficient solution.

# Consequences

- Generate random IDs for every node in the graph as `_nodeID` field. The field, will be visible in GraphQL schema.

  The ID field for every node in graph will change periodically. However, the field should be well documented, and then
  API consumers will know that they shouldn’t depend on it.

- Implement OCH content update algorithm according to the document.
