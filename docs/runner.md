# Runner

Runner handles the execution of the Action according to the `action` property, which defines what should be executed. There are multiple runners, such as Helm Runner, CloudSQL Runner, Argo Workflow Runner.

## Overview

Each Engine implementation needs to have at least one built-in runner. The built-in runner has only the Interface and no Implementation definition. To ensure that, you need to add the `spec.abstract: true` property in the Interface that describes the runner.

The [Kubernetes Engine](../cmd/k8s-engine) uses Argo Workflow as the built-in runner. The [cap.interface.runner.argo](../och-content/interface/runner/argo/run.yaml) Interface defines the schema for the runner arguments. The Argo Workflow was selected, as it is a Kubernetes-native implementation, and it supports passing data between steps.

In the future, we will abstract all the built-in runners under a common Interface by using the [OCI Image](https://github.com/opencontainers/image-spec) and [OCI Runtime](https://github.com/opencontainers/runtime-spec) specifications. Each Engine implementation will have to fulfill this OCI runner Interface.

The goal for the runner concept is not to limit you to built-in runners only but also to give you the ability to define and use different runners. You can define a different runner such as Helm Runner. First, you need to create an Interface definition to describe possible input and output values.

```yaml
ocfVersion: 0.0.1
kind: Interface
metadata:
  prefix: cap.interface.runner.helm
  name: run
# ...
spec:
  input:
    parameters:
      jsonSchema:
        value: |-
          {
            "$schema": "http://json-schema.org/draft-07/schema",
            "type": "object",
            "properties": {
              "name": {
                "type": "string",
                "title": "The name of installed release"
              },
              "namespace": {
                "type": "string",
                "title": "The namespace schema",
              },
           }
  output:
    typeInstances:
      helm-release:
        typeRef:
          path: cap.type.helm.chart.release
          revision: 0.1.0
```

Next, you need to create at least one Implementation. To do so, you need to create an OCI image with a binary inside. The binary should read mounted data and execute a given functionality. For example, for Helm Runner it will be `helm install`, `helm delete`, etc.

The custom runner Implementation manifest should be described using one of the built-in runners, so Engine knows how to mount rendered input data and how to run the OCI image. For Argo Workflow built-in Runner, the Helm Runner manifest could look like this:

```yaml
ocfVersion: 0.0.1
kind: Implementation
metadata:
  prefix: cap.implementation.runner.helm
  name: run
# ...
action:
  runnerInterface: cap.interface.runner.argo
  args:
    workflow:
      entrypoint: helm
      templates:
        - name: helm
          inputs:
            artifacts:
              - name: input-parameters # The input parameters that holds information what should be executed
                path: "/runner-args"
              - name: runner-context
                path: "/runner-context"
          outputs:
            artifacts:
              - name: helm-release
                path: "/out/helm-release"
          container:
            image: gcr.io/projectvoltron/helm-runner:0.1.0
            env:
              - name: RUNNER_CONTEXT_PATH
                value: "{{inputs.artifacts.runner-context.path}}"
              - name: RUNNER_ARGS_PATH
                value: "{{inputs.artifacts.input-parameters.path}}"
```

As you see, for this definition, as a part of Argo workflow, Kubernetes Engine runs the `gcr.io/projectvoltron/helm-runner:0.1.0` OCI image on Kubernetes and handles dedicated input and output data for this runner. The `helm-values` and `helm-release` arguments need to be described under dedicated Interface for Helm Runner.

## Architecture

The following diagram visualizes how Engine runs an Action using built-in Argo Workflow Runner.

![](./assets/runner-arch.svg)

1. Capact Engine watches the Action custom resources. Once the Action is rendered and a user approved it, Engine executes it.

2. Capact Engine creates a Kubernetes Secret with the [input data](#input-data).

3. Capact Engine creates a Kubernetes Job with the Argo Workflow Runner, and mounts the Secret from the 2nd step as the volume.

4. Argo Workflow Runner reads the input data from the filesystem and based on it creates the Argo Workflow custom resource.

5. Argo Workflow Controller watches Argo Workflows and executes them in a given Namespace. As a result, the actual Action is executed, e.g. Jira installation, cluster benchmarks, etc.

### Input data

Each runner must consume the following environment variables:

| Environment Variable Name | Description                                                                                                          |
|---------------------------|----------------------------------------------------------------------------------------------------------------------|
| **RUNNER_CONTEXT_PATH**   | Specifies the input path for the YAML file that contains the runner context.                                              |
| **RUNNER_ARGS_PATH**      | Specifies the input path for the YAML file that stores rendered data from the Implementation `spec.action.args` property. |

The file with runner context has the following structure: 

```yaml
name: "action-name"          # Specifies Action name. The runner should use this name to correlate the resource it creates.
dryRun: true                 # Specifies whether Action Runner should perform only dry-run action without persisting the resource.
timeout: "10m"               # Specifies the runner timeout when waiting for competition. The zero value means no timeout.
platform:                    # Specifies platform-specific values. Currently, only the Kubernetes platform is supported.
  # Kubernetes platform context properties:
  namespace: "k8s-ns-name"      # Specifies the Kubernetes Namespace where Action is executed. The runner must create all Kubernetes resources in this Namespace.
  serviceAccountName: "sa-name" # Specifies the Kubernetes ServiceAccount. The runner must use it to create all Kubernetes resources.        
  ownerRef: # Specifies owner reference details (Action Custom Resource controller)
    apiVersion: core.capact.io/v1alpha1 # Specifies the owner resource apiVersion
    kind: Action # Specifies the owner resource kind
    blockOwnerDeletion: true # The owner cannot be deleted before the referenced object
    controller: true # Specifies whether the reference points to the managing controller
    name: action-name # Specifies the name of the Action Custom Resource
    uid: 3826a747-cfac-49c7-a81e-1d48cc23096f # Specifies the UID of the Action Custom Resource
```

The runner must read input files from the `RUNNER_CONTEXT_PATH` and `RUNNER_ARGS_PATH` location.

To simplify the development process, we provide Manager, which handles reading the data from disk. All available data is passed for each method execution.

## Development

Read this section to learn how to develop a new runner in Go language.

The [`runner`](./../pkg/runner) package provides Manager, which holds the general logic and allows execution of all runners in the same fashion. This way, each runner implementation holds only business-specific logic.

### Add a runner implementation

Add a new runner under the `pkg/runner/{name}` directory, and implement the [Runner](./../pkg/runner/api.go) interface:

```go
type Runner interface {
	Start(ctx context.Context, in StartInput) (*StartOutput, error)
	WaitForCompletion(ctx context.Context, in WaitForCompletionInput) (*WaitForCompletionOutput, error)
	Name() string
}
```

This allows you to focus on implementing only a business logic related to the new runner.

Optionally you can implement `LoggerInjector` interface. If implemented method is detected, Manager injects [zap](https://github.com/uber-go/zap) logger before executing any other methods on runner.

```go
// LoggerInjector is used by the Manager to inject logger to Runner.
type LoggerInjector interface {
	InjectLogger(*zap.Logger)
}
```

### Create binary

A new runner can be added under the [cmd](../cmd) package.

```go
func main() {
	// Create your runner
	argoRunner := argo.NewRunner()

	// Create status reporter.
	// It is Engine specific implementation that allows Manager
	// report status in a way that Engine knows how to consume it.
	// For non-built-in runners use NOP status reporter.
	statusReporter := statusreporter.NewK8sConfigMap()

	// Create and run Manager.
	mgr, err := runner.NewManager(argoRunner, statusReporter)
	exitOnError(err)

	err = mgr.Execute(stop)
	exitOnError(err)
}
```

Use the following environment variables to configure the Manager:

| Environment Variable Name  | Description                                                                                                                           |
|----------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| **RUNNER_LOGGER_DEV_MODE** | Specifies whether to use the development logger that writes `DebugLevel` and above logs to standard error in a human-friendly format. |

## Available runners

### Argo Workflow Runner

The Argo Workflow Runner implementation is defined in the [pkg/runner/argo](../pkg/runner/argo) package. It creates the Argo Workflow CR and waits for completion using the Kubernetes *watch* functionality. It exits with error when the Argo Workflow with the `context.name` name already exists. Argo Workflow ServiceAccount is always overridden with the one provided via the `context.platform.serviceAccountName` property in the input file.

The implemented dry run functionality only executes the Argo Workflow manifest static validation, and sends a request to the server with the `dry-run` flag, which renders the manifest with the server's representation without creating it.

The Argo Workflow Runner is published to the [gcr.io/projectvoltron/argo-runner](gcr.io/projectvoltron/argo-runner) registry.

> **CAUTION:** As the Argo Workflow does not get created, the nested Action Runners are not executed with the `dry-run` flag.
