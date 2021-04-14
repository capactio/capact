package main

import (
	"context"
	"log"

	"capact.io/capact/pkg/runner"
	"capact.io/capact/pkg/runner/cloudsql"
	statusreporter "capact.io/capact/pkg/runner/status-reporter"

	"github.com/vrischmann/envconfig"
	"google.golang.org/api/option"
	sqladmin "google.golang.org/api/sqladmin/v1beta4"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

type Config struct {
	GCP    cloudsql.GCPConfig
	Output cloudsql.OutputConfig
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "RUNNER")
	exitOnError(err, "failed to load config")

	gcpCreds, err := cloudsql.LoadGCPCredentials(cfg.GCP)
	exitOnError(err, "failed to load GCP credentials")

	service, err := sqladmin.NewService(context.Background(), option.WithCredentials(gcpCreds))
	exitOnError(err, "failed to create GCP service client")

	cloudsqlRunner := cloudsql.NewRunner(cfg.Output, service, gcpCreds.ProjectID)

	statusReporter := statusreporter.NewNoop()

	mgr, err := runner.NewManager(cloudsqlRunner, statusReporter)
	exitOnError(err, "failed to create manager")

	stop := signals.SetupSignalHandler()

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
