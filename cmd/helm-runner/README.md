# Helm runner

Helm runner is a runner, which creates and manages Helm releases

## Supported features:

- creating new Helm releases

## How to build

```bash
# build docker image
make build-app-image-helm-runner

# build only binary
go build -o bin/helm-runner cmd/helm-runner/main.go
```

## How to use

1. Create the runner input file:
```bash
cat <<EOF > helm-args.yaml
context:
  name: "helm-example"
  dryRun: false
  timeout: "10m"
  platform:
    namespace: "default"
args:
  command: "install"
  generateName: true
  chart:
    name: "postgresql"
    repo: "https://charts.bitnami.com/bitnami"
  values:
    fullnameOverride: postgresql-server
    postgresqlDatabase: postgres
    postgresqlPassword: s3cr3t
  output:
    directory: "."
    helmRelease:
      fileName: "helm-release"
    additional:
      fileName: "additional"
      value: |-
        host: "{{ template "postgresql.fullname" . }}"
        port: "{{ template "postgresql.port" . }}"
        defaultDBName: "{{ template "postgresql.database" . }}"
        superuser:
          username: "{{ template "postgresql.username" . }}"
          password: "{{ template "postgresql.password" . }}"
EOF
```

2. Set the following env vars with the runner input file paths:
```bash
export RUNNER_INPUT_PATH=helm-args.yaml
```

3. Run the runner locally
```bash
go run cmd/helm-runner/main.go
cat helm-release
cat additional
```

## Hacking

Main source code is in:
- `cmd/helm-runner/` - binary main
- `pkg/runner/helm/` - CloudSQL runner code
