# populator voltron-install-type-instances

Produces and uploads TypeInstances which describe Voltron installation.

## Usage

```shell
populator voltron-install-type-instances
```

## Configuration

You can set the following environment variables to configure:

| Name                    | Required | Default                                           | Description                                                                                            |
|-------------------------|----------|---------------------------------------------------|--------------------------------------------------------------------------------------------------------|
| LOCAL_OCH_ENDPOINT      | no       | `https://voltron-och-local.voltron.local/graphql` | Defines local OCH Endpoint.                                                                            |
| VOLTRON_RELEASE_NAME    | no       | `voltron`                                         | Defines Voltron Helm release name.                                                                     |
| HELM_REPOSITORY_PATH    | no       | `voltron`                                         | Defines Helm chart repository URL where the Voltron charts are located.                                |
| HELM_RELEASES_NS_LOOKUP | yes      | -                                                 | Defines Kubernetes Namespaces in which Voltron components were deployed. It is a comma separated list. |
| LOGGER_DEV_MODE         | no       | `false`                                           | Enable development mode logging.                                                                       |
