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

### Running

To run it locally you need to enable port forwarding for Local OCH:
```bash
kubectl -n voltron-system port-forward svc/voltron-och-local --address 0.0.0.0 8888:80
```

For downloading at least one Type Instance needs to exist. Passing structs using environment variables looks like this: {field1,field2}. For example APP_DOWNLOAD_CONFIG="{ID,path}"

```bash
APP_ACTION=DownloadAction APP_DOWNLOAD_CONFIG="{2282814e-7571-4708-9279-717aea3c6d08,/tmp/action.yaml}" APP_LOCAL_OCH_ENDPOINT=http://localhost:8888/graphql ./argo-actions
```

## Configuration

The following environment variables can be set:

| Name                     | Required | Default                                         | Description                                            |
|--------------------------|----------|-------------------------------------------------|--------------------------------------------------------|
| APP_ACTION               | yes      |                                                 | Defines action to perform |
| APP_LOCAL_OCH_ENDPOINT   | no       | https://voltron-och-local.voltron.local/graphql | Defines local OCH Endpoint |
| APP_DOWNLOAD_CONFIG      | no       |                                                 | For download action defines Type Instances to download |
| APP_LOGGER_DEV_MODE      | no       | `false`                                         | Enable additional log messages            |

## Development

To read more about development, see the [`development.md`](../../docs/development.md) document.
