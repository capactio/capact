package helm

import (
	"context"

	"k8s.io/client-go/rest"
	"projectvoltron.dev/voltron/pkg/runner"
)

var _ runner.Runner = &runnerAdapter{}

// TODO: Remove adapter once Runner interface changes

type runnerAdapter struct {
	r   *helmRunner
	out *runner.WaitForCompletionOutput
}

func NewRunner(k8sCfg *rest.Config, cfg Config) runner.Runner {
	return &runnerAdapter{
		r: newHelmRunner(k8sCfg, cfg),
	}
}

func (r *runnerAdapter) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
	var err error
	r.out, err = r.r.Do(ctx, in)
	if err != nil {
		return nil, err
	}

	return &runner.StartOutput{
		Status: map[string]interface{}{
			"phase": "Installing",
		},
	}, nil
}

func (r *runnerAdapter) WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	return r.out, nil
}

func (r *runnerAdapter) Name() string {
	return r.r.Name()
}
