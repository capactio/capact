package runner

import "context"

type (
	StartInput struct {
		ExecCtx  ExecutionContext
		Manifest []byte
	}

	StartOutput struct {
		Status interface{}
	}

	WaitForCompletionInput struct {
		ExecCtx ExecutionContext
	}
)

// ActionRunner provide functionality to execute runner in a generic way
type ActionRunner interface {
	Start(ctx context.Context, in StartInput) (*StartOutput, error)
	WaitForCompletion(ctx context.Context, in WaitForCompletionInput) error
	Name() string
}

// StatusReporter provide functionality to report status.
type StatusReporter interface {
	Report(ctx context.Context, execCtx ExecutionContext, status interface{}) error
}
