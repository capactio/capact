# Development guidelines

In order to contribute to this project, please follow all listed guidelines in this document.

## Architecture

Our architecture leans on [12factor.net](https://www.12factor.net/) principles and is designed to be platform agnostic. 
There should be strong separation between generic Capact business logic and platform-specific implementation, such as Kubernetes, CloudFoundry, OpenShift, and similar.  

## Development tools

We are using Make targets to execute commonly used development tasks. All development tools should be placed in [hack](../hack) folder.

We assume that the Go and Docker is installed both on CI and on a local machine in a proper version. Currently, all scripts don't validate it.
 
If tool is written in a different language, then we use Docker to run that tool, e.g. `quicktype` for Go struct generation from JSON Schemas. 

All tools should:
 - consume `SKIP_DEPS_INSTALLATION` environment variable that can be set to `true` or `false`. If set to `false` dependencies should be installed to a tempdir and cleaned up after execution. 
 - define its stable versions in [./lib/const.sh](../hack/lib/const.sh) file.
 - works both on a local machine and on CI.

> **NOTE:** We are not using dedicated GitHub Actions on CI as we want to have a control and deterministic executions of our tools both on CI and local machines.

## Testing

For assertions in Go unit-tests we are using [testify](https://github.com/stretchr/testify) library as this gives as simple matchers functions which we do not have to write by ourself. 

For integration tests we are using [BDD Ginkgo testing framework](https://github.com/onsi/ginkgo) in combination with [Gomega matcher library](https://github.com/onsi/gomega). 

It is used by default by the Kubebuilder which we used for boostraping Kubernetes controller, and we see value in that if it comes for integration tests.
Thanks to BDD approach the integration tests are more readable and describes better tested contract. Additionally, you can use such assertions out-of-the-box:
- `Consistently()` to ensure that the given state remains same over a duration of time.
- `Eventually()` to repeat a given function every interval seconds until function’s output matches what’s expected.

## Logging

In the Capact project, we use the structured logging concept. 

With structured logging, we associate a *constant* log message with some variable key-value pairs. For instance, suppose we wanted to log that we were starting HTTP server for a given application. In the Go standard library logger, you would write:

```go
log.Printf("starting %s HTTP server on addr %s", appName, svrAddr)
```

With structured logging, you write:

```go
logger.Info("starting HTTP server", "app", appName, "addr", svrAddr)
```

### Kubernetes engine

We use the [`logr`](https://github.com/go-logr/logr) logger in the Kubernetes engine.
The reason is that the library that we use, `contoller-runtime`, uses that under the hood.
For the underlying implementation, we use the [Zap](https://github.com/uber-go/zap) logger.

Read [this](https://github.com/kubernetes-sigs/controller-runtime/blob/master/TMP-LOGGING.md#logging-guidelines) document if you want to learn more about logging practices in the Kubernetes engine domain.

### Others applications

In other applications, we use [Zap](https://github.com/uber-go/zap) logger directly without the [`logr`](https://github.com/go-logr/logr) abstraction. 

While it's possible to use other log levels, it's recommended that you stick with `Info` and `Error` levels. 
Read [this](https://dave.cheney.net/2015/11/05/lets-talk-about-logging) article to understand why other levels are obsolete and lead to inconsistent logs.
