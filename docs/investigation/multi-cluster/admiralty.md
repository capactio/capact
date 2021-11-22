```bash
capact environment create k3d --wait 5m
```

```bash
capact install --helm-repo @latest
```

```bash
for CLUSTER_NAME in us eu
do
  kind create cluster --name $CLUSTER_NAME
done
```

```bash
kubectl --context k3d-dev-capact create namespace admiralty

helm install admiralty admiralty/multicluster-scheduler \
--kube-context k3d-dev-capact \
--namespace admiralty \
--version 0.14.1 \
--wait --debug
```


```bash
for CLUSTER_NAME in us eu
do
  # i. create a Kubernetes service account in the workload cluster for the management cluster,
  kubectl --context kind-$CLUSTER_NAME create serviceaccount cd

  # ii. extract its default token,
  SECRET_NAME=$(kubectl --context kind-$CLUSTER_NAME get serviceaccount cd \
    --output json | \
    jq -r '.secrets[0].name')
  TOKEN=$(kubectl --context kind-$CLUSTER_NAME get secret $SECRET_NAME \
    --output json | \
    jq -r '.data.token' | \
    base64 --decode)

  # iii. get a Kubernetes API address that is routable from the management cluster—here, the IP address of the kind workload cluster's only (master) node container in your machine's shared Docker network,
  IP=$(docker inspect $CLUSTER_NAME-control-plane \
    --format "{{ .NetworkSettings.Networks.kind.IPAddress }}")

  # iv. prepare a kubeconfig using the token and address found above, and the server certificate from your kubeconfig (luckily also valid for this address, not just the address in your kubeconfig),
  CONFIG=$(kubectl --context kind-$CLUSTER_NAME config view \
    --minify --raw --output json | \
    jq '.users[0].user={token:"'$TOKEN'"} | .clusters[0].cluster.server="https://'$IP':6443"')

  # v. save the prepared kubeconfig in a secret in the management cluster:
  kubectl --context k3d-dev-capact create secret generic $CLUSTER_NAME \
    --from-literal=config="$CONFIG"
done
```

```bash
for CLUSTER_NAME in us eu
do
  cat <<EOF | kubectl --context k3d-dev-capact apply -f -
apiVersion: multicluster.admiralty.io/v1alpha1
kind: Target
metadata:
  name: $CLUSTER_NAME
spec:
  kubeconfigSecret:
    name: $CLUSTER_NAME
EOF
done
```

```bash
for CLUSTER_NAME in us eu
do
  cat <<EOF | kubectl --context kind-$CLUSTER_NAME apply -f -
apiVersion: multicluster.admiralty.io/v1alpha1
kind: Source
metadata:
  name: cd
spec:
  serviceAccountName: cd
EOF
done
```

```bash
kubectl --context k3d-dev-capact get nodes --watch
# --watch until virtual nodes are created,
# this may take a few minutes, then control-C
```

```bash
kubectl --context k3d-dev-capact label ns default multicluster-scheduler=enabled
```

```bash
cat <<EOF | kubectl --context k3d-dev-capact apply -f -
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

while true
do
clear
for CLUSTER_NAME in cd us eu
do
kubectl --context kind-$CLUSTER_NAME get pods -o wide
done
sleep 2
done
# control-C when all pods have Completed
