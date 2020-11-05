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
   helm install ingress-nginx ./ingress-nginx --create-namespace -n=ingress-nginx
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

1. Upgrade Voltron Custom Resource Definitions:
    
   ```bash
   kubectl apply -f ./crds
   ``` 

1. Upgrade Voltron Helm chart:
    
    ```bash
    helm upgrade voltron ./charts/voltron -n voltron-system 
    ```

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
