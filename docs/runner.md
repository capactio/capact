#  Action Runner

Action runners are responsible for running a given action, e.g. Argo Workflow Runner, Helm Runner, etc. The [runner](./../pkg/runner/) package provides Manager which holds the general logic and allows execution of a specific runner in the same fashion. By doing so, each runner implementation holds only business-specific logic.

##  Development

Each Action Runner must consume the following environment variables:

| Environment Variable Name                   | Description                                                                                                                           |
|---------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| **RUNNER_INPUT_MANIFEST_PATH**              | Specifies the input path for the file which holds the rendered manifest defined in the Implementation `spec.action.args` property.        |
| **RUNNER_CONTEXT_NAME**                     | Specifies Action name. The Action Runner should use this name to correlate the resource it creates.                                   |
| **RUNNER_CONTEXT_PLATFORM_{PROPERTY_NAME}** | Specifies platform-specific values. Currently, only the Kubernetes platform is supported.                                             |
| **RUNNER_LOGGER_DEV_MODE**                  | Specifies whether to use the development logger that writes `DebugLevel` and above logs to standard error in a human-friendly format. |
| **RUNNER_TIMEOUT**                          | Specifies the runner timeout when waiting for competition.                                                                            |

Kubernetes platform context properties:

| Environment Variable Name                        | Description                                                                                                                           |
|--------------------------------------------------|---------------------------------------------------------------------------------------------------------------------------------------|
| **RUNNER_CONTEXT_PLATFORM_NAMESPACE**            | Specifies the Kubernetes Namespace where Action is executed. The Action Runners must create all Kubernetes resources in this Namespace. |
| **RUNNER_CONTEXT_PLATFORM_SERVICE_ACCOUNT_NAME** | Specifies the Kubernetes ServiceAccount. The Action Runners must use it to create all Kubernetes resources.                             |

Additionally, Action Runner must read input file from the `RUNNER_INPUT_MANIFEST_PATH` location.

To simplify the development process, we provide Manager, which handles reading the manifest from disk and reading environment variables. All available data is passed for each method execution via `ExecutionContext`.

###  Add a runner implementation

Add a new runner under the `pkg/runner/{name}` directory, and implement the [ActionRunner](./../pkg/runner/api.go) interface:

```go
type ActionRunner interface {
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
