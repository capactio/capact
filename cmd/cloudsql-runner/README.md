# CloudSQL runner

CloudSQL runner is a runner, which creates and manages CloudSQL instances and databases on Google Cloud Platform

## Supported features:

- creating new CloudSQL database instances

## How to build

```bash
# build docker image
make build-app-image-cloudsql-runner

# build only binary
go build -o bin/cloudsql-runner cmd/cloudsql-runner/main.go
```

## How to use

1. Get your GCP credentials file using `gcloud` CLI. This should create the credentials file in `$HOME/.config/gcloud/application_default_credentials.json`. Verify the file has a `.project_id` key, if not, add it manually:
```bash
$ gcloud auth application-default login
$ cat /home/damian/.config/gcloud/application_default_credentials.json
{
  [...]
  "type": "authorized_user",
  "project_id": "<your-gcp-project-id>" # add this, if not present
}
```

2. Create the runner input file:
```bash
cat <<EOF > cloudsql-args.yaml
context:
  name: "cloudsql-example"
  dryRun: false
  timeout: "10m"
  platform: {}
args:
  command: "install"
  generateName: true
  instance:
    tier: "db-g1-small"
    databaseVersion: "POSTGRES_11"
    region: "us-central"
    settings:
      tier: "db-g1-small"
      ipConfiguration:
        authorizedNetworks:
          - name: internet
            value: "0.0.0.0/0"
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

3. Set the following env vars with the runner input file paths:
```bash
export RUNNER_GCP_SERVICE_ACCOUNT_FILEPATH=$HOME/.config/gcloud/application_default_credentials.json
export RUNNER_INPUT_PATH=cloudsql-args.yaml
```

4. Run the runner locally
```bash
go run cmd/cloudsql-runner/main.go
cat helm-release
cat additional
```

## Hacking

Main source code is in:
- `cmd/cloudsql-runner/` - binary main
- `pkg/runner/cloudsql` - CloudSQL runner code
