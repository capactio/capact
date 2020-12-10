# CloudSQL runner

## Overview

CloudSQL runner is a runner, which creates and manages CloudSQL instances and databases on Google Cloud Platform

## Prerequisites

- GCP project with CloudSQL access
- Go compiler 1.14+

## Usage

### Run locally

You can run the CloudSQL runner locally without Voltron Engine:

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
    defaultDBName: postgres
    rootPassword: s3cr3t
    settings:
      tier: "db-g1-small"
      ipConfiguration:
        authorizedNetworks:
          - name: internet
            value: "0.0.0.0/0"
  output:
    directory: "."
    cloudSQLInstance:
      fileName: "cloudSQLInstance
    additional:
      fileName: "additional"
      value:
        host: "{{ (index .DBInstance.IpAddresses 0).IpAddress  }}"
        port: "{{ .Port }}"
        defaultDBName: "{{ .DefaultDBName }}"
        superuser:
          username: "{{ .Username }}"
          password: "{{ .Password }}"
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
cat cloudSQLInstance
cat additional
```
