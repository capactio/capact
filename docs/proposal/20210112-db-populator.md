# db-populator

Created on 2021-01-12 by Łukasz Oleś (@lukaszo)

## Overview

This document describes way, how to populate OCF manifests from Git repository into Neo4j database.

<!-- toc -->

- [Motivation](#motivation)
  * [Goal](#goal)
  * [Non-goal](#non-goal)
- [Proposal](#proposal)
  + [Load as JSON using CYPHER and APOC](#load-as-json-using-cypher-and-apoc)
  + [Alternatives](#alternatives)
    - [neo4j-admin import](#neo4j-admin-import)
    - [Cypher - LOAD CSV](#cyphera--load-csv)
    - [Cypher - CREATE/MERGE](#cypher---createmerge)
    - [GraphQL Mutations](#graphql-mutations)
- [Consequences](#consequences)

<!-- tocstop -->

## Motivation

OCF manifests are stored in yaml format defined in OCF spec. We need a fast way to populate the manifests into the Neo4j db.

### Goal

- Prepare a strategy to populate data into db.

### Non-goal

- Preparing strategy for life-cycle-management

## Proposal

### Load as JSON using Cypher and APOC

1. Convert manifests to JSON
   
   yaml format:
   ```
   ocfVersion: 0.0.1
   revision: 0.1.0
   kind: InterfaceGroup
   metadata:
     prefix: cap.interface.productivity # Computed during fetching the manifest
     name: jira
     displayName: "Jira"
     description: "The #1 software development tool used by agile teams"
     documentationURL: https://support.atlassian.com/jira-software-cloud/resources/
     supportURL: https://www.atlassian.com/software/jira
     iconURL: https://www.atlassian.com/pl/dam/jcr:e33efd9e-e0b8-4d61-a24d-68a48ef99ed5/Jira%20Software@2x-blue.png
     maintainers:
       - email: team-dev@projectvoltron.dev
         name: Voltron Dev Team
         url: https://projectvoltron.dev
   signature:
     och: eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
   ```

   JSON format:
   ```
   {
     "ocfVersion": "0.0.1",
     "revision": "0.1.0",
     "kind": "InterfaceGroup",
     "metadata": {
       "name": "jira",
       "path": "cap.interface.productivity.jira",
       "prefix": "cap.interface.productivity",
       "displayName": "Jira",
       "description": "Jira Application",
       "maintainers": [
         {
           "name": "Voltron Dev Team",
           "email": "team-dev@projectvoltron.dev",
           "url": "https://projectvoltron.dev"
         }
       ],
       "documentationURL": "https://projectvoltron.dev",
       "supportURL": "https://projectvoltron.dev/contact",
       "iconURL": "https://projectvoltron.dev/favicon.ico"
     },
     "signature": {
       "och": "eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9"
     },
   }
   ```

2. Upload JSON to minio object store
3. Use APOC function and CYPHER to load data
  ```
  cypher
  call apoc.load.json("http://minio.svc.argo.cluster.local:8000/interfaceGroups.json") yield value
  merge (g:GroupInterface)
  with value, g
  unwind value.metadata as m
  merge (gim:GroupInterfaceMetadata {name: m.name, prefix: m.prefix, path: m.path})
  with value, g, gim
  unwind value.signature as sig
  merge (s:Signature{och: sig.och})
  merge (g)-[:DESCRIBED_BY]->(gim)
  merge (g)-[:SIGNED_WITH]->(s);

  ```

#### Summary 

1. Pros
    - Fast approach, should be good for up to 10M nodes.
    - JSON structure is similar to GraphQL output.
2. Cons
    - Additional storage is required to store JSON objects which are lated downloaded by Neo4j.

### Alternatives

#### neo4j-admin import 
Example:
```
  ../bin/neo4j-admin import --database public
       --nodes=GroupInterface=groupInterface.csv
       --nodes=groupInterfaceMetadata.csv
       --nodes=Signature="signature_header.csv,signatures.csv,signatures-2.csv"
       --relationships=DESCRIBED_BY=metadataForGroupInterface.csv
       --relationships=SIGNED_WITH="InterfaceGroupSigned.csv"
```

*groupInterface.csv*
```
groupInterfaceId:ID(GroupInterface)
1
```

*groupInterfaceMetadata.csv*
```
groupInterfaceMetadataId:ID(GroupInterfaceMetadata), name, path, prefix
1,jira,cap.interface.productivity.jira,cap.interface.productivity
```

*metadataForGroupInterface.csv*
```
:START_ID(GroupInterface),:END_ID(GroupInterfaceMetadata)
1,1
```

1. Pros
    - The fastest approach
2. Cons
    - Requires using neo4j admin binary which I guess is written in java. There may be also some licensing issues.
    - Requires converting yaml to CSV.


  
#### Cypher - LOAD CSV
https://neo4j.com/developer/guide-import-csv/#import-load-csv

Groups.csv
```
groupID,GroupMetadataID,SignatureID
1,1,12
```
  
*GroupMetadata.csv*
```
id, name, path, prefix
1,jira,cap.interface.productivity.jira,cap.interface.productivity
```

*Signatures.csv*
```
id, och
12,eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9
```
  
*Cypher*
```
LOAD CSV WITH HEADERS FROM 'file:///signatures.csv' AS row
MERGE (s:Signature {id: row.id, och: row.och})
RETURN count(s);
 
LOAD CSV WITH HEADERS FROM 'file:///GroupMetadata.csv' AS row
MERGE (m:GroupInterfaceMetadata {id: row.Id, name: row.name, path: row.path, prefix: row.prefix}
RETURN count(c);
  
LOAD CSV WITH HEADERS FROM 'file:///groups.csv' AS row
Merge (g:GroupInterface{id: row.id})
MATCH (s:Signature {id: row.SignatureID})
MATCH (m:GroupInterfaceMetadata {id: row.GroupMetadataID})
MERGE (g)-[:SIGNED_WITH]->(s)
MERGE (g)-[:DESCRIBED_BY]->(m)
RETURN *;
```

1. Pros
    - Should be good enough for 10M records.
2. Cons
    - Requires converting yaml to CSV.
  
#### Cypher - CREATE/MERGE
1. Create everything converting manifests to CYPHER queries
    ```
    CREATE p = (GroupInterfaceMetadata {name: "jira", path: "cap.interface.productivity.jira", prefix: "cap.interface.productivity.jira"})-[:DESCRIBED_BY]->(:GroupInterface)<-[:SIGNED_WITH]-(:Signature {och: "eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9"}) Return p
    ```
2. Pros
    - We control every aspect of the process
3. Cons
    - Queries will be generated from Go code. It will be harder to debug and maintain.


#### GraphQL Mutations

1. Pros
    - One GrapQL API for everything
    - One schema

2. Cons
    - Slower than JSON/CSV approach.
    - Requires restricting public OCH mutations.

# Consequences

- Additional integration with Object Store is required.
