# CloudSQL runner

## Overview

CloudSQL runner is a runner, which creates and manages CloudSQL instances and databases on Google Cloud Platform

## Prerequisites

- Go compiler 1.14+
- GCP project with CloudSQL access

You need to get a GCP credentials file using `gcloud` CLI, so the runner can connect to your GCP project:
```bash
gcloud auth application-default login
```
This should create the credentials file in `$HOME/.config/gcloud/application_default_credentials.json`. Verify the file has a `.project_id` key. If not, add it manually:
```bash
cat $HOME/.config/gcloud/application_default_credentials.json
{
  [...]
  "type": "authorized_user",
  "project_id": "<your-gcp-project-id>" # add this, if not present
}
```

## Usage

To start the runner type:
```bash
RUNNER_GCP_SERVICE_ACCOUNT_FILEPATH=$HOME/.config/gcloud/application_default_credentials.json \
  RUNNER_INPUT_PATH=cmd/cloudsql-runner/example-input.yml
  go run cmd/cloudsql-runner/main.go
```

## Configuration

The following environment variables can be set:

| Name                                | Default            | Description                           |
|-------------------------------------|--------------------|---------------------------------------|
| RUNNER_INPUT_PATH                   |                    | Path of the runner YAML input file    |
| RUNNER_LOGGER_DEV_MODE              | `false`            | Enable additional log messages        |
| RUNNER_GCP_SERVICE_ACCOUNT_FILEPATH | `/etc/gcp/sa.json` | Path to the GCP JSON credentials file |
