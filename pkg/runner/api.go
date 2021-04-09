package runner

import (
	"context"
	"encoding/json"
	"fmt"
)

// API types used to ensure extendable function input/output.
type (
	StartInput struct {
		// RunnerCtx contains Runner data provided by Engine.
		RunnerCtx Context
		// Args that was provided by Engine.
		Args json.RawMessage
	}

	StartOutput struct {
		// Status holds generic status object that is later marshalled to JSON format.
		Status interface{}
	}

	WaitForCompletionInput struct {
		// RunnerCtx contains Runner data provided by Engine.
		RunnerCtx Context
	}

	WaitForCompletionOutput struct {
		// Succeeded indicates if runner finished successfully or not.
		Succeeded bool
		// Message holds a human readable message indicating details about why the is in this condition.
		Message string
	}
)

// ErrorOrNil returns error if action finished unsuccessfully.
func (o WaitForCompletionOutput) ErrorOrNil() error {
	if !o.Succeeded {
		return fmt.Errorf("finished unsuccessfully [details: %q]", o.Message)
	}
	return nil
}

// Runner provide functionality to execute runner in a generic way.
type Runner interface {
	Start(ctx context.Context, in StartInput) (*StartOutput, error)
	WaitForCompletion(ctx context.Context, in WaitForCompletionInput) (*WaitForCompletionOutput, error)
	Name() string
}

// StatusReporter provide functionality to report status.
type StatusReporter interface {
	Report(ctx context.Context, runnerCtx Context, status interface{}) error
}
