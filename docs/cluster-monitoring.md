# Monitoring of long-running GPC cluster

## Prerequisites

You need to have the following tools installed on your operating system:

- [`gcloud`](https://cloud.google.com/sdk/docs/install)
- [`jq`](https://stedolan.github.io/jq/download/) - most distributions have this in repositories
- [`kubectl`](https://kubernetes.io/docs/tasks/tools/)

## Setup access to GKE cluster

We are using the `voltron-dev3` GKE cluster in `europe-north1` region as our long-running Voltron cluster. Set the following environment variables in your shell:
```bash
export REGION=europe-north1
export CLUSTER_NAME=voltron-dev3
```

Get the kubeconfig for long-running Voltron GKE cluster:
```
gcloud container clusters get-credentials ${CLUSTER_NAME} --region ${REGION}
```

This command adds a new context to your local kubeconfig file. The name of the context will be in the format `gke_<project_name>_<region>_<cluster_name>`. Switch to the long-running Voltron cluster context:
```bash
$ kubectl config get-contexts
CURRENT   NAME                                                 CLUSTER                                              AUTHINFO                                             NAMESPACE
          gke_projectvoltron_europe-north1_voltron-dev3        gke_projectvoltron_europe-north1_voltron-dev3        gke_projectvoltron_europe-north1_voltron-dev3        
*         kind-kind-dev-voltron                                kind-kind-dev-voltron                                kind-kind-dev-voltron

$ kubectl config use-context gke_projectvoltron_europe-north1_voltron-dev3
Switched to context "gke_projectvoltron_europe-north1_voltron-dev3".
```

Now run the script to add your public IP address to the authorized control plane networks, so you will be able to make queries to the GKE API server:
```bash
./hack/monitoring/manage-ip.sh add
```

Now execute the following script to get Grafana credentials and setup port forwarding to Grafana:
```bash
$ ./hack/monitoring/grafana-forward.sh 
Username: *****
Password: *****
Forwarding from 127.0.0.1:3000 -> 3000
Forwarding from [::1]:3000 -> 3000
Handling connection for 3000
Handling connection for 3000
```

Open http://127.0.0.1:3000 and login into Grafana using the credentials printed out in the script.

## Check cluster health

1. Check the `Kubernetes/Computer Resources/Cluter` dashboard. Check, if the cluster has enough CPU and memory resources, by looking on the CPU and memory requests commitments. Also look on the CPU and memory usage graphs.
2. On the `Kubernetes/Compute Resources/Namespace (Pods)` dashboard check the resource usages for pods in all namespaces. Verify the pods have enough resources and do not experience issues like out of memory kills.

## Remove your IP from the authorized list

After you are done, run the following script to remove your IP from the authorized GKE control plane networks:
```bash
./hack/monitoring/manage-ip.sh remove
```
