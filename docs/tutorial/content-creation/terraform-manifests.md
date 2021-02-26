# Creating Terraform manifests

This document describes how to use Terraform content to run PostgreSQL installation.
Based on this guide you may prepare your own manifests and utilize your own Terraform content.

## Setup

### Starting cluster

Run development cluster:

```bash
make dev-cluster
```

### Configuring Minio

In Voltron dev-cluster there is a Minio object store available and can be used for storing modules.
Terraform Runner internally uses [go-getter](https://github.com/hashicorp/go-getter) so different sources are supported.

To use Minio to upload modules, enable port forward:

```shell
kubectl -n argo port-forward svc/argo-minio --address 0.0.0.0 9000:9000
```

Next, if missing, download Minio client from [https://min.io](https://min.io/download)  and configure the access:

```shell
SECRETKEY=$(kubectl  -n argo get secret argo-minio -o jsonpath='{.data.secretkey}' | base64 --decode)
ACCESSKEY=$(kubectl  -n argo get secret argo-minio -o jsonpath='{.data.accesskey}' | base64 --decode)

mc alias set minio http://localhost:9000 ${ACCESSKEY} ${SECRETKEY}
```

Verify that you can access Minio:

```shell
mc ls minio
```

On the list, you should see the `terraform` bucket, which is created by default.

### Uploading modules

In `testdata` directory there is a Terraform module to configure CloudSQL Postgresql instance.
Let's create tar directory first:

```shell
cd ./assets/source && tar -zcvf ../cloudsql.tgz . && cd -
```

And upload it to Minio:

```shell
mc cp ./assets/cloudsql.tgz minio/terraform/cloudsql/cloudsql.tgz
```

As the `terraform` bucket has `download` policy set by default, you can access all files with unauthenticated HTTP calls.
As you port-forwarded in-cluster Minio installation, you can check that by using `wget`. Run:

```shell
wget http://localhost:9000/terraform/cloudsql/cloudsql.tgz
````

## Preparing Voltron manifests

To use the module, you need to prepare Voltron manifests - InterfaceGroup, Interface, Implementation and Types.

In this example, we have them all already defined for PostgreSQL installation. To create your own manifests, you can base on them:
- [InterfaceGroup](../../och-content/interface/database/postgresql.yaml)
- [Interface](../../och-content/interface/database/postgresql/install.yaml)
- [Implementation](../../och-content/implementation/terraform/gcp/cloudsql/postgresql/install.yaml), which uses Terraform Runner.\
  
   Instead of using GCS as module source, you can use internal Minio URL, such as "http://argo-minio.argo:9000/terraform/cloudsql/cloudsql.tgz".

- [Input Type](../../och-content/type/database/postgresql/install-input.yaml)
- [Output Type](../../och-content/type/database/postgresql/config.yaml)

## Installing PostgreSQL

If the Minio is populated with Terraform content and all manifests are ready, trigger the PostgreSQL installation.

### Creating TypeInstance

1. Get JSON file with ServiceAccount according to the instruction [here](https://github.com/Project-Voltron/go-voltron/tree/master/docs/tutorial/jira-installation#install-jira-with-managed-cloud-sql). 
   
1. Convert JSON to JS (remove quotes around object keys) and paste it into `in.value` for `createTypeInstance` input:

    ```bash
    cat sa.json | sed -E 's/(^ *)"([^"]*)":/\1\2:/'
    ```

1. Navigate to [https://gateway.voltron.local](https://gateway.voltron.local). Copy the JS object with ServiceAccount to input of the mutation and create GCP SA TypeInstance:
    
    ```graphql
    mutation CreateGCPSATypeInstance {
      createTypeInstance(
        in: {
          typeRef: { path: "cap.type.gcp.auth.service-account", revision: "0.1.0" }
          value: {type:"service_account",project_id:"projectvoltron",private_key_id:"...", ... <-- replace this }
          attributes: [
            { path: "cap.attribute.cloud.provider.gcp", revision: "0.1.0" }
          ]
        }
      ) {
        metadata {
          id
        }
        spec {
          typeRef {
            path
            revision
          }
        }
      }
    }
    ```
   
1. Note the TypeInstance ID. You will need that for Policy configuration.

### Configure Policy

Configure the Policy with the following command:

```bash
kubectl edit configmap -n voltron-system voltron-engine-cluster-policy
```

Copy and paste the following content.

- Replace `{gcp-sa-uuid}` with the actual TypeInstance ID with GCP Service Account from one of previous steps.

```yaml
data:
  cluster-policy.yaml: |
    apiVersion: 0.1.0

    rules:
      cap.interface.database.postgresql.install:
        oneOf:
          - implementationConstraints:
              requires:
                - path: "cap.type.gcp.auth.service-account"
                  revision: "0.1.0"
              attributes:
                - path: "cap.attribute.cloud.provider.gcp"
                  # any revision
                - path: "cap.attribute.infra.iac.terraform"
                  # any revision
            injectTypeInstances:
              - id: {gcp-sa-uuid}
                typeRef:
                  path: "cap.type.gcp.auth.service-account"
                  revision: "0.1.0"
      cap.*:
        oneOf:
          - implementationConstraints:
              requires:
                - path: "cap.core.type.platform.kubernetes"
                  # any revision
          - implementationConstraints: {}
```

### Create Action
Now you can create an Action using the GraphQL API on [https://gateway.voltron.local](https://gateway.voltron.local):

To provision standalone CloudSQL run:

```graphql
mutation CreateCloudSQLAction {
    createAction(
        in: {
            name: "action-install",
            actionRef: {
                path: "cap.interface.database.postgresql.install",
                revision: "0.1.0",
            },
            dryRun: false,
            advancedRendering: false,
            input: {
                parameters: "{\r\n  \"superuser\": {\r\n    \"username\": \"postgres\",\r\n    \"password\": \"s3cr3t\"\r\n  },\r\n  \"defaultDBName\": \"postgres\"\r\n}"
            }
        }
    ) {
        name
        input {
            parameters
        }
    }
}
```

If you want to run Jira installation with CloudSQL, then use:

```graphql
mutation CreateJiraAction {
    createAction(
        in: {
            name: "action-install",
            actionRef: {
                path: "cap.interface.database.postgresql.install",
                revision: "0.1.0",
            },
            dryRun: false,
            advancedRendering: false,
            input: {
                parameters: "{\r\n  \"superuser\": {\r\n    \"username\": \"postgres\",\r\n    \"password\": \"s3cr3t\"\r\n  },\r\n  \"defaultDBName\": \"postgres\"\r\n}"
            }
        }
    ) {
        name
        input {
            parameters
        }
    }
}
```

### Run Action

Run Action:

```graphql
mutation Run {
    runAction(name: "action-install") {
        name
    }
}
```

### Observe Running Action:

Observe the running workflow with Argo CLI:

```bash
argo watch action-install
```

> **NOTE**: CloudSQL instance provisioning can take around  10 minutes.

### Cleanup

1. Navigate to [https://console.cloud.google.com](https://console.cloud.google.com) and delete the CloudSQL instance.
1. To delete Jira and other Helm charts in the `default` namespace, run:

   ```bash
   helm delete $(helm list -q)
   ```
