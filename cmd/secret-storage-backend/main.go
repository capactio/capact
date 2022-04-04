package main

import (
	"log"
	"net"

	"capact.io/capact/internal/healthz"
	"capact.io/capact/internal/logger"
	secret_storage_backend "capact.io/capact/internal/secret-storage-backend"
	"capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"github.com/pkg/errors"
	tellerpkg "github.com/spectralops/teller/pkg"
	tellercore "github.com/spectralops/teller/pkg/core"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Config holds application related configuration.
type Config struct {
	// GRPCAddr is the TCP address the gRPC server binds to.
	GRPCAddr string `envconfig:"default=:50051"`

	// HealthzAddr is the TCP address the health probes endpoint binds to.
	HealthzAddr string `envconfig:"default=:8082"`

	// SupportedProviders holds enabled secret providers separated by comma.
	SupportedProviders []string `envconfig:"default=aws_secretsmanager"`

	Logger logger.Config
}

const appName = "secret-storage-backend"

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	ctx := signals.SetupSignalHandler()

	// setup logger
	unnamedLogger, err := logger.New(cfg.Logger)
	exitOnError(err, "while creating zap logger")

	logger := unnamedLogger.Named(appName)

	// setup servers
	parallelServers := new(errgroup.Group)

	healthzServer := healthz.NewHTTPServer(logger, cfg.HealthzAddr, appName)
	parallelServers.Go(func() error { return healthzServer.Start(ctx) })

	logger.Info("loaded secret providers", zap.Strings("providers", cfg.SupportedProviders))
	providers, err := loadProviders(cfg.SupportedProviders)
	exitOnError(err, "while loading providers")

	handler := secret_storage_backend.NewHandler(logger, providers)

	listenCfg := net.ListenConfig{}
	listener, err := listenCfg.Listen(ctx, "tcp", cfg.GRPCAddr)
	exitOnError(err, "while listening")

	srv := grpc.NewServer()
	storage_backend.RegisterValueAndContextStorageBackendServer(srv, handler)

	go func() {
		<-ctx.Done()
		logger.Info("Stopping server gracefully")
		srv.GracefulStop()
	}()

	parallelServers.Go(func() error {
		logger.Info("Starting TCP server", zap.String("addr", cfg.GRPCAddr))
		return srv.Serve(listener)
	})

	err = parallelServers.Wait()
	exitOnError(err, "while waiting for servers to finish gracefully")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}

func loadProviders(providerNames []string) (map[string]tellercore.Provider, error) {
	builtInProviders := tellerpkg.BuiltinProviders{}
	providersMap := map[string]tellercore.Provider{}

	for _, providerName := range providerNames {
		provider, err := builtInProviders.GetProvider(providerName)
		if err != nil {
			return nil, errors.Wrapf(err, "while loading provider %q", provider)
		}

		providersMap[providerName] = provider
	}

	if len(providersMap) == 0 {
		return nil, errors.New("at least one secret provider has to be configured")
	}

	return providersMap, nil
}
