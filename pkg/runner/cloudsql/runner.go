package cloudsql

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/yaml"
)

type Args struct {
	Group            string                    `json:"group"`
	Command          string                    `json:"command"`
	GenerateName     bool                      `json:"generateName"`
	Configuration    sqladmin.DatabaseInstance `json:"configuration"`
	AdditionalOutput struct {
		Path  string          `json:"path"`
		Value json.RawMessage `json:"value"`
	} `json:"additionalOutput"`
}

type Output struct {
	DBInstance    *sqladmin.DatabaseInstance
	Port          int    `json:"port"`
	DefaultDBName string `json:"defaultDBName"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

type Runner struct {
	sqladminService *sqladmin.Service
}

func NewRunner(sqladminService *sqladmin.Service) *Runner {
	return &Runner{
		sqladminService: sqladminService,
	}
}

func (r *Runner) Name() string {
	return "CloudSQL"
}

func (r *Runner) Start(ctx context.Context, in runner.StartInput) (*runner.StartOutput, error) {
	args := &Args{}

	project := "projectvoltron"

	if err := json.Unmarshal(in.Args, args); err != nil {
		return nil, errors.Wrap(err, "wrong input args")
	}

	instanceInput := r.getDatabaseInstanceToCreate(&in.ExecCtx, args)
	err := r.createDatabaseInstance(project, instanceInput)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create DB instance")
	}

	db, err := r.waitForDatabaseInstanceRunning(project, instanceInput.Name)
	if err != nil {
		return nil, errors.Wrap(err, "cannot timed out waiting for DB instance to be running")
	}

	output := &Output{
		DBInstance:    db,
		Port:          5432,
		DefaultDBName: "postgres",
		Username:      "postgres",
		Password:      args.Configuration.RootPassword,
	}

	yamlBytes, err := yaml.JSONToYAML(args.AdditionalOutput.Value)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New("output").Parse(string(yamlBytes))
	if err != nil {
		return nil, errors.Wrap(err, "cannot create template from output")
	}

	fd, err := os.Create(args.AdditionalOutput.Path)
	if err != nil {
		return nil, errors.Wrap(err, "cannot open output file for write")
	}
	defer fd.Close()

	err = tmpl.Execute(fd, output)
	if err != nil {
		return nil, err
	}

	return &runner.StartOutput{
		Status: "super",
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

func (r *Runner) createDatabaseInstance(project string, instance *sqladmin.DatabaseInstance) error {
	call := r.sqladminService.Instances.Insert(project, instance)

	// TODO Operation can also have an error, handle it
	_, err := call.Do()
	if err != nil {
		return err
	}
	return nil
}

func (r *Runner) waitForDatabaseInstanceRunning(project string, instanceName string) (*sqladmin.DatabaseInstance, error) {
	// TODO handle timeout, ctx with chans
	for {
		db, err := r.getDatabaseInstance(project, instanceName)
		if err != nil {
			return nil, err
		}

		if db.State == "RUNNABLE" {
			return db, nil
		}

		time.Sleep(time.Second * 10)
	}
}

func (r *Runner) getDatabaseInstance(project, name string) (*sqladmin.DatabaseInstance, error) {
	call := r.sqladminService.Instances.Get(project, name)
	return call.Do()
}
