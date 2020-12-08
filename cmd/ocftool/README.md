# ocftool

ocftool is a command-line tool, which helps working with OCF manifests

## Supported features:

- validating OCF manifests agains JSON schemas

## How to build

```bash
make build-tool-ocftool
# or
go build -o bin/ocftool cmd/ocftool/main.go
```

## How to use

Use the help included in the `ocftool`:
```bash
ocftool --help
ocftool validate --help
```

### Manifest validation

```bash
# validate a OCF manifest file `my-created-implementation.yml`
ocftool validate my-created-implementation.yaml
# validate all yaml's in och-content directory
ocftool validate ./och-content/**/*.yaml
```

## Hacking

Main source code is in:
- `cmd/ocftool/` - binary main
- `pkg/sdk/manifests/` - manifest validation SDK
- `internal/ocftool/` - private source code
