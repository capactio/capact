package runner

import (
	"context"
	"fmt"
)

// API types used to ensure extendable function input/output.
type (
	StartInput struct {
		// ExecCtx holds all information provided by Engine.
		ExecCtx ExecutionContext
		// Manifest that was provided by Engine.
		Manifest []byte
	}

	StartOutput struct {
		// Status holds generic status object that is later marshaled to JSON format.
		Status interface{}
	}

	WaitForCompletionInput struct {
		// ExecCtx holds all information provided by Engine.
		ExecCtx ExecutionContext
	}

	WaitForCompletionOutput struct {
		// FinishedSuccessfully indicates if runner finished successfully or not.
		FinishedSuccessfully bool
		// Message holds a human readable message indicating details about why the is in this condition.
		Message string
	}
)

// ErrorOrNil returns error if action finished unsuccessfully.
func (o WaitForCompletionOutput) ErrorOrNil() error {
	if !o.FinishedSuccessfully {
		return fmt.Errorf("finished unsuccessfully [details: %q]", o.Message)
	}
	return nil
}

// ActionRunner provide functionality to execute runner in a generic way.
type ActionRunner interface {
	Start(ctx context.Context, in StartInput) (*StartOutput, error)
	WaitForCompletion(ctx context.Context, in WaitForCompletionInput) (*WaitForCompletionOutput, error)
	Name() string
}

// StatusReporter provide functionality to report status.
type StatusReporter interface {
	Report(ctx context.Context, execCtx ExecutionContext, status interface{}) error
}
