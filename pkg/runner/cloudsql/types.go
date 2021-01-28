package cloudsql

import (
	"encoding/json"
	"errors"
	"os"
	"time"

	sqladmin "google.golang.org/api/sqladmin/v1beta4"
)

const (
	PostgresPort          = 5432
	PostgresDefaultDBName = "postgres"
	PostgresRootUser      = "postgres"

	createWaitDelay = 10 * time.Second

	artifactsFileMode os.FileMode = 0644
)

type CommandType string

const (
	CreateCommandType = "create"
)

var (
	ErrInstanceCreateTimeout = errors.New("timed out waiting for DB instance to be ready")
	ErrUnknownCommand        = errors.New("unknown command")
)

type OutputConfig struct {
	CloudSQLInstanceFilePath string `envconfig:"default=/tmp/cloudSQLInstance.yaml"`
	AdditionalFilePath       string `envconfig:"default=/tmp/additional.yaml"`
}

type Args struct {
	Group        string                    `yaml:"group"`
	Command      CommandType               `yaml:"command"`
	GenerateName bool                      `yaml:"generateName"`
	Instance     sqladmin.DatabaseInstance `yaml:"instance"`
	Output       OutputArgs                `yaml:"output"`
}

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
