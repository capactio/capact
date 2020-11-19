#  Runner

Runners are responsible for running a given Action, e.g. Argo Workflow Runner, Helm Runner, etc. The [runner](./../pkg/runner/) package provides Manager which holds the general logic and allows execution of a specific runner in the same fashion. By doing so, each runner implementation holds only business-specific logic.

##  Development

Each runner must consume the following environment variables:

| Environment Variable Name | Description                                                                                                                          |
|---------------------------|--------------------------------------------------------------------------------------------------------------------------------------|
| **RUNNER_INPUT_PATH**     | Specifies the input path for the file which holds the context and rendered data from the Implementation `spec.action.args` property. |

Input file syntax:

```yaml
context:
    name: "action-name"          # Specifies Action name. The runner should use this name to correlate the resource it creates.
    timeout: "10m"               # Specifies the runner timeout when waiting for competition.
    platform:                    # Specifies platform-specific values. Currently, only the Kubernetes platform is supported.
      # Kubernetes platform context properties:
      namespace: "k8s-ns-name"      # Specifies the Kubernetes Namespace where Action is executed. The runner must create all Kubernetes resources in this Namespace.
      serviceAccountName: "sa-name" # Specifies the Kubernetes ServiceAccount. The runner must use it to create all Kubernetes resources.        

args:
    # Rendered data data from the Implementation `spec.action.args` property.
```

The runner must read input file from the `RUNNER_INPUT_PATH` location.

To simplify the development process, we provide Manager, which handles reading the data from disk. All available data is passed for each method execution.

###  Add a runner implementation

Add a new runner under the `pkg/runner/{name}` directory, and implement the [Runner](./../pkg/runner/api.go) interface:

```go
type Runner interface {
	Start(ctx context.Context, in StartInput) (*StartOutput, error)
	WaitForCompletion(ctx context.Context, in WaitForCompletionInput) (*WaitForCompletionOutput, error)
	Name() string
}
```

This allows you to focus on implementing only a business logic related to the new runner.

###  Create binary

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

| Environment Variable Name | Description                                                                                                                           |
|---------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| **RUNNER_LOGGER_DEV_MODE**       | Specifies whether to use the development logger that writes `DebugLevel` and above logs to standard error in a human-friendly format. |
