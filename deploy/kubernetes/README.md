# Installing Capact on Kubernetes

Read this document to learn how to manage Capact installation on Kubernetes.

## Prerequisites

Before you begin, make sure you have the following tools installed:

- [kubectl](https://kubernetes.io/docs/tasks/tools/install-kubectl/)
- [Helm 3](https://helm.sh/docs/intro/install/)

## Install

To install Capact, run the following steps:

1. Create dedicated namespace:
   
   ```bash
   kubectl create namespace capact-system
   ```
   
1. Install Capact Custom Resource Definitions:
    
   ```bash
   kubectl apply -f ./crds
   ``` 

1. Install NGINX Ingress Controller:
    
    ```bash
   helm install ingress-nginx ./ingress-nginx -n capact-system
   ```

1. Install Argo Workflow:

    ```bash
   helm install argo ./argo -n capact-system
   ```

1. **[Optional]** To run Argo workflows in any namespace, follow these steps:

    1. Install kubed:

        ```bash
        helm install kubed ./charts/kubed -n capact-system 
        ``` 
   
   1. Annotate Minio secret to synchronize it to all namespaces:
       
       ```bash
       kubectl annotate secret -n capact-system argo-minio kubed.appscode.com/sync=""
       ```

1. **[Optional]** Install monitoring stack:

    ```bash
    helm install monitoring ./charts/monitoring -n capact-system
    ```
   
    > **NOTE:** This command installs the Prometheus and Grafana with default Kubernetes metrics exporters and Grafana dashboards.
    Installed Capact components configure automatically with monitoring stack by creating ServiceMonitor and dedicated Grafana dashboards.
    For more information check [instrumentation](https://capact.io/docs/development/development-guide#instrumentation) section.

1. **[Optional]** Install Cert Manager:

    ```bash
    helm install cert-manager ./charts/cert-manager -n capact-system
    ```

1. Install Capact Helm chart:
    
    ```bash
    helm install capact ./charts/capact -n capact-system
    ```

## Upgrade

> **NOTE:** Migration to a new major version of Capact release may require manual actions. Before upgrading to a new major version, read the release instructions.

To upgrade Capact installation, do the following steps:

1. Build CLI:

   ```bash
   # {OS} - possible values: linux, darwin, windows
   CLI_OS={OS} make build-tool-cli
   ```

2. Log into the cluster:

   ```bash
   # {OS} - same as in the first step
   ./bin/capact-{OS}-amd64 login {CLUSTER_GATEWAY_URL} -u {GATEWAY_USERNAME} -p {GATEWAY_PASSWORD}
   ```
   
3. Trigger cluster upgrade:

   ```bash
   # {OS} - same as in the first step
   # Upgrade Capact components to the newest available version
   ./bin/capact-{OS}-amd64 upgrade
   ```
   
   >**NOTE:** To check possible configuration options, run: `./bin/capact-{OS}-amd64 upgrade --help`
                 
## Uninstall

To uninstall Capact, follow the steps:

1. Uninstall Capact Helm charts:
    
    ```bash
    helm uninstall -n capact-system $(helm list -q -n capact-system)
    ```
1. Delete Capact namespace:

   ```bash
   kubectl delete namespace capact-system
   ```

1. Delete all Capact Custom Resource Definitions:
    
   ```bash
   kubectl delete crd actions.core.capact.io
   ``` 
