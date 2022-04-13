package main

import (
	"fmt"
	"log"

	"google.golang.org/grpc"

	"capact.io/capact/internal/logger"
	tivaluefetcher "capact.io/capact/internal/ti-value-fetcher"
	"github.com/vrischmann/envconfig"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Config holds TypeInstance Value resolver configuration.
type Config struct {
	Input struct {
		TIFilePath        string `envconfig:"default=/tmp/input-ti.yaml"`
		BackendTIFilePath string `envconfig:"default=/tmp/storage-backend.yaml"`
	}
	OutputFilePath string `envconfig:"default=/tmp/output.yaml"`

	Logger logger.Config
}

const appName = "ti-value-fetcher"

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	ctx := signals.SetupSignalHandler()

	// setup logger
	unnamedLogger, err := logger.New(cfg.Logger)
	exitOnError(err, "while creating zap logger")

	logger := unnamedLogger.Named(appName)

	tiValueFetcher := tivaluefetcher.New(logger)

	tiArtifact, storageBackendValue, err := tiValueFetcher.LoadFromFile(cfg.Input.TIFilePath, cfg.Input.BackendTIFilePath)
	exitOnError(err, "while loading input files")

	res, err := tiValueFetcher.Do(ctx, tiArtifact, storageBackendValue, grpc.WithInsecure())
	exitOnError(err, "while resolving TI value")

	err = tiValueFetcher.SaveToFile(cfg.OutputFilePath, res)
	exitOnError(err, fmt.Sprintf("while saving output to file %q", cfg.OutputFilePath))
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
