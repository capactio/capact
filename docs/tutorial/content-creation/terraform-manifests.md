# Creating Terraform manifests

This document describes how to prepare content which uses Terraform Runner.

## Prerequisites

- [MinIO client](https://min.io/download)
- [Voltron development cluster](../../development.md#development-cluster)
    
    > **NOTE:** Use `ENABLE_POPULATOR=false` environmental variable, as you will manually upload your OCF manifests into OCH.

## MinIO access configuration

One of Voltron components is [Minio](https://min.io), which is an object store. It can be used for storing modules.
Terraform Runner internally uses [go-getter](https://github.com/hashicorp/go-getter) so different sources are supported.

To use Minio to upload modules, enable port forward:

```shell
kubectl -n argo port-forward svc/argo-minio --address 0.0.0.0 9000:9000
```

Using MinIO client, configure the access:

```shell
SECRETKEY=$(kubectl  -n argo get secret argo-minio -o jsonpath='{.data.secretkey}' | base64 --decode)
ACCESSKEY=$(kubectl  -n argo get secret argo-minio -o jsonpath='{.data.accesskey}' | base64 --decode)

mc alias set minio http://localhost:9000 ${ACCESSKEY} ${SECRETKEY}
```

Verify that you can access MinIO:

```shell
mc ls minio
```

On the list, you should see the `terraform` bucket, which is created by default.

## Uploading Terraform modules

In the `och-content/implementation/gcp/cloudsql/postgresql/install-0.2.0-module` directory there is a Terraform module to configure CloudSQL Postgresql instance.

1. Create tar directory first:

    ```shell
    cd och-content/implementation/gcp/cloudsql/postgresql/install-0.2.0-module && tar -zcvf /tmp/cloudsql.tgz . && cd -
    ```

1. Upload it to MinIO:

    ```shell
    mc cp /tmp/cloudsql.tgz minio/terraform/cloudsql/cloudsql.tgz
    ```

1. As the `terraform` bucket has `download` policy set by default, you can access all files with unauthenticated HTTP calls.
As you port-forwarded in-cluster MinIO installation, you can check that by using `wget`. Run:

    ```shell
    wget http://localhost:9000/terraform/cloudsql/cloudsql.tgz
    ````

## Preparing Voltron manifests

To use the module, you need to prepare Voltron manifests - InterfaceGroup, Interface, Implementation and Types.

In this example, we have them all already defined for PostgreSQL installation. To create your own manifests, you can base on them:
- [InterfaceGroup](../../och-content/interface/database/postgresql.yaml)
- [Interface](../../och-content/interface/database/postgresql/install.yaml)
- [Implementation](../../och-content/implementation/terraform/gcp/cloudsql/postgresql/install.yaml). The manifest uses Terraform Runner.
  
  Instead of using GCS as module source, you can use internal MinIO URL, such as `http://argo-minio.argo:9000/terraform/cloudsql/cloudsql.tgz`.

- [Input Type](../../och-content/type/database/postgresql/install-input.yaml)
- [Output Type](../../och-content/type/database/postgresql/config.yaml)

## Populating content

To read more how to populate content, see the [Populate the manifests into OCH](./README.md#populate-the-manifests-into-och) section in `README.md` document.

## Running Action

If the MinIO is populated with Terraform content and all manifests are ready, trigger the Jira installation, which will use CloudSQL provisioned with Terraform Runner.

To read how to do it, see the [Install Jira with an external CloudSQL database](../jira-installation/README.md#install-jira-with-an-external-cloudsql-database) section in Jira installation tutorial.
To make sure the Terraform-based Implementation is selected, you may use additional, Attribute-based `implementationConstraint` in Cluster Policy:

```yaml
   # (...)
   rules:
     cap.interface.database.postgresql.install:
       oneOf:
         - implementationConstraints:
             attributes:
               - path: "cap.attribute.cloud.provider.gcp"
               - path: "cap.attribute.infra.iac.terraform" # Add this line
             requires:
               - path: "cap.type.gcp.auth.service-account"
           injectTypeInstances:
             # (...)
```
