# Dgraph Go Lambda Server

It is a Go implementation of Dgraph [lambda server](https://dgraph.io/docs/master/graphql/lambda/server/) with a custom logic for demo purposes.

This project was generated using [dgraph-lambda-go](https://github.com/Schartey/dgraph-lambda-go).

## Prerequisites

- [Go](https://golang.org)
- [Docker](https://docs.docker.com/get-docker/)

## Usage

To run the lambda server use the following command:
```bash
go run main.go
```

To start the whole Dgraph ecosystem, run:
```bash
docker compose up --build
```
