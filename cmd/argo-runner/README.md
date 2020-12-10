# Argo runner

## Overview

Argo runner is a [Voltron workflow runner](../../docs/runner.md), which executes Argo workflows

## Prerequisites

- Running Kubernetes cluster with Argo installed
- kubectl
- Go compiler 1.14+

## Usage

Normally the Argo runner is started by the Voltron Engine, but you can run the runner locally without the Engine.

### Run Argo runner locally

The following steps show, how to execute an Argo workflow with the runner without the Voltron engine.

1. Ensure the Serivce Account used by the Argo workflow has proper RBAC permissions. Example command to add permissions for using default service account in default namespace:
```bash
kubectl create clusterrolebinding default-default-admin --clusterrole admin --serviceaccount default:default
```

2. Create a dummy job to use for ownerReference for the workflow

```bash
kubectl create job dummy --image alpine
```

3. Create secret of the runners status reporter
```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: dummy
EOF
```

4. Create runner input file
```bash
cat <<EOF > argo-args.yaml
context:
  name: dummy
  dryRun: false
  platform:
    namespace: default
    ownerRef:
      apiVersion: batch/v1
      kind: Job
      name: dummy
      uid: $(kubectl get jobs dummy -ojsonpath='{.metadata.uid}')
args:
  workflow:
    entrypoint: main
    templates:
      - name: main
        container:
          image: docker/whalesay
          command: ["/bin/bash", "-c"]
          args: ["sleep 2 && cowsay hello world"]
EOF

export RUNNER_INPUT_PATH=argo-args.yaml
```

5. Run the runner:
```bash
go run cmd/argo-runner/main.go
```

6. Check the workflow status. You can use the Argo UI on http://localhost:2746, after you forward the ports:
```bash
kubectl port-forward -n argo svc/argo-server 2746
```
