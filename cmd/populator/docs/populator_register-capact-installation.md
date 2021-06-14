# populator register capact-installation

Produces and uploads TypeInstances which describe Capact installation.

## Usage

```shell
populator register capact-installation
```

## Configuration

You can set the following environment variables to configure:

| Name                    | Required | Default                                         | Description                                                                                            |
|-------------------------|----------|-------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| LOCAL_HUB_ENDPOINT      | no       | `http://capact-hub-local.capact-system/graphql` | Defines local Hub Endpoint.                                                                            |
| CAPACT_RELEASE_NAME     | no       | `capact`                                        | Defines Capact Helm release name.                                                                     |
| HELM_REPOSITORY_PATH    | no       | `capact`                                        | Defines Helm chart repository URL where the Capact charts are located.                                |
| HELM_RELEASES_NS_LOOKUP | yes      | -                                               | Defines Kubernetes Namespaces in which Capact components were deployed. It is a comma separated list. |
| LOGGER_DEV_MODE         | no       | `false`                                         | Enable development mode logging.                                                                       |
