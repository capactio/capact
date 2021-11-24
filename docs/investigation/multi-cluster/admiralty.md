# Admiralty with Capact showcase

This guide showcases how to create Capact cluster and connect it with another cluster via Admiralty. To simplify the tutorial, the direct Argo Workflow is used.

## Steps

### Bootstrapping

1. Create kind cluster for Capact:

    ```bash
    capact environment create kind --wait 5m
    ```

2. Install Capact on kind cluster:

    ```bash
    capact install --helm-repo @latest
    ```

3. Create workload cluster:

    ```bash
    kind create cluster --name eu
    ```

### Configuration

1. Label the workload cluster nodes (we'll use this label as node selectors):

    ```bash
    kubectl --context kind-eu label nodes --all topology.kubernetes.io/region=eu
    ```

2. Install cert-manager in workload cluster:

   > **NOTE:** Admiralty Open Source uses cert-manager to generate a server certificate for its mutating pod admission webhook. In Capact cluster, cert-manager is already installed.

    ```bash
    helm repo add jetstack https://charts.jetstack.io
    helm repo update
    
    kubectl --context kind-eu create namespace cert-manager
    kubectl --context kind-eu apply --validate=false -f https://github.com/jetstack/cert-manager/releases/download/v0.16.1/cert-manager.crds.yaml
    helm install cert-manager jetstack/cert-manager \
    --kube-context kind-eu \
    --namespace cert-manager \
    --version v0.16.1 \
    --wait
    ```
3. Install Admiralty in each cluster:

    ```bash
    helm repo add admiralty https://charts.admiralty.io
    helm repo update
    
    for CLUSTER_NAME in dev-capact eu
    do
      kubectl --context kind-$CLUSTER_NAME create namespace admiralty
      helm install admiralty admiralty/multicluster-scheduler \
        --kube-context kind-$CLUSTER_NAME \
        --namespace admiralty \
        --version 0.14.1 \
        --wait --debug
      # --wait to ensure release is ready before next steps
      # --debug to show progress, for lack of a better way,
      # as this may take a few minutes
    done
    ```

4. Creates kubeconfigs from the ServiceAccount tokens:

    ```bash
    # i. create a Kubernetes service account in the workload cluster for the management cluster,
    kubectl --context kind-eu create serviceaccount cd
    
    # ii. extract its default token,
    SECRET_NAME=$(kubectl --context kind-eu get serviceaccount cd \
      --output json | \
      jq -r '.secrets[0].name')
    TOKEN=$(kubectl --context kind-eu get secret $SECRET_NAME \
      --output json | \
      jq -r '.data.token' | \
      base64 --decode)
    
    # iii. get a Kubernetes API address that is routable from the management cluster—here, the IP address of the kind workload cluster's only (master) node container in your machine's shared Docker network,
    IP=$(docker inspect eu-control-plane \
      --format "{{ .NetworkSettings.Networks.kind.IPAddress }}")
    
    # iv. prepare a kubeconfig using the token and address found above, and the server certificate from your kubeconfig (luckily also valid for this address, not just the address in your kubeconfig),
    CONFIG=$(kubectl --context kind-eu config view \
      --minify --raw --output json | \
      jq '.users[0].user={token:"'$TOKEN'"} | .clusters[0].cluster.server="https://'$IP':6443"')
    
    # v. save the prepared kubeconfig in a secret in the management cluster:
    kubectl --context kind-dev-capact create secret generic eu \
      --from-literal=config="$CONFIG"
    ```

5. In the Capact cluster, create a Target for workload cluster:

    ```yaml
    cat <<EOF | kubectl --context kind-dev-capact apply -f -
    apiVersion: multicluster.admiralty.io/v1alpha1
    kind: Target
    metadata:
      name: eu
    spec:
      kubeconfigSecret:
        name: eu
    EOF
    ```

7. In the workload clusters, create a Source for the Capact cluster:

     ```yaml
     cat <<EOF | kubectl --context kind-$CLUSTER_NAME apply -f -
     apiVersion: multicluster.admiralty.io/v1alpha1
     kind: Source
     metadata:
       name: cd
     spec:
       serviceAccountName: cd
     EOF
     ```

8. Check that virtual nodes have been created in the Capact cluster to represent workload cluster:

    ```bash
    kubectl --context kind-dev-capact get nodes
    ```

9. Label the default Namespace in the Capact cluster to enable multi-cluster scheduling at the namespace level:

    ```bash
    kubectl --context kind-dev-capact label ns default multicluster-scheduler=enabled
    ```

## Demo

1. Create Argo Workflow in Capact cluster, targeting the `eu` workload cluster:

    ```yaml
    cat <<EOF | kubectl --context kind-dev-capact apply -f -
    apiVersion: batch/v1
    kind: Job
    metadata:
      name: eu-foo
      namespace: default
    spec:
      template:
        metadata:
          annotations:
            multicluster.admiralty.io/elect: ""
        spec:
          nodeSelector:
            topology.kubernetes.io/region: eu
          containers:
          - name: c
            image: busybox
            command: ["sh", "-c", "echo Processing item foo && sleep 5"]
            resources:
              requests:
                cpu: 100m
          restartPolicy: Never
    EOF
    ```

2. Watch the Argo Workflow execution:

    ```bash
    argo watch eu-foo
    ```

3. Check Argo Workflow logs:

    ```bash
    argo log eu-foo
    ```

4. Check that the pod was scheduled on `eu` workload node:

    ```bash
    kubectl --context kind-eu get pods -o wide -n default
    ```

