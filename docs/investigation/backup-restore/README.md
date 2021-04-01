# Backup and Restore actions for Voltron

This proof of concept shows [Velero](https://velero.io) as a tool used for running backup and restore for workloads deployed by Voltron.

## Prerequisites
- Deployed Voltron dev cluster with disabled populator
- Installed locally [Velero CLI](https://velero.io/docs/v1.5/basic-install/)

## Velero installation

For Object Store we will deploy a new MinIO instance.

```shell
helm repo add minio https://helm.min.io/
helm install --namespace minio --create-namespace --set accessKey=minio,secretKey=miniominio,defaultBucket.enabled=true minio minio/minio 
```

Create `credentials-velero` file with credentials to our MinIO server:

```shell
cat > /tmp/credentials-velero << EOF
[default]
aws_access_key_id=minio 
aws_secret_access_key=miniominio
EOF
```

and install Velero

```shell
velero install --provider aws --bucket=bucket --secret-file /tmp/credentials-velero --backup-location-config region=minio,s3ForcePathStyle="true",s3Url=http://minio.minio.svc:9000 --plugins velero/velero-plugin-for-aws:v1.0.0 --use-volume-snapshots=false --use-restic
```

Velero stores logs in object store. To see logs locally we need access to `minio.minio.svc`. The easiest way
is to add it to `/etc/hosts` and to enable port forwarding:

```shell
kubectl -n argo port-forward service/argo-minio 9000
```

## Installing Persistent Volume provider

For local development we need a Persistent Volume provider which is not using `hostPath`.
We will use [openebs/lvm-localpv](https://github.com/openebs/lvm-localpv/) here. It's still in alpha stage but provides all we need.

1. Create LVM Volume Group inside docker container

   ```shell
   docker exec -it kind-dev-voltron-control-plane bash
   apt update
   apt install lvm2
   modprobe dm-snapshot
   truncate -s 1024G /tmp/disk.img
   disk=`losetup -f /tmp/disk.img --show`
   pvcreate "$disk"
   vgcreate lvmvg "$disk"
   exit
   ```

1. Install `lvm-localpv`

   ```shell
   kubectl apply -f https://raw.githubusercontent.com/openebs/lvm-localpv/master/deploy/lvm-operator.yaml
   ```

1. Configure it to be a default storage:

   ```shell
   cat <<EOF >sc.yaml
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   https://github.com/openebs/lvm-localpv/metadata:
     name: openebs-lvmpv
   parameters:
     storage: "lvm"
     volgroup: "lvmvg"
   provisioner: local.csi.openebs.io
   EOF

   kubectl apply -f sc.yaml

   kubectl patch storageclass standard -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"false"}}}'
   kubectl patch storageclass openebs-lvmpv -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
   ```

1. Recreate Voltron to use new storage.
   To be able to backup and restore Voltron components it needs to use a new persistent storage.

   ```shell
   kubectl delete ns argo,neo4j

   ENABLE_POPULATOR=false make dev-cluster-update
   ```

   Make sure that Neo4j and MinIO are using new storage class `openebs-lcmpv`:

   ```shell
   kubectl get pvc --all-namespaces
   ```

## Usage

1. Copy och-content to the main och-content directory

   ```bash
   cp -a och-content/. ../../../och-content/
   ```

1. Populate database

   ```bash
   kubectl -n neo4j port-forward svc/neo4j-neo4j 7687:7687

   APP_JSONPUBLISHADDR=http://{IP} APP_LOGGER_DEV_MODE=true APP_MANIFESTS_PATH=./och-content go run cmd/populator/main.go .
   ```

   New action is available now `cap.interface.data.backup`. Despite the name, you can use it to run backups and restores.

1. Deploy RocketChat

   Create and run action `cap.interface.productivity.rocketchat.install`. When installed, configure RocketChat,
   setup user, write some messages on chat.

   <details>
     <summary>GraphQL query</summary>

   ```graphql
   mutation CreateRocketChatAction {
     createAction(
       in: {
         name: "install-rocketchat",
         actionRef: {
           path: "cap.interface.productivity.rocketchat.install",
           revision: "0.1.0",
         },
         dryRun: false,
         advancedRendering: false,
         input: {
           parameters: "{\"host\":\"rocket.voltron.local\"}"
         }
      }
    ) {
      name
      input {
        parameters
      }
     }
   }
   
   query Get {
     action(name:"install-rocketchat"){
       name
       status{
         phase
         runner{
           status
         }
       }
     }
   }
   
   mutation Run {
       runAction(name: "install-rocketchat") {
           name
       }
   }
   
   mutation Delete {
       deleteAction(name: "install-rocketchat") {
           name
       }
   }
   ```
   </details>

1. Backup MongoDB

   Create a backup action. All it needs is a backup name and selector to find k8s objects.

   <details>
     <summary>GraphQL query</summary>

   ```graphql
   mutation CreateMongoDBBackup {
    createAction(
        in: {
            name: "backup-mongo",
            actionRef: {
                path: "cap.interface.data.backup",
                revision: "0.1.0",
            },
            advancedRendering: false,
            input: {
              parameters: "{\"action\":\"backup\",\"name\":\"mongo\", \"selector\":\"helm.sh/chart=mongodb-10.3.1\"}",
            }
        }
    ) {
        name
        input {
            parameters
        }
      }
   }

   query Get {
     action(name:"backup-mongo"){
       name

       status{
         phase
         runner{
           status
         }
       }
     }
   }
 
   mutation Run {
       runAction(name: "backup-mongo") {
           name
       }
   }
 
   mutation Delete {
      deleteAction(name: "backup-mongo") {
          name
      }
   }
   ```
   </details>

1. Delete MongoDB

   ```shell
   kubectl delete all -l helm.sh/chart=mongodb-10.3.1
   ```

   Delete also PVC for MongoDB.

1. Make sure that RocketChat is not working.

1. Restore MongoDB

   <details>
     <summary>GraphQL query</summary>

   ```graphql
   mutation RestoreMongoDBBackup {
    createAction(
        in: {
            name: "backup-mongo",
            actionRef: {
                path: "cap.interface.data.backup",
                revision: "0.1.0",
            },
            dryRun: false,
            advancedRendering: false,
            input: {
              parameters: "{\"action\":\"restore\",\"name\":\"mongo\"}",
            }
        }
    ) {
        name
        input {
            parameters
        }
      }
   }
 
   mutation Run {
       runAction(name: "backup-mongo") {
           name
       }
   }
   ```
   </details>

1. Wait few minutes and check RocketChat, it should be working again and all data should be available.

## Velero backup options

Velero can be used in 2 ways:

1. Full backup and restore.

   Backup all volumes, backup all k8s objects. Objects can be selected using selector.
   It's possible to backup only selected objectes from selected namespaces. It's also
   possible to skip some objects and namespaces.

1. Backup only some k8s objects

   Same as above but skipping volumes. This can be used to backup for example CRDs only.
   For example to backup Voltron actions.

Velero also allows to set [scheduled job](https://velero.io/docs/v1.5/api-types/schedule/#docs) for backups.

## Limitations:

* Cannot restore only data volumes. It always restores a pod which attached a volume https://github.com/vmware-tanzu/velero/issues/504
* Cannot do partial restores. It's all or nothing approach https://github.com/vmware-tanzu/velero/issues/904
* Can use only one selector
  If we want to backup for example Jira and PostgreSQL together we need to set them the same label first.
  Another option is to create two backups.
* Cannot overwrite existing objects. 
  This means that we need to delete the current objects to do a restore.
* When restic is used it doesn't backup hostPath volumes.
* No Point of Time backups. 
  It's up to user to block new writes or to block access to resources during the backup.
  Velero exposes some [hooks](https://velero.io/docs/v1.5/backup-hooks/) to help with this.
* Cannot restore OwnerReference and Status. Writing restore plugin is required.

## Security concerns

Velero requires super user powers. Depends on what will be included in a backup it may need to access all namespaces.
Backup/Restore CRDs are stored in velero namespace. Voltron action needs access to it.

## Voltron restore options:

Voltron is using following components:

- Voltron engine and OCH API 

  Stateless applications which don't need backups.
- Neo4j

  Database stores data on Persistent Volumes. Backup and Restore work properly.
- Argo and MinIO

  MinIO stores data on Persistent Volumes. Backup and Restore work properly.
- Voltron Actions

  Velero can restore CRDs but cannot restore Status and OwnerReference. Creating
  custom Restore Plugin is required.

Velero is not a perfect tool to support scenario "clean install and restore". It cannot overwrite existing objects, so we would need to delete some objects first.
Such approach will require deleting PVCs and databases(Neo4j and MinIO) during `Restore action`.

Much safer options would be to restore Voltron without installing it first.

### Backup and Restore

To speedup the process we will use `velero` CLI now.

1. Create a backup

   ```shell
   velero create backup voltron --include-namespaces=neo4j,argo,voltron-system --default-volumes-to-restic
   ```

1. Delete namespaces used by Voltron

   ```shell
   kubectl delete ns neo4j,voltron-system,argo
   ```

1. Restore Voltron

   ```shell
   velero restore create  --from-backup voltron
   ```

1. Verify

   * Gateway should be working again.
   * TypeInstances created during RocketChat installation should be still available

     <details>
     <summary>GraphQL query</summary>

     ```graphql
     query GetTypeInstances {
       typeInstances{
         id
         typeRef{
           path
           revision
         }
         latestResourceVersion{
           spec{
             value
           }
         }
       }
     }
     ```
     </details>

   * MinIO bucket should still have a data.

## Local development

Velero can backup Persistent Volumes using few ways(CSI volume-snapshot, plugins and restic). In kind cluster for persistent storage [local-path-provisioner](https://github.com/rancher/local-path-provisioner) is used. It [doesn't support](https://github.com/rancher/local-path-provisioner/issues/81) volume-snapshots and there is no plugin for it.
The only option is to use restic for backups. One of its limitation is that it doesn't support host-path volumes.
`local-path-provisioner` for PV is using host-path though. 
This means that in our dev cluster we have to use different storage provider.

I've tried few storage providers:

* OpenEBS did not start. It looks like an issue with NodeManager.
* Rook/Ceph needs empty, not formated disk to use. On Macs where virtual machine is used it may be easier to configure. It also installs many pods.
* [Longhorn](https://longhorn.io/) is using ISCSi which doesn't work in our docker. For working, it needs host network and privileged mode.

In the end I've used [lvm-local-pv](https://github.com/openebs/lvm-localpv) from OpenEBS. It's in alpha stage
and is not yet integrated with OpenEBS, but it can be run as a standalone project and works well. It has initial
support for volume-snapshots but not yet complete. For now, we can use `restic` to do backups.

## Helm charts support 

The idea is to pass helm release name and backup the whole release. There is
a plugin for Helm2 and updated version for Helm3 https://github.com/Project-Voltron/velero-plugin-helm,
but it looks like it still needs some fixes as it doesn't work.

## Summary

Most of the storage providers like OpenEBS and all CloudProvides have option to create a snapshot of a volume.
Most of them support new CSI volume-snapshot API.

Velero also supports(beta feature) volume-snapshots but goes beyond that. For storage providers  which
don't support snapshots and for other volume types it uses plugins and [restic](https://restic.net/). Besides that
it allows to backup other K8s objects which is a unique feature in free tools.
