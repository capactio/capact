package helm

import (
	"context"

	"capact.io/capact/pkg/runner"
	"go.uber.org/zap"
	"k8s.io/client-go/rest"
)

var _ runner.Runner = &runnerAdapter{}

// TODO: Remove adapter once Runner interface changes

type runnerAdapter struct {
	underlying *helmRunner
	out        *runner.WaitForCompletionOutput
}

// NewRunner returns new instance of Helm runner.
func NewRunner(k8sCfg *rest.Config, cfg Config) runner.Runner {
	return &runnerAdapter{
		underlying: newHelmRunner(k8sCfg, cfg),
	}
}

// Start the Helm runner operation.
func (r *runnerAdapter) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
	var err error
	r.out, err = r.underlying.Do(ctx, in)
	if err != nil {
		return nil, err
	}

	return &runner.StartOutput{
		Status: map[string]interface{}{
			"phase": "Installing",
		},
	}, nil
}

// WaitForCompletion waits for the Helm runner operation to complete.
func (r *runnerAdapter) WaitForCompletion(_ context.Context, _ runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	return r.out, nil
}

// Name returns the name of the Helm runner.
func (r *runnerAdapter) Name() string {
	return r.underlying.Name()
}

// InjectLogger sets the logger on the runner.
func (r *runnerAdapter) InjectLogger(logger *zap.Logger) {
	r.underlying.InjectLogger(logger)
}
