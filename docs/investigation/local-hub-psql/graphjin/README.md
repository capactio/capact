# GraphJin simple example

This is a part of the [Local Hub with PostgreSQL](../README.md) investigation. Read the linked document to learn more.

## Run

```bash
docker-compose run api db create
docker-compose run api db setup
docker-compose up
```

## Execute queries and mutations

Navigate to [http://localhost:8080/](http://localhost:8080/) and run queries and mutations:

> **NOTE**: Copy and paste queries and mutations one by one. I'm not sure why, but if you paste the full snippet below, then executing a given named query/mutation fails with `OpQuery: invalid query` error. I didn't investigate it further as there were no point to do so.

### Queries

```graphql
query tis {
    typeinstances {
        id
        type_ref {
            id
            path
            revision
        }
    }
}

query TypeRefs {
    type_references {
        id
        path
        revision
    }
}

mutation createTITypeRef {
    type_references(insert: $typeRef) {
        path
        revision
    }
}

mutation createTI{
    typeinstances(insert: $ti) {
        id
        type_ref {
            id
            path
            revision
        }
    }
}

mutation createTINested {
    typeinstances(insert: $tiNested) {
        id
        type_ref {
            id
            path
            revision
        }
    }
}
```

### Variables

> **NOTE:** For `tiNested` variables: I needed to guess how to do the nested insert as documentation lacks such guidance. Apparently, you need to provide database name as the property and GraphJin will set proper references. This is counter-intuitive and also against GraphQL contract.

```json
{
  "typeRef": {
    "id": "5bd9f385-a009-4f4e-bae3-681e5ef75c0b",
    "path": "path",
    "revision": "0.1.0"
  },
  "ti": {
    "id": "d993499f-9c95-4749-8629-7fe2818f8eed",
    "type_ref": "5bd9f385-a009-4f4e-bae3-681e5ef75c0b"
  },
  "tiNested": {
    "id": "6d092784-6eaf-4630-ae5c-dadbf1c2a562",
    "type_references": {
      "id": "ee6e7907-413d-4774-b197-1fdea4c090fa",
      "path": "path2",
      "revision": "0.2.0"
    }
  }
}
```
