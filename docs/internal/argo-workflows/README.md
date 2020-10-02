# Argo Workflows investigation

The document describes various experiments with Argo Workflows. The experiments are helpful for future development of built-in Runner based on Argo.

## Prerequisites

- [Minikube](https://kubernetes.io/docs/setup/learning-environment/minikube/)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation) for running [Argo on kind](#argo-on-kind) experiment

## Usage

Install Argo and MinIO stable version on Minikube. After successful installation you can run experiments described in the document.

In order to run Minikube with Argo and MinIO installed, execute the following script:

```bash
./run.sh
```

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
kubectl apply -n argo -f ./experiments/artifacts.yaml 
```

The workflow succeeds. Expose MinIO UI:

```bash
kubectl port-forward minio -n argo 9000:9000
```

Access the [MinIO UI](http://localhost:9000). Log in with `admin/password` credentials, navigate to `my-bucket` and observe how the artifacts are stored.

### Nested workflows

#### Different workflow depth level

Create nested workflows with different depth levels:

```bash
kubectl apply -n argo -f ./experiments/nested2.yaml
kubectl apply -n argo -f ./experiments/nested3.yaml
kubectl apply -n argo -f ./experiments/nested10.yaml
kubectl apply -n argo -f ./experiments/nested15.yaml
```

Wait for the workflows to finish:
```bash
kubectl get workflow -n argo -w
```

All workflows succeed. You can use [Argo UI](#argo-ui) to observe workflow execution results.

#### Global input: Passing input from depth level 1 to depth level 3 

Observe the behavior when nested workflow tries to read input from parent workflow:

```bash
kubectl apply -n argo -f ./experiments/global-input.yaml
```

The workflow fails with message `unable to resolve references: Unable to resolve: {{steps.generate1.outputs.artifacts.out-artifact}}`, because inputs and outputs are scoped to a given template.

#### Global output: Passing output from depth level 3 to depth level 1 

Observe the behavior when workflow tries to read input from nested workflow output without :

```bash
kubectl apply -n argo -f ./experiments/global-output.yaml
```

The workflow succeeds, accessing global parameter and artifact from different nested step.

#### Infinite loop

Observe the behavior when Argo Workflow contains infinite loop of workflows:

```bash
kubectl apply -n argo -f ./experiments/infinite-loop.yaml 
```

The Workflow Controller crashes with exit code 137 (Reason: Error), which means it receives SIGKILL signal. Issue for Argo is [already reported](https://github.com/argoproj/argo/issues/4180), as the Workflow Controller should detect infinite loop and fail fast.

### Argo on kind

In order to run kind with Argo and MinIO installed, execute the following script:

```bash
./experiments/argo-kind/run.sh
```

To make Argo work on kind, the [`k8sapi` Workflow Executor](https://argoproj.github.io/argo/workflow-executors) is used. 

Run [Artifacts](#artifacts) workflow experiment example and observe how it fails with error `kubelet executor does not support outputs from base image layer. must use emptyDir`.

To make it succeeded, run modified Artifacts experiment:

```bash
kubectl apply -n argo -f ./experiments/artifacts-volumes.yaml 
```

Every Argo workflow would need to use container volumeMounts to be able to run on kind with `k8sapi` executor.

## Cleanup

To delete the Minikube cluster, run the following command:
```bash
minikube delete
```

If you ran the [Argo on kind](#argo-on-kind) experiment, run the following command:
```bash
kind delete cluster
```

## Findings

### Artifacts

- Artifacts are stored in TAR archive, compressed with gzip. They can be not only files, but also directories. Artifact is saved under `{bucket}/{workflow_name}/{pod-name-which-saves-artifact}/{artifact-name}.tgz` path.
- When deleting Argo Workflow, artifacts on MinIO [are not cleaned up](https://github.com/argoproj/argo/issues/3390).
- There is no way to take all global artifacts from a workflow as an input to a step. To achieve uploading artifacts to OCH, we may access MinIO `{bucket}/{workflow_name}` directory and find all artifacts (`.tgz` files) in subdirectories.

### Logs

- Argo stores in MinIO not only artifacts, but also logs from all pods from a given workflow. They are available under `{bucket}/{workflow_name}/{pod-name}/main.log` path.

### Nested Workflows

- Nested workflows work fine, at least for depth up to 10. In Argo template resolving logic [there is a code which should limit nested template references to 10](https://github.com/argoproj/argo/blob/06c4bd60cf2dc85362b3370acd44e4bc3977dcbc/workflow/templateresolution/context.go#L194). Even if the experiment with nested 15 workflows works fine, it may be probably a bug. Issue for Argo is [already reported](https://github.com/argoproj/argo/issues/4180).
- If a given nested workflow just passes input and output, no Pods are scheduled for the given depth level. For the workflow with depth 10, 4 containers are scheduled - just the ones user specified in workflow.
    
  **NOTE:** In case we switch from Argo in the future to more generic approach, we would need to schedule as many containers as the nested workflow depth, as every nested workflow will be a separate workflow.

- Inputs and outputs are scoped to a given template (that is, given workflow depth). Passing input from Workflow depth level 1 to Workflow depth level 3 is not possible. However, you can expose an output parameter or artifact to global scope with [`globalName`](https://argoproj.github.io/argo/swagger/#ioargoprojworkflowv1alpha1artifact) property.
- Argo Workflow Controller doesn't detect infinity loop in Workflows and it crashes while running such workflow.

### Others

- [Argo doesn't work on Kind with default Docker executor](https://github.com/argoproj/argo/issues/2376). It is because [kind uses containerd](https://github.com/kubernetes-sigs/kind/issues/508#issuecomment-490745016). The workaround is to use different [Argo Workflow Executor](https://argoproj.github.io/argo/workflow-executors) - `k8sapi`. However, the workflows would need to be adjusted, as with `k8sapi` executor output artifacts can only be saved on volumes (such as `emptyDir`). Modified workflow with `emptyDir` volume volumeMounts for containers works on both Workflow Executors.

  ```yaml
  message: 'invalid spec: templates.artifact-example.steps[0].generate-artifact templates.whalesay.outputs.artifacts.hello-art:
     k8sapi executor does not support outputs from base image layer. must use emptyDir'
  phase: Failed
   ```
- Argo is able to run [Cron Workflows](https://argoproj.github.io/argo/cron-workflows/), which may be useful in future for us as there are some plans to have cronjob-like Actions.
- There is a pattern of running [Workflow of workflows](https://argoproj.github.io/argo/workflow-of-workflows/), however at this point it doesn't seem to bring any value for us. This is what we will do, but in a generic way using dedicated containers.