package cloudsql

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"capact.io/capact/pkg/runner"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	"go.uber.org/zap"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"sigs.k8s.io/yaml"
)

type createAction struct {
	logger          *zap.Logger
	sqladminService *sqladmin.Service
	gcpProjectName  string
	args            *Args
	dbInstance      *sqladmin.DatabaseInstance
	outputCfg       OutputConfig
}

func (a *createAction) Start(_ context.Context, in *runner.StartInput) (*runner.StartOutput, error) {
	var err error

	a.dbInstance, err = a.prepareCreateDatabaseInstanceParameters(&in.RunnerCtx, a.args)
	if err != nil {
		return nil, errors.Wrap(err, "while preparing create database instance parameters")
	}

	a.logger = a.logger.With(zap.String("instanceName", a.dbInstance.Name))
	a.logger.Info("creating database")

	err = a.createDatabaseInstance(a.dbInstance)
	if err != nil {
		return nil, errors.Wrap(err, "while creating database instance")
	}

	return &runner.StartOutput{
		Status: "Creating database instance",
	}, nil
}

func (a *createAction) WaitForCompletion(ctx context.Context, _ runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	a.logger.Info("waiting for database to be running")

	createdDb, err := a.waitForDatabaseInstanceRunning(ctx, a.dbInstance.Name)
	if err != nil {
		return nil, errors.Wrap(err, "while waiting for database to be ready")
	}

	a.logger.Info("database ready")

	output := &createOutputValues{
		DBInstance:    createdDb,
		Port:          PostgresPort,
		DefaultDBName: PostgresDefaultDBName,
		Username:      PostgresRootUser,
		Password:      a.dbInstance.RootPassword,
	}

	if err := a.createOutputFiles(a.outputCfg, &a.args.Output, output); err != nil {
		return nil, errors.Wrap(err, "while writing output")
	}

	return &runner.WaitForCompletionOutput{
		Succeeded: true,
		Message:   fmt.Sprintf("Create database %s", a.dbInstance.Name),
	}, nil
}

func (a *createAction) prepareCreateDatabaseInstanceParameters(runnerCtx *runner.Context, args *Args) (*sqladmin.DatabaseInstance, error) {
	instance := args.Instance
	instance.Project = a.gcpProjectName

	if args.GenerateName {
		UUID := uuid.New()
		instance.Name = fmt.Sprintf("%s-%s", runnerCtx.Name, UUID.String())
	}

	if instance.RootPassword == "" {
		passwd, err := password.Generate(16, 4, 4, false, false)
		if err != nil {
			return nil, errors.Wrap(err, "while generating random root password")
		}

		instance.RootPassword = passwd
	}

	return &instance, nil
}

func (a *createAction) createDatabaseInstance(instance *sqladmin.DatabaseInstance) error {
	_, err := a.sqladminService.Instances.Insert(a.gcpProjectName, instance).Do()
	return err
}

func (a *createAction) waitForDatabaseInstanceRunning(ctx context.Context, instanceName string) (*sqladmin.DatabaseInstance, error) {
	for {
		select {
		case <-time.After(createWaitDelay):
			a.logger.Debug("checking db instance status")
			db, err := a.getDatabaseInstance(instanceName)
			if err != nil {
				return nil, errors.Wrap(err, "while getting DB instance")
			}

			if db.State == "RUNNABLE" {
				return db, nil
			}
		case <-ctx.Done():
			return nil, ErrInstanceCreateTimeout
		}
	}
}

func (a *createAction) getDatabaseInstance(name string) (*sqladmin.DatabaseInstance, error) {
	return a.sqladminService.Instances.Get(a.gcpProjectName, name).Do()
}

func (a *createAction) createOutputFiles(cfg OutputConfig, args *OutputArgs, values *createOutputValues) error {
	if err := a.createCloudSQLInstanceOutputFile(cfg.CloudSQLInstanceFilePath, values); err != nil {
		return errors.Wrap(err, "while creating default artifact")
	}

	if args.GoTemplate == nil {
		return nil
	}

	if err := a.createAdditionalOutputFile(cfg.AdditionalFilePath, args, values); err != nil {
		return errors.Wrap(err, "while creating additional artifact")
	}

	return nil
}

func (a *createAction) createCloudSQLInstanceOutputFile(path string, output *createOutputValues) error {
	artifact := &cloudSQLOutput{
		Name:            output.DBInstance.Name,
		Project:         output.DBInstance.Project,
		Region:          output.DBInstance.Region,
		DatabaseVersion: output.DBInstance.DatabaseVersion,
	}

	data, err := yaml.Marshal(artifact)
	if err != nil {
		return errors.Wrap(err, "while marshaling artifact to YAML")
	}

	if err := ioutil.WriteFile(path, data, artifactsFileMode); err != nil {
		return errors.Wrapf(err, "while writing artifact file %s", path)
	}

	return nil
}

func (a *createAction) createAdditionalOutputFile(path string, args *OutputArgs, values *createOutputValues) error {
	// yaml.Unmarshal converts YAML to JSON then uses JSON to unmarshal into an object
	// but the GoTemplate is defined via YAML, so we need to revert that change
	artifactTemplate, err := yaml.JSONToYAML(args.GoTemplate)
	if err != nil {
		return errors.Wrap(err, "while converting GoTemplate property from JSON to YAML")
	}

	tmpl, err := template.New("output").Parse(string(artifactTemplate))
	if err != nil {
		return errors.Wrap(err, "failed to load template")
	}

	fd, err := os.Create(filepath.Clean(path))
	if err != nil {
		return errors.Wrap(err, "cannot open output file to write")
	}
	defer func() {
		if err := fd.Close(); err != nil {
			a.logger.Error("failed to close output file descriptor", zap.Error(err))
		}
	}()

	err = tmpl.Execute(fd, values)
	if err != nil {
		return errors.Wrap(err, "failed to render output file")
	}

	return nil
}
