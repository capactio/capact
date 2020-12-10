# Helm runner

## Overview

Helm runner is a runner [Voltron runner](../../docs/runner.md), which creates and manages Helm releases

## Prerequisites

- Running Kubernetes cluster
- Go compiler 1.14+
- kubectl

## Usage

Normally the runner is started by Voltron Engine, but you can run the runner locally without the Engine.

### Run Helm runner locally

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

2. Set the following env var with the runner input file paths:
```bash
export RUNNER_INPUT_PATH=helm-args.yaml
```

3. Run the runner locally
```bash
go run cmd/helm-runner/main.go
```

4. Verify the results:
```bash
cat helm-release
cat additional
kubectl get pods
```
