# Argo runner

Argo runner is a runner, which executes Argo workflows

## Supported features:

- executing Argo workflows

## How to build

```bash
make build-tool-argo-runner
# or
go build -o bin/argo-runner cmd/argo-runner/main.go
```

## How to use

1. Setup the dev KinD cluster
```bash
make dev-cluster
# or
make dev-cluster-update
```

2. Ensure the Serivce Account used by the Argo workflow has proper RBAC permissions. Ex for using default namespace and default service account:
```bash
kubectl create clusterrolebinding default-default-admin --clusterrole admin --serviceaccount default:default
```

3. Create a dummy job to use for ownerReference for the workflow

```bash
kubectl create job dummy --image alpine
```

4. Create secret of the runners status reporter
```bash
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: Secret
metadata:
  name: dummy
EOF
```

5. Create runner input file
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

6. Run the runner:
```bash
go run cmd/argo-runner/main.go
```

7. Check the workflow status. You can use the Argo UI on http://localhost:2746, after you forward the ports:
```bash
kubectl port-forward -n argo svc/argo-server 2746
```

## Hacking

Main source code is in:
- `cmd/argo-runner/` - binary main
- `pkg/runner/argo/` - manifest validation SDK
