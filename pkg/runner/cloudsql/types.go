package cloudsql

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

const (
	// PostgresPort defines the CloudSQL DB port.
	PostgresPort = 5432
	// PostgresDefaultDBName defines the CloudSQL default database name.
	PostgresDefaultDBName = "postgres"
	// PostgresRootUser defines the CloudSQL database instance root user name.
	PostgresRootUser = "postgres"

	createWaitDelay = 10 * time.Second

	artifactsFileMode os.FileMode = 0644
)

// CommandType represents the operation type to be performed by the runner.
type CommandType string

const (
	// CreateCommandType is an operation to create a new CloudSQL instance.
	CreateCommandType = "create"
)

var (
	// ErrInstanceCreateTimeout indicates the operation timed out.
	ErrInstanceCreateTimeout = errors.New("timed out waiting for DB instance to be ready")
	// ErrUnknownCommand indicates an unknown operation command.
	ErrUnknownCommand = errors.New("unknown command")
)

// OutputConfig stores the configuration for the CloudSQL runner output files.
type OutputConfig struct {
	CloudSQLInstanceFilePath string `envconfig:"default=/tmp/cloudSQLInstance.yaml"`
	AdditionalFilePath       string `envconfig:"default=/tmp/additional.yaml"`
}

// Args stores the input arguments for the CloudSQL runner operation.
type Args struct {
	Group        string                    `yaml:"group"`
	Command      CommandType               `yaml:"command"`
	GenerateName bool                      `yaml:"generateName"`
	Instance     sqladmin.DatabaseInstance `yaml:"instance"`
	Output       OutputArgs                `yaml:"output"`
}

// OutputArgs stores the arguments for the output of the CloudSQL runner.
type OutputArgs struct {
	GoTemplate json.RawMessage `yaml:"goTemplate"`
}

type createOutputValues struct {
	DBInstance    *sqladmin.DatabaseInstance
	Port          int    `yaml:"port"`
	DefaultDBName string `yaml:"defaultDBName"`
	Username      string `yaml:"username"`
	Password      string `yaml:"password"`
}

type cloudSQLOutput struct {
	Name            string `yaml:"name"`
	Project         string `yaml:"project"`
	Region          string `yaml:"region"`
	DatabaseVersion string `yaml:"databaseVersion"`
}
