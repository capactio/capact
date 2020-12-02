package cloudsql

import (
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"time"

	"github.com/avast/retry-go"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sethvargo/go-password/password"
	"go.uber.org/zap"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"projectvoltron.dev/voltron/pkg/runner"
	"sigs.k8s.io/yaml"
)

type createAction struct {
	logger          *zap.Logger
	sqladminService *sqladmin.Service
	gcpProjectName  string
	args            *Args
	dbInstance      *sqladmin.DatabaseInstance
}

type createOutputValues struct {
	DBInstance    *sqladmin.DatabaseInstance
	Port          int    `json:"port"`
	DefaultDBName string `json:"defaultDBName"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

type cloudSQLOutput struct {
	Name            string `json:"name"`
	Project         string `json:"project"`
	Region          string `json:"region"`
	DatabaseVersion string `json:"databaseVersion"`
}

const (
	PostgresPort          = 5432
	PostgresDefaultDBName = "postgres"
	PostgresRootUser      = "postgres"

	createWaitDelay    = 10 * time.Second
	createWaitAttempts = 60

	artifactsDirFileMode os.FileMode = 0775
	artifactsFileMode    os.FileMode = 0644
)

var (
	ErrInstanceNotReady = errors.New("DB instance not ready")
)

func (a *createAction) Start(_ context.Context, in *runner.StartInput) (*runner.StartOutput, error) {
	var err error

	a.dbInstance, err = a.prepareCreateDatabaseInstanceParameters(&in.ExecCtx, a.args)
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

func (a *createAction) WaitForCompletion(_ context.Context, _ runner.WaitForCompletionInput) (*runner.WaitForCompletionOutput, error) {
	a.logger.Info("waiting for database to be running")

	db, err := a.waitForDatabaseInstanceRunning(a.dbInstance.Name)
	if err != nil {
		return nil, errors.Wrap(err, "while waiting for database to be ready")
	}

	a.dbInstance = db

	a.logger.Info("database ready")

	output := &createOutputValues{
		DBInstance:    a.dbInstance,
		Port:          PostgresPort,
		DefaultDBName: PostgresDefaultDBName,
		Username:      PostgresRootUser,
		Password:      a.dbInstance.RootPassword,
	}

	if err := a.createOutputFiles(&a.args.Output, output); err != nil {
		return nil, errors.Wrap(err, "while writing output")
	}

	return &runner.WaitForCompletionOutput{
		Succeeded: true,
		Message:   fmt.Sprintf("Create database %s", a.dbInstance.Name),
	}, nil
}

func (a *createAction) prepareCreateDatabaseInstanceParameters(execCtx *runner.ExecutionContext, args *Args) (*sqladmin.DatabaseInstance, error) {
	instance := args.Instance
	instance.Project = a.gcpProjectName

	if args.GenerateName {
		UUID := uuid.New()
		instance.Name = fmt.Sprintf("%s-%s", execCtx.Name, UUID.String())
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

func (a *createAction) waitForDatabaseInstanceRunning(instanceName string) (*sqladmin.DatabaseInstance, error) {
	logger := a.logger.With(zap.String("instanceName", instanceName))

	var db *sqladmin.DatabaseInstance

	err := retry.Do(
		func() error {
			var err error

			logger.Debug("checking db status")
			db, err = a.getDatabaseInstance(instanceName)
			if err != nil {
				return err
			}

			if db.State == "RUNNABLE" {
				return nil
			}

			return ErrInstanceNotReady
		},
		retry.Delay(createWaitDelay),
		retry.Attempts(createWaitAttempts),
	)

	if err != nil {
		return nil, errors.Wrap(err, "while waiting for DB instance to be ready")
	}

	return db, err
}

func (a *createAction) getDatabaseInstance(name string) (*sqladmin.DatabaseInstance, error) {
	return a.sqladminService.Instances.Get(a.gcpProjectName, name).Do()
}

func (a *createAction) createOutputFiles(args *OutputArgs, values *createOutputValues) error {
	if err := os.MkdirAll(args.Directory, artifactsDirFileMode); err != nil {
		return err
	}

	if err := a.createCloudSQLInstanceOutputFile(args, values); err != nil {
		return errors.Wrap(err, "while creating default artifact")
	}

	if args.Additional != nil {
		if err := a.createAdditionalOutputFile(args, values); err != nil {
			return errors.Wrap(err, "while creating additional artifact")
		}
	}

	return nil
}

func (a *createAction) createCloudSQLInstanceOutputFile(args *OutputArgs, output *createOutputValues) error {
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

	artifactFilepath := fmt.Sprintf("%s/%s", args.Directory, args.CloudSQLInstance.Filename)

	if err := ioutil.WriteFile(artifactFilepath, data, artifactsFileMode); err != nil {
		return errors.Wrapf(err, "while writing artifact file %s", artifactFilepath)
	}

	return nil
}

func (a *createAction) createAdditionalOutputFile(args *OutputArgs, values *createOutputValues) error {
	artifactTemplate, err := yaml.JSONToYAML(args.Additional.Value)
	if err != nil {
		return errors.Wrap(err, "while converting JSON to YAML")
	}

	tmpl, err := template.New("output").Parse(string(artifactTemplate))
	if err != nil {
		return errors.Wrap(err, "failed to load template")
	}

	filepath := fmt.Sprintf("%s/%s", args.Directory, args.Additional.Path)

	fd, err := os.Create(filepath)
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
