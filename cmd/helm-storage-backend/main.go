package main

import (
	"fmt"
	"log"
	"net"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	helm_storage_backend "capact.io/capact/internal/helm-storage-backend"

	"capact.io/capact/internal/healthz"
	"capact.io/capact/internal/logger"
	"capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Mode describes the selected handler for the Helm storage backend gRPC server.
type Mode string

const (
	// HelmReleaseMode describes the Helm release functionality of the Helm storage backend.
	HelmReleaseMode = "release"
	// HelmTemplateMode describes the Helm templating functionality of the Helm storage backend.
	HelmTemplateMode = "template"
)

// Config holds application related configuration.
type Config struct {
	// GRPCAddr is the TCP address the gRPC server binds to.
	GRPCAddr string `envconfig:"default=:50051"`

	// HealthzAddr is the TCP address the health probes endpoint binds to.
	HealthzAddr string `envconfig:"default=:8082"`

	// Mode describes the selected handler for the Helm storage backend gRPC server.
	Mode Mode

	Logger logger.Config
}

const appName = "helm-storage"

func main() {
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	ctx := signals.SetupSignalHandler()

	// setup logger
	unnamedLogger, err := logger.New(cfg.Logger)
	exitOnError(err, "while creating zap logger")

	logger := unnamedLogger.Named(appName).Named(string(cfg.Mode))

	// k8s

	k8sCfg, err := config.GetConfig()
	exitOnError(err, "while getting K8s config")

	helmCfgFlags := helmCfgFlagsForK8sCfg(k8sCfg)

	// setup servers
	parallelServers := new(errgroup.Group)

	// create handler
	var handler storage_backend.StorageBackendServer
	switch cfg.Mode {
	case HelmReleaseMode:
		handler = helm_storage_backend.NewReleaseHandler(logger, helmCfgFlags)
	case HelmTemplateMode:
		handler = helm_storage_backend.NewTemplateHandler(logger, helmCfgFlags)
	default:
		exitOnError(fmt.Errorf("invalid mode %q", cfg.Mode), "while loading storage backend handler")
	}

	healthzServer := healthz.NewHTTPServer(logger, cfg.HealthzAddr, fmt.Sprintf("%s-%s", appName, cfg.Mode))

	parallelServers.Go(func() error { return healthzServer.Start(ctx) })

	listenCfg := net.ListenConfig{}
	listener, err := listenCfg.Listen(ctx, "tcp", cfg.GRPCAddr)
	exitOnError(err, "while listening")

	srv := grpc.NewServer()
	storage_backend.RegisterStorageBackendServer(srv, handler)

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

func helmCfgFlagsForK8sCfg(k8sCfg *rest.Config) *genericclioptions.ConfigFlags {
	return &genericclioptions.ConfigFlags{
		APIServer:   &k8sCfg.Host,
		Insecure:    &k8sCfg.Insecure,
		CAFile:      &k8sCfg.CAFile,
		BearerToken: &k8sCfg.BearerToken,
	}
}
