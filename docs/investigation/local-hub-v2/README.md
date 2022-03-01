# Local Hub

The following directory contains different investigations of using different backend for Local Hub implementation.
We're looking for a good replacement to Neo4j, to resolve problems with high resource consumption and licensing.
Ideally, we want to implement Local Hub in Go.

There are the following proof of concept projects:
- [PostgreSQL](./postresql/README.md) - The goal of this investigation is to find an efficient way to implement Local Hub backed with PostgreSQL.
- [Dgraph](./dgraph/README.md) - The goal of this investigation is to check whether Dgraph can be used as a Local Hub replacement.
- [Generate TypeScript gRPC client](./ts-grpc/README.md) - The goal of this investigation is to find which tools we should use to generate the gRPC client for delegated storage backend.
