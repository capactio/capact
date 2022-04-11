# argo-actions

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)

## Overview

argo-actions is intended to be run as an Argo Workflow step which downloads, uploads, updates and deletes Type Instances.

## Prerequisites

- [Go](https://golang.org)
- [Local Capact development cluster](https://capact.io/community/development/development-guide#development-cluster) created with the `USE_TEST_SETUP=true make dev-cluster` command

## Usage

To run it locally you need to enable port forwarding for the Local and Public Hub:
```bash
kubectl -n capact-system port-forward svc/capact-hub-local --address 0.0.0.0 8888:80
kubectl -n capact-system port-forward svc/capact-hub-public --address 0.0.0.0 8890:80
```

### Download

For downloading at least one Type Instance needs to exist. Passing structs using environment variables looks like this: {field1,field2}. For example APP_DOWNLOAD_CONFIG="{ID,path}"

```bash
APP_ACTION=DownloadAction APP_DOWNLOAD_CONFIG="{2282814e-7571-4708-9279-717aea3c6d08,/tmp/action.yaml}" APP_LOCAL_HUB_ENDPOINT=http://localhost:8888/graphql go run cmd/argo-actions/main.go
```

### Upload

To upload example TypeInstances:

1. Install [Helm Storage Backends](https://capact.io/docs/feature/storage-backends/helm) and use them in Policy
2. Install PostgreSQL according to the [Helm Runner installation](../helm-runner/README.md#installation) section.
3. In the [`payload.yaml`](./example-input/upload/payload.yaml) file:
   - replace `{helm release storage backend ID}` with the ID of the installed Helm Release backend 
   - replace `{test storage backend ID}` with the ID of the installed Helm Release backend 
4. Run:

    ```bash
    APP_ACTION=UploadAction \
      APP_UPLOAD_CONFIG_PAYLOAD_FILEPATH=cmd/argo-actions/example-input/upload/payload.yaml \
      APP_UPLOAD_CONFIG_TYPE_INSTANCES_DIR=cmd/argo-actions/example-input/upload/typeinstances \
      APP_LOCAL_HUB_ENDPOINT=http://localhost:8888/graphql \
      APP_PUBLIC_HUB_ENDPOINT=http://localhost:8890/graphql \
      go run cmd/argo-actions/main.go
    ```

## Configuration

The following environment variables can be set:

| Name                     | Required | Default                                         | Description                                            |
|--------------------------|----------|-------------------------------------------------|--------------------------------------------------------|
| APP_ACTION               | yes      |                                                 | Defines action to perform |
| APP_LOCAL_HUB_ENDPOINT   | no       | http://capact-hub-local.capact-system/graphql   | Defines local Hub Endpoint |
| APP_PUBLIC_HUB_ENDPOINT  | no       | http://capact-hub-public.capact-system/graphql  | Defines public Hub Endpoint |
| APP_DOWNLOAD_CONFIG      | no       |                                                 | For download action defines Type Instances to download |
| APP_LOGGER_DEV_MODE      | no       | `false`                                         | Enable additional log messages            |

## Development

To read more about development, see the [Development guide](https://capact.io/community/development/development-guide).
