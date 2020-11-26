package main

import (
	"log"

	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"projectvoltron.dev/voltron/internal/graphqlutil"
	"projectvoltron.dev/voltron/internal/healthz"
	"projectvoltron.dev/voltron/internal/och"
)

// Config holds application related configuration
type Config struct {
	//OCHMode represents the possible modes for OCH
	OCHMode och.Mode
	// GraphQLAddr is the TCP address the GraphQL endpoint binds to.
	GraphQLAddr string `envconfig:"default=:8080"`
	// HealthzAddr is the TCP address the health probes endpoint binds to.
	HealthzAddr string `envconfig:"default=:8082"`
	// LoggerDevMode sets the logger to use (or not use) development mode (more human-readable output, extra stack traces
	// and logging information, etc).
	LoggerDevMode bool `envconfig:"default=false"`
	// MockGraphQL sets the grapql servers to use mocked data
	MockGraphQL bool `envconfig:"default=false"`
}

func main() {
	// init config
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	stop := signals.SetupSignalHandler()

	// setup logger
	var logCfg zap.Config
	if cfg.LoggerDevMode {
		logCfg = zap.NewDevelopmentConfig()
	} else {
		logCfg = zap.NewProductionConfig()
	}

	logger, err := logCfg.Build()
	exitOnError(err, "while creating zap logger")

	// healthz server
	hsvr := healthz.NewHTTPServer(logger, cfg.HealthzAddr, cfg.OCHMode.String())

	// GraphQL server
	if cfg.MockGraphQL {
		logger.Info("Using mocked version of OCH API", zap.String("OCH mode", string(cfg.OCHMode)))
	}
	gsvr := graphqlutil.NewHTTPServer(
		logger,
		och.GraphQLSchema(cfg.OCHMode, cfg.MockGraphQL),
		cfg.GraphQLAddr,
		cfg.OCHMode.String(),
	)

	// start servers
	parallelServers := new(errgroup.Group)
	parallelServers.Go(func() error { return hsvr.Start(stop) })
	parallelServers.Go(func() error { return gsvr.Start(stop) })

	err = parallelServers.Wait()
	exitOnError(err, "while waiting for servers to finish gracefully")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
