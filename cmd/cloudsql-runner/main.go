package main

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"projectvoltron.dev/voltron/pkg/runner"
	"projectvoltron.dev/voltron/pkg/runner/cloudsql"
	statusreporter "projectvoltron.dev/voltron/pkg/runner/status-reporter"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type Config struct {
	GCPServiceAccountFilepath string `envconfig:"default=/etc/gcp/sa.json"`
}

var scopes = []string{
	"https://www.googleapis.com/auth/cloud-platform",
	"https://www.googleapis.com/auth/sqlservice.admin",
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "RUNNER")
	exitOnError(err, "failed to load config")

	credsBytes, err := loadGCPCredentialsFileBytes(&cfg)
	exitOnError(err, "failed to get GCP service account")

	gcpCreds, err := google.CredentialsFromJSON(context.Background(), credsBytes, scopes...)
	exitOnError(err, "failed to read GCP credentials")

	service, err := sqladmin.NewService(context.Background(), option.WithCredentials(gcpCreds))
	exitOnError(err, "failed to create GCP service client")

	cloudsqlRunner := cloudsql.NewRunner(service, gcpCreds.ProjectID)

	statusReporter := statusreporter.NewNoop()

	mgr, err := runner.NewManager(cloudsqlRunner, statusReporter)
	exitOnError(err, "failed to create manager")

	stop := signals.SetupSignalHandler()

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")
}

func loadGCPCredentialsFileBytes(cfg *Config) ([]byte, error) {
	rawInput, err := ioutil.ReadFile(cfg.GCPServiceAccountFilepath)
	if err != nil {
		return rawInput, errors.Wrap(err, "while reading GCP credentials file")
	}

	return rawInput, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
