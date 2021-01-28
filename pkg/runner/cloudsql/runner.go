package cloudsql

import (
	"context"

	"projectvoltron.dev/voltron/pkg/runner"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"sigs.k8s.io/yaml"
)

type runnerAction interface {
	Start(ctx context.Context, in *runner.StartInput) (*runner.StartOutput, error)
	WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error)
}

type Runner struct {
	logger          *zap.Logger
	sqladminService *sqladmin.Service
	gcpProjectName  string
	action          runnerAction
	outputCfg       OutputConfig
}

func NewRunner(cfg OutputConfig, sqladminService *sqladmin.Service, gcpProjectName string) *Runner {
	return &Runner{
		outputCfg:       cfg,
		logger:          &zap.Logger{},
		sqladminService: sqladminService,
		gcpProjectName:  gcpProjectName,
	}
}

func (r *Runner) InjectLogger(logger *zap.Logger) {
	r.logger = logger
}

func (r *Runner) Name() string {
	return "cloudsql"
}

func (r *Runner) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
	args := &Args{}

	if err := yaml.Unmarshal(in.Args, args); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling input parameters")
	}

	switch args.Command {
	case CreateCommandType:
		r.action = &createAction{
			logger:          r.logger,
			gcpProjectName:  r.gcpProjectName,
			sqladminService: r.sqladminService,
			args:            args,
			outputCfg:       r.outputCfg,
		}
	default:
		return nil, ErrUnknownCommand
	}

	return r.action.Start(ctx, &in)
}

func (r *Runner) WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	return r.action.WaitForCompletion(ctx, in)
}
