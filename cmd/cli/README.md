# Capact CLI

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Development](#development)

## Overview

Capact CLI is a command-line tool, which manages Capact resources. It supports validating OCF manifests against manifest JSON schemas.

Capact CLI uses GraphQL API for managing various Capact resources, such as Action, TypeInstances or Policy.

## Prerequisites

- [Go](https://golang.org)

## Usage

To run the tool use the following command:
```bash
go run cmd/cli/main.go
```

See [capact.md](./docs/capact.md) for details how to use the tool.

## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
