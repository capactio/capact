package cloudsql

import (
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"projectvoltron.dev/voltron/pkg/runner"
)

type Args struct {
	Group         string                    `json:"group"`
	Command       CommandType               `json:"command"`
	GenerateName  bool                      `json:"generateName"`
	Configuration sqladmin.DatabaseInstance `json:"configuration"`
	Output        OutputArgs                `json:"output"`
}

type CommandType string

const (
	CreateCommandType = "create"
)

type OutputArgs struct {
	Directory        string `json:"directory"`
	CloudSQLInstance struct {
		Filename string `json:"filename"`
	} `json:"cloudSQLInstance"`
	Additional *struct {
		Path  string          `json:"filename"`
		Value json.RawMessage `json:"value"`
	}
}

type runnerAction interface {
	Start(ctx context.Context, in *runner.StartInput) (*runner.StartOutput, error)
	WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error)
}

type Runner struct {
	logger          *zap.Logger
	sqladminService *sqladmin.Service
	gcpProjectName  string
	action          runnerAction
}

var (
	ErrUnkownCommand = errors.New("unknown command")
)

func NewRunner(sqladminService *sqladmin.Service, gcpProjectName string) *Runner {
	return &Runner{
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

	if err := json.Unmarshal(in.Args, args); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling input parameters")
	}

	switch args.Command {
	case CreateCommandType:
		r.action = &createAction{
			logger:          r.logger,
			gcpProjectName:  r.gcpProjectName,
			sqladminService: r.sqladminService,
			args:            args,
		}
	default:
		return nil, ErrUnkownCommand
	}

	return r.action.Start(ctx, &in)
}

func (r *Runner) WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	return r.action.WaitForCompletion(ctx, in)
}
