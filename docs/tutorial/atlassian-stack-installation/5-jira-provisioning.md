# 5. Provision Jira

Follow these steps to install Atlassian Jira using existing PostgreSQL installation.

1. Set the Jira license key, domain name under which Jira should be available and PostgresSQL AWS RDS TypeInstance ID enviornment variables:
    ```bash
    export JIRA_LICENSE_KEY="<your-jira-license-key>"
    export DOMAIN_NAME="<your-domain-name>" # e.g. demo.cluster.capact.dev
    export POSTGRESQL_TI="<postgresql-ti-from-rds-provisioning>"
    ```
    
1. Create a file with parameters:
    ```bash
    cat > /tmp/jira-params.yaml << ENDOFFILE
    replicaCount: 1
    ingress:
      host: "jira.${DOMAIN_NAME}"
    volumes:
      sharedHome:
        persistentVolumeClaim:
          storageClassName: efs-sc
          resources:
            requests:
              storage: 25Gi
      localHome:
        persistentVolumeClaim:
          resources:
            requests:
              storage: 25Gi
    resources:
      jvm:
        maxHeap: 2g
        minHeap: 512m
    jira:
      licenseKeyInBase64: "${JIRA_LICENSE_KEY}"
    ENDOFFILE
    ```

> **NOTE:** It is recommended to start with a single Jira replica before finishing the configuration process.

1. Create file a with TypeInstance input:
    ```bash
    cat > /tmp/jira-tis.yaml << ENDOFFILE
    typeInstances:
      - name: "postgresql"
        id: "${POSTGRESQL_TI}"
    ENDOFFILE
    ```

1. Create a Kubernetes Namespace for Jira:
    ```bash
    kubectl create namespace jira
    ```

1. Create an Action:
    ```bash
    capact act create cap.interface.productivity.jira.install --name jira --namespace jira --parameters-from-file /tmp/jira-params.yaml --type-instances-from-file /tmp/jira-tis.yaml
    ```

1. Run the Action:
    ```bash
    capact act run jira --namespace jira
    ```

1. Watch the Action:
    ```bash
    capact act watch jira --namespace jira
    ```

1. Once the Action succeeded, open Jira in your browser. It should be available under `https://jira.${DOMAIN_NAME}`.
 
1. Setup the Jira installation. This can take a few minutes. If you get an `503 Service Unavailable` error, refresh the page and continue.

1. (Optional) After Jira is configured and you are able to login, you can scale out the Jira StatefulSet:
    ```bash
    kubectl -n jira scale statefulsets.apps -l app.kubernetes.io/name=jira --replicas 2
    ```

    After a few minutes, the second node should appear on the `Jira Administration -> System -> Clustering` page.

    ![jira-clustering-dashboard](./assets/3-jira-clustering-dashboard.png)
  