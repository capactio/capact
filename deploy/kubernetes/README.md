# Installing Voltron on Kubernetes

Read this document to learn how to manage Voltron installation on Kubernetes.

## Prerequisites

Before you begin, make sure you have the following tools installed:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Helm 3](https://helm.sh/docs/intro/install/)

## Install

To install Voltron, run the following steps:

1. Install Voltron Custom Resource Definitions:
    
   ```bash
   kubectl apply -f ./crds
   ``` 

1. Install NGINX Ingress Controller:
    
    ```bash
   helm install ingress-nginx ./ingress-nginx --create-namespace -n ingress-nginx
   ```

1. Install Argo Workflow:

    ```bash
   helm install argo ./argo --create-namespace -n argo
   ```

1. **[Optional]** To run Argo workflows in any namespace, follow these steps:

    1. Install kubed:

        ```bash
        helm install kubed ./charts/kubed --create-namespace -n kubed 
        ``` 
   
   1. Annotate Minio secret to synchronize it to all namespaces:
       
       ```bash
       kubectl annotate secret -n argo argo-minio kubed.appscode.com/sync=""
       ```

1. **[Optional]** Install monitoring stack:

    ```bash
    helm install monitoring ./charts/monitoring --create-namespace -n monitoring
    ```
   
    > **NOTE:** This command installs the Prometheus and Grafana with default Kubernetes metrics exporters and Grafana dashboards.
    Installed Voltron components configure automatically with monitoring stack by creating ServiceMonitor and dedicated Grafana dashboards.
    For more information check [instrumentation](../../docs/development.md#instrumentation) section.

1. Install Voltron Helm chart:
    
    ```bash
    helm install voltron ./charts/voltron --create-namespace -n voltron-system
    ```

## Upgrade

> **NOTE:** Migration to a new major version of Voltron release may require manual actions. Before upgrading to a new major version, read the release instructions.

To upgrade Voltron installation, do the following steps:

1. Build CLI:

   ```bash
   # {OS} - possible values: linux, darwin, windows
   CLI_OS={OS} make build-tool-ocftool
   ```

2. Log into the cluster:

   ```bash
   # {OS} - same as in the first step
   ./bin/ocftool-{OS}-amd64 login {CLUSTER_GATEWAY_URL} -u {GATEWAY_USERNAME} -p {GATEWAY_PASSWORD}
   ```
   
3. Trigger cluster upgrade:

   ```bash
   # {OS} - same as in the first step
   # Upgrade Capact components to the newest available version
   ./bin/ocftool-{OS}-amd64 upgrade
   ```
   
   >**NOTE:** To check possible configuration options, run: `./bin/ocftool-{OS}-amd64 upgrade --help`
                 
## Uninstall

To uninstall Voltron, follow the steps:

1. Uninstall Voltron Helm chart:
    
    ```bash
    helm uninstall voltron -n voltron-system
    ```

1. Delete all Voltron Custom Resource Definitions:
    
   ```bash
   kubectl delete crd actions.core.projectvoltron.dev
   ``` 
