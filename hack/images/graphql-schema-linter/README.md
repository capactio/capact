# graphql-schema-linter Docker image

## Overview

This folder contains Docker image with a linter for GraphQL files.

The Docker image consists of [`graphql-schema-linter`](https://github.com/cjoudrey/graphql-schema-linter) tool, installed globally and a helper script, `lint-multiple-files.sh`. The script is an entrypoint of the image, and it is used to run linter against multiple separate GraphQL schemas.

## Installation

To build the Docker image, run this command:

```bash
docker build -t graphql-schema-linter .
```

## Configuration

You can configure the linter script passing the following arguments:

| Flag                        | Optional | Description                                                                                     |
| --------------------------- | -------- | ----------------------------------------------------------------------------------------------- |
| `--src "{path-to-schema}`   | No       | Path to GraphQL schema to validate. You can use the flag multiple times to lint multiple files. |
| `--linter-args "{options}"` | No       | Additional arguments for `graphql-schema-linter`.                                              |
