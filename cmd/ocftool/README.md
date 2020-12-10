# ocftool

## Overview

ocftool is a command-line tool, which helps working with OCF manifests. For now it supports validating OCF manifests against manifest JSON schemas

## Prerequisites

- Go compiler 1.14+
- Manifest JSON schemas

## Usage

To build the tool you just use the Go compiler:
```bash
go build cmd/ocftool/main.go
```

Use the help included in the `ocftool` to view available commands:
```bash
ocftool --help
```

### Manifest validation

To validate OCF manifest you can use the `ocftool validate command`

```bash
# validate a OCF manifest file `my-created-implementation.yml`
ocftool validate my-created-implementation.yaml
# validate all yaml's in och-content directory
ocftool validate ./och-content/**/*.yaml
```
