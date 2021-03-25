# Backup and Restore actions for Voltron

This proof of concept shows [Velero](https://velero.io) as a tool used for running backup and restore
for workloads deployed by Voltron.

## Prerequisites
- Deployed Voltron dev cluster with disabled populator
- Installed locally [Velero CLI](https://velero.io/docs/v1.5/basic-install/)

## Velero installation

For Object Store we will just use existing terraform bucket in MinIO.

Create `credentials-velero` file with a following content:

```yaml
[default]
aws_access_key_id = MINIO_SECRET_KEY
aws_secret_access_key = MINIO_ACCESS_KEY
```

and install Velero

```shell
./velero install --provider aws --bucket=terraform --secret-file ./credentials-velero --backup-location-config region=minio,s3ForcePathStyle="true",s3Url=http://argo-minio.argo.svc:9000 --plugins velero/velero-plugin-for-aws:v1.0.0 --use-volume-snapshots=false --use-restic
```

Velero stores logs in object store. To see logs locally we need access to `argo-minio.argo.svc`. The easiest way
is to add it to `/etc/hosts` and to enable port forwarding:

```shell
kubectl -n argo port-forward service/argo-minio 9000
```

## Usage

1. Copy och-content to the main och-content directory

   ```bash
   cp -a och-content/. ../../../och-content/
   ```

   and populate database. 

   Now new action is available `cap.interface.data.backup`. Despite the name, you can use it to run backups and restores.

2. Deploy RocketChat

   Create and run action `cap.interface.productivity.rocketchat.install`. When installed, configure RocketChat,
   setup user, write some messages on chat.

3. Backup MongoDB

   Create a backup action. All it needs is backup name and selector to find k8s objects.

   ```graphql
   mutation CreateMongoDBBackup {
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
   ```

4. Delete MongoDB

   ```shell
   kubectl delete all -l helm.sh/chart=mongodb-10.3.1
   ```

5. Make sure that RocketChat is not working.

6. Restore MongoDB

   ```graphql
   mutation CreateMongoDBBackup {
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
   ```

7. Wait few minutes and check RocketChat, it should be working again and all data should be available.

## Velero backup options

Velero can be used in 2 ways:

1. Full backup and restore.

   Backup all volumes, backup all k8s objects. Objects can be selected using selector.
   It's possible to backup only selected objectes from selected namespaces. It's also
   possible to skip some objects and namespaces.

2. Backup only some k8s objects

   Same as above but skipping volumes. This can be used to backup for example CRDs only.
   For example to backup Voltron actions.

Velero also allows to set [scheduled job](https://velero.io/docs/v1.5/api-types/schedule/#docs) for backups.

## Limitations:

* Can not restore only data volumes. It always restores a pod which attached a volume https://github.com/vmware-tanzu/velero/issues/504
* Can not do partial restores. It's all or nothing approach https://github.com/vmware-tanzu/velero/issues/904
* Can use only one selector
  If we want to backup for example Jira and PostgreSQL together we need to set them the same label first.
  Another option is to create two backups.
* Can not overwrite existing objects. 
  This means that we need to delete the current objects to do a restore.
* When restic is used it doesn't backup hostpath volumes.
* No Point of Time backups. 
  It's up to user to block new writes or to block access to resources during the backup.
  Velero exposes some [hooks](https://velero.io/docs/v1.5/backup-hooks/) to help with this.

## Security concerns

Velero requires super user powers. Depends on what will be included in a backup it may need to access all namespaces.
Backup/Restore CRDs are stored in velero namespace. Voltron action needs access to it.

## Voltron restore options:

Velero is not a perfect tool to support scenario "clean install and restore". It can not
overwrite existing objects, so we would need to delete some objects first.
Such approach will require deleting PVCs and databases(Neo4j and MinIO) during `Restore action`.

Much safer options would be to restore Voltron without installing it first.

## Local development

Velero can backup Peristent Volumes using CSI volume-snapshots. In kind cluster for persistent storage [local-path-provisioner](https://github.com/rancher/local-path-provisioner) is used. It [doesn't support](https://github.com/rancher/local-path-provisioner/issues/81) volume-snapshots.
As a second option Velero can use restic for backups. One of its limitation is that it doesn't support host-path volumes.
`local-path-provisioner` for PV is using host-path though. 
This means that in our dev cluster we can not backup PV.

I've tried few storage providers but none of them worked:

* OpenEBS was complaining about NVMe partition. This may be related to my laptop only.
* Rook/Ceph needs empty, not formated disk to use. On Macs where virtual machine is used it may be easier to configure.
* [Longhorn](https://longhorn.io/) is using ISCSi which doesn't work in our docker. For working, it needs host network and privileged mode.


## Helm charts support 

The idea is to pass helm release name and backup the whole release. There is
a plugin for Helm2 and updated version for Helm3 https://github.com/lukaszo/velero-plugin-helm
but it looks like it still needs some fixes as it doesn't work.

## Summary

Most of the storage providers like OpenEBS and all CloudProvides have option to create a snapshot of a volume.
Most of them support new CSI volume-snapshot API.

Velero also supports(beta feature) volume-snapshots but goes beyond that. For storage providers  which
don't support snapshots and for other volume types it uses [restic](https://restic.net/). Besides that
it allows to backup other K8s objects which is a unique feature in free tools.
