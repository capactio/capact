package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"

	"sigs.k8s.io/yaml"

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

type serviceAccount struct {
	Key         json.RawMessage `json:"key"`
	ProjectName string          `json:"projectName"`
}
type Config struct {
	GCPServiceAccountFilepath string `envconfig:"default=/etc/gcp/sa.yaml"`
}

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "failed to load config")

	sa, err := getGCPServiceAccount(&cfg)
	exitOnError(err, "failed to get GCP service account")

	creds, err := google.CredentialsFromJSON(context.Background(), sa.Key)
	exitOnError(err, "failed to read GCP credentials")

	service, err := sqladmin.NewService(context.Background(), option.WithCredentials(creds))
	exitOnError(err, "failed to create GCP service client")

	fmt.Println(sa.ProjectName)

	cloudsqlRunner := cloudsql.NewRunner(service, sa.ProjectName)

	statusReporter := statusreporter.NewNoop()

	mgr, err := runner.NewManager(cloudsqlRunner, statusReporter)
	exitOnError(err, "failed to create manager")

	stop := signals.SetupSignalHandler()

	err = mgr.Execute(stop)
	exitOnError(err, "while executing runner")
}

func getGCPServiceAccount(cfg *Config) (serviceAccount, error) {
	sa := serviceAccount{}

	rawInput, err := ioutil.ReadFile(cfg.GCPServiceAccountFilepath)
	if err != nil {
		return sa, errors.Wrap(err, "while reading GCP service account file")
	}

	err = yaml.Unmarshal(rawInput, &sa)
	if err != nil {
		return sa, errors.Wrap(err, "while unmarshaling GCP service account file")
	}

	return sa, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
