package main

import (
	wfclientset "github.com/argoproj/argo/pkg/client/clientset/versioned"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"log"
	"projectvoltron.dev/voltron/pkg/runner"
	"projectvoltron.dev/voltron/pkg/runner/argo"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type Config struct {
	Runner runner.Config
	// LoggerDevMode sets the logger to use (or not use) development mode (more human-readable output, extra stack traces
	// and logging information, etc).
	LoggerDevMode bool `envconfig:"default=false"`
}

// TODO: go with template method pattern or with composable blocks?
func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	// setup logger
	var logCfg zap.Config
	if cfg.LoggerDevMode {
		logCfg = zap.NewDevelopmentConfig()
	} else {
		logCfg = zap.NewProductionConfig()
	}

	log, err := logCfg.Build()
	exitOnError(err, "while creating zap logger")

	k8sCfg, err := config.GetConfig()
	exitOnError(err, "while creating k8s config")


	// create the workflow client
	wfClient := wfclientset.NewForConfigOrDie(k8sCfg).ArgoprojV1alpha1()

	argo.NewRunner(log, wfClient)
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
