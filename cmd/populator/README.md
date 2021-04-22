# Populator

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Development](#development)

## Overview

The populator is a command-line tool, which helps to populate various Capact content.

## Prerequisites

- [Go](https://golang.org)

## Usage

To run the tool use the following command:
```bash
go run cmd/populator/main.go
```

Check below documents for details how to use the tool: 
* [populator_register-ocf-manifests.md](./docs/populator_register-ocf-manifests.md)	- Populates locally available manifests into Neo4j database.
* [populator_register-capact-installation.md](./docs/populator_register-capact-installation.md)	- Produces and uploads TypeInstances which describe Capact installation.

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
