# ocftool

## Overview

ocftool is a command-line tool, which helps working with OCF manifests. For now it supports validating OCF manifests against manifest JSON schemas.

## Prerequisites

- Go compiler 1.14+

## Usage

To build the ocftool binary use make:
```bash
make build-tool-ocftool        
```

The built binaries for different operating systems are stored in `./bin` directory.

See [ocftool.md](./docs/ocftool.md) for details how to use the tool.
