# Creating Terraform manifests

In this document you will learn how to use your Terraform modules with Voltron Project.

# Setup

## Starting cluster

During Manifest development you should disable database populator so your manifests are not overwritten.
For already running dev cluster:

```shell
ENABLE_POPULATOR=false make dev-cluster-update
```

To start new cluster without populator:

```shell
ENABLE_POPULATOR=false make dev-cluster
```

## Configuring Minio

In Voltron dev-cluster there is a Minio object store available and can be used for storing modules.
Runner internally is using [go-getter](https://github.com/hashicorp/go-getter) so different sources are supported.

To use Minio, to upload modules, enable port forward:

```shell
kubectl -n argo port-forward svc/argo-minio --address 0.0.0.0 9000:9000
```

Next, if missing, download Minio client from https://min.io/download#/linux  and configure the access:

```shell
SECRETKEY=$(kubectl  -n argo get secret argo-minio -o jsonpath='{.data.secretkey}' | base64 --decode)
ACCESSKEY=$(kubectl  -n argo get secret argo-minio -o jsonpath='{.data.accesskey}' | base64 --decode)

mc alias set minio http://localhost:9000 ${ACCESSKEY} ${SECRETKEY}
```

Verify that you can access Minio:

```shell
mc list minio
```

There should be bucket called `terraform`

## Uploading modules

In `testdata` directory there is a Terraform module to configure CloudSQL Postgresql instance.
Let's upload it:

```shell
mc cp testdata/main.tf minio/terraform/cloudsql/main.tf
```

# Preparing Voltron manifests

To use the module in Voltron you need to prepare Implementation manifest. In Voltron there is
already available Interface(cap.interface.database.postgresql.install) for Postgresql installation.
There are no policies implemented yet, so we can not use it. For now, you will create a new InterfaceGroup, Interface,
and Implementation.

To create CloudSQL instance you need to be authorized. For now, you will use OAuth2 access token. To get it run:

```shell
gcloud auth print-access-token
```

Copy output of this command and replace string <ACCESS_TOKEN> in `manifests/implementation/terraform/gcp/cloudsql/postgresql/install.yaml` file.
Remember that this token is valid only for an hour, so you would need to replace it later.
In the same file replace:

* <PROJECT_NAME> with your projetct name
* <ACCESSKEY> with access key for Minio
* <SECRETKEY> with secret key for Minio

`manifests/implementation/terraform/gcp/cloudsql/postgresql/install.yaml` is a main file where you can configure your Terraform module. You can:

- Set environment variable for terraform binary. This is how we pass ACCESS_TOKEN and GOOGLE_PROJECT here.
- Pass module variables
- Set path to your module.

Now just copy content of `manifests` directory to `och-content` in main directory. Populate new manifests using populator and create an action for path `cap.interface.terraform.database.postgresql.install`. When action is rendered, run it.

You can create an action using for example the GraphQL API:

```graphql
mutation createAction(
        in: {
            name: "postgresql-install",
            actionRef: {
                path: "cap.interface.terraform.database.postgresql.install"
                revision: "0.1.0"
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
```

After around 10 minutes new CloudSQL instance should be running.
