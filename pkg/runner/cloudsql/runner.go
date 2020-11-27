package cloudsql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"projectvoltron.dev/voltron/pkg/runner"
)

type Args struct {
	Group         string                    `json:"group"`
	Command       string                    `json:"command"`
	GenerateName  bool                      `json:"generateName"`
	Configuration sqladmin.DatabaseInstance `json:"configuration"`
	Output        OutputArgs                `json:"output"`
}

type OutputArgs struct {
	Directory string `json:"directory"`
	Default   struct {
		Filename string `json:"filename"`
	} `json:"default"`
	Additional struct {
		Path  string          `json:"filename"`
		Value json.RawMessage `json:"value"`
	}
}

type Runner struct {
	logger          *zap.Logger
	sqladminService *sqladmin.Service
	gcpProjectName  string
}

var (
	CreateTimeout time.Duration = time.Minute * 5
)

func NewRunner(logger *zap.Logger, sqladminService *sqladmin.Service, gcpProjectName string) *Runner {
	return &Runner{
		logger:          logger,
		sqladminService: sqladminService,
		gcpProjectName:  gcpProjectName,
	}
}

func (r *Runner) Name() string {
	return "cloudsql"
}

func (r *Runner) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
	args := &Args{}

	if err := json.Unmarshal(in.Args, args); err != nil {
		return nil, errors.Wrap(err, "wrong input args")
	}

	instanceInput := r.getDatabaseInstanceToCreate(&in.ExecCtx, args)

	logger := r.logger.With(zap.String("instanceName", instanceInput.Name))
	logger.Info("creating database")

	err := r.createDatabaseInstance(instanceInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DB instance")
	}

	logger.Info("waiting for database to be running")

	db, err := r.waitForDatabaseInstanceRunning(instanceInput.Name)
	if err != nil {
		return nil, errors.Wrap(err, "timed out waiting for DB instance to be running")
	}

	logger.Info("database ready")

	output := &Output{
		DBInstance:    db,
		Port:          5432,
		DefaultDBName: "postgres",
		Username:      "postgres",
		Password:      args.Configuration.RootPassword,
	}

	if err := writeOutput(&args.Output, output); err != nil {
		return nil, errors.Wrap(err, "failed to write output files")
	}

	return &runner.StartOutput{
		Status: "Completed",
	}, nil
}

func (r *Runner) WaitForCompletion(ctx context.Context, in runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	return &runner.WaitForCompletionOutput{
		Succeeded: true,
	}, nil
}

func (r *Runner) getDatabaseInstanceToCreate(execCtx *runner.ExecutionContext, args *Args) *sqladmin.DatabaseInstance {
	instance := args.Configuration

	if args.GenerateName {
		UUID := uuid.New()
		instance.Name = fmt.Sprintf("%s-%s", execCtx.Name, UUID.String())
	}

	return &instance
}

func (r *Runner) createDatabaseInstance(instance *sqladmin.DatabaseInstance) error {
	call := r.sqladminService.Instances.Insert(r.gcpProjectName, instance)

	_, err := call.Do()
	if err != nil {
		return err
	}

	return nil
}

func (r *Runner) waitForDatabaseInstanceRunning(instanceName string) (*sqladmin.DatabaseInstance, error) {
	logger := r.logger.With(zap.String("instanceName", instanceName))

	ctx, cancel := context.WithTimeout(context.Background(), CreateTimeout)
	defer cancel()

	for {
		select {
		case <-time.After(time.Second * 10):
			logger.Debug("checking db status")
			db, err := r.getDatabaseInstance(instanceName)
			if err != nil {
				return nil, err
			}

			if db.State == "RUNNABLE" {
				return db, err
			}
		case <-ctx.Done():
			logger.Debug("timeout on waiting for db to run")
			return nil, nil
		}
	}
}

func (r *Runner) getDatabaseInstance(name string) (*sqladmin.DatabaseInstance, error) {
	call := r.sqladminService.Instances.Get(r.gcpProjectName, name)
	return call.Do()
}
