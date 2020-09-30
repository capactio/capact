# Argo Workflows investigation

The document describes various experiments with Argo Workflows. The experiments are helpful for future development of Built-in Runner based on Argo.

## Prerequisites

- Minikube
- Docker

## Usage

In order to run Minikube with Argo and MinIO installed, execute the following script:

```bash
./run.sh
```

After successful installation you can run experiments described below.

### Argo UI

To see all executed workflows in UI, expose the Argo UI:

```bash
kubectl -n argo port-forward deployment/argo-server 2746:2746
```

To access exposed UI, navigate to the [localhost:2746](http://localhost:2746).

## Experiments

### Artifacts

This is the default Argo example with artifacts. It's a good start to observe how the workflow artifacts are stored.

To execute the workflow, run the following command:

```bash
kubectl apply -n argo -f ./workflows/artifacts.yaml 
```

Expose MinIO UI:

```bash
kubectl port-forward minio -n argo 9000:9000
```

Access the [MinIO UI](http://localhost:9000). Log in with `admin/password` credentials, navigate to `my-bucket` and observe how the artifacts are stored.

### Nested workflows

#### Different workflow depth level

Create nested workflows with different depth levels:

```bash
kubectl apply -n argo -f ./workflows/nested2.yaml
kubectl apply -n argo -f ./workflows/nested3.yaml
kubectl apply -n argo -f ./workflows/nested10.yaml
```

Wait for the workflows to finish:
```bash
kubectl get workflow -n argo -w
```

You can use [Argo UI](#argo-ui) to observe workflow execution results.

#### Passing input from depth level 1 to depth level 3 

Observe the behavior when nested workflow tries to read input from parent workflow:

```bash
kubectl apply -n argo -f ./workflows/input-different-depth-lvl.yaml
```

#### Infinity loop

Observe the behavior when Argo Workflow contains infinity loop of workflows:

```bash
kubectl apply -n argo -f ./workflows/infinity-loop.yaml 
```

## Cleanup

To delete the Minikube cluster, run the following command:
```bash
minikube delete
```

## Findings

### Artifacts

- Artifacts are stored in TAR archive, compressed with gzip. They can be not only files, but also directories. Artifact is saved under `{bucket}/{workflow_name}/{pod-name-which-saves-artifact}/{artifact-name}.tgz` path.
- When deleting Argo Workflow, artifacts on MinIO [are not cleaned up](https://github.com/argoproj/argo/issues/3390).

### Logs

- Argo stores in MinIO not only artifacts, but also logs from all pods from a given workflow. They are available under `{bucket}/{workflow_name}/{pod-name}/main.log` path.

### Nested Workflows

- Nested workflows workes fine, at least for depth up to 10. There is no strict depth limit in Argo documentation.
- If a given nested workflow just passes input and output, no Pods are scheduled for the given depth level. For the workflow with depth 10, 4 containers are scheduled - just the ones user specified in workflow.
    
  **NOTE:** In case we switch from Argo in the future to more generic approach, we would need to schedule as many containers as the nested workflow depth, as every nested workflow will be a separate workflow.

- Inputs and outputs are scoped to a given template (that is, given workflow depth). For example, passing input from Workflow depth level 1 to Workflow depth level 3 is not possible.
- Argo Workflow Controller doesn't detect infinity loop in Workflows and it crashes while running such workflow.
