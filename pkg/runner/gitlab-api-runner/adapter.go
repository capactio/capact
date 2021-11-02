package gitlabapi

import (
	"context"

	"capact.io/capact/pkg/runner"
	"go.uber.org/zap"
)

var _ runner.Runner = &runnerAdapter{}

// TODO: Add support for async and sync runners in Runner Manager

type runnerAdapter struct {
	underlying *RESTRunner
	out        *runner.WaitForCompletionOutput
}

// NewRESTRunner returns new instance of GitLab REST API runner.
func NewRESTRunner(cfg Config) runner.Runner {
	return &runnerAdapter{
		underlying: &RESTRunner{
			cfg: cfg,
		},
	}
}

// Start starts the GitLab REST API runner operation.
func (r *runnerAdapter) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
	var err error
	r.out, err = r.underlying.Do(ctx, in)
	if err != nil {
		return nil, err
	}

	return &runner.StartOutput{
		Status: map[string]interface{}{
			"phase": "Initializing",
		},
	}, nil
}

// WaitForCompletion waits for the GitLab REST API runner operation to complete.
func (r *runnerAdapter) WaitForCompletion(_ context.Context, _ runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	return r.out, nil
}

// Name returns the name of the GitLab REST API runner.
func (r *runnerAdapter) Name() string {
	return r.underlying.Name()
}

// InjectLogger sets the logger on the runner.
func (r *runnerAdapter) InjectLogger(logger *zap.Logger) {
	r.underlying.InjectLogger(logger)
}
