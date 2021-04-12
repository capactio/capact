# Monitoring of long-running GCP cluster

## Table of contents
<!-- toc -->
- [Prerequisites](#prerequisites)
- [Setup access to GKE cluster](#setup-access-to-gke-cluster)
- [Check cluster health](#check-cluster-health)
- [Remove your IP from the authorized list](#remove-your-ip-from-the-authorized-list)
<!-- tocstop -->

## Prerequisites

You need to have the following tools installed on your operating system:

- [`gcloud`](https://cloud.google.com/sdk/docs/install)
- [`jq`](https://stedolan.github.io/jq/download/) - most distributions have this in repositories
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/)

You need to configure the `gcloud` CLI, so it's able to access `projectvoltron` project on GCP. You can follow [this](https://cloud.google.com/sdk/docs/authorizing) guide to configure it.

## Setup access to GKE cluster

We are using the `capact-dev` GKE cluster in `europe-north1` region for our long-running Capact cluster. Set the following environment variables in your shell:
```bash
export REGION=europe-north1
export CLUSTER_NAME=capact-dev
```

Get the kubeconfig for the long-running Capact GKE cluster:
```
gcloud container clusters get-credentials ${CLUSTER_NAME} --region ${REGION}
```

This command adds a new context to your local kubeconfig file. The name of the context will be in the format `gke_<project_name>_<region>_<cluster_name>`. Switch to the long-running Capact cluster context:
```bash
kubectl config get-contexts
```
```bash
CURRENT   NAME                                                 CLUSTER                                              AUTHINFO                                             NAMESPACE
          gke_projectvoltron_europe-north1_capact-dev        gke_projectvoltron_europe-north1_capact-dev        gke_projectvoltron_europe-north1_capact-dev        
*         kind-kind-dev-capact                                kind-kind-dev-capact                                kind-kind-dev-capact
```
```bash
kubectl config use-context gke_projectvoltron_europe-north1_capact-dev
```

Now run the script to add your public IP address to the authorized control plane networks, so you will be able to make queries to the GKE API server:
```bash
./hack/monitoring/manage-ip.sh add
```

Now execute the following script to get Grafana credentials and setup port forwarding to Grafana:
```bash
./hack/monitoring/grafana-forward.sh 
```
```bash
Username: *****
Password: *****
Forwarding from 127.0.0.1:3000 -> 3000
Forwarding from [::1]:3000 -> 3000
Handling connection for 3000
Handling connection for 3000
```

This script will run in the foreground in your shell during the port-forwarding.

Open [http://127.0.0.1:3000](http://127.0.0.1:3000) and login into Grafana using the credentials printed out in the script.

## Check cluster health

1. Open the `Kubernetes/Computer Resources/Cluster` dashboard. Check, if the cluster has enough CPU and memory resources, by looking on the CPU and memory requests commitments. Also look on the CPU and memory usage graphs.
2. On the `Kubernetes/Compute Resources/Namespace (Pods)` dashboard check the resource usages for pods in all namespaces. Verify the pods have enough resources and do not experience issues like out of memory kills.

## Remove your IP from the authorized list

Once you are done, run the following script to remove your IP from the authorized GKE control plane networks:
```bash
./hack/monitoring/manage-ip.sh remove
```
