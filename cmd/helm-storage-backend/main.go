package main

import (
	"fmt"
	"log"
	"net"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/rest"

	helm_storage_backend "capact.io/capact/internal/helm-storage-backend"

	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"capact.io/capact/internal/healthz"
	"capact.io/capact/internal/logger"
	"capact.io/capact/pkg/hub/api/grpc/storage_backend"
	"capact.io/capact/pkg/hub/client/local"
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
	// LocalHubEndpoint is an endpoint to the Local Hub.
	LocalHubEndpoint string `envconfig:"default=http://capact-hub-local.capact-system/graphql"`

	// KubeconfigTypeinstanceID is the optional Kubeconfig TypeInstance.
	KubeconfigTypeinstanceID string `envconfig:"optional"`

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

	if cfg.KubeconfigTypeinstanceID != "" {
		hubClient := local.NewDefaultClient(cfg.LocalHubEndpoint)
		kubeconfigFetcher := helm_storage_backend.NewKubeconfigFetcher(hubClient)
		err := kubeconfigFetcher.SetKubeconfigBasedOnTypeInstance(ctx, logger, cfg.KubeconfigTypeinstanceID)
		if err != nil {
			exitOnError(err, fmt.Sprintf("while setting kubeconfig based on TypeInstance ID: %s", cfg.KubeconfigTypeinstanceID))
		}
	}

	k8sCfg, err := config.GetConfig()
	exitOnError(err, "while getting K8s config")

	helmCfgFlags := helmCfgFlagsForK8sCfg(k8sCfg)

	relFetcher := helm_storage_backend.NewHelmReleaseFetcher(helmCfgFlags)

	// setup servers
	parallelServers := new(errgroup.Group)

	// create handler
	var handler storage_backend.ContextStorageBackendServer
	switch cfg.Mode {
	case HelmReleaseMode:
		handler, err = helm_storage_backend.NewReleaseHandler(logger, relFetcher)
		exitOnError(err, "while creating Helm Release backend storage")
	case HelmTemplateMode:
		handler = helm_storage_backend.NewTemplateHandler(logger, relFetcher)
	default:
		exitOnError(fmt.Errorf("invalid mode %q", cfg.Mode), "while loading storage backend handler")
	}

	healthzServer := healthz.NewHTTPServer(logger, cfg.HealthzAddr, fmt.Sprintf("%s-%s", appName, cfg.Mode))

	parallelServers.Go(func() error { return healthzServer.Start(ctx) })

	listenCfg := net.ListenConfig{}
	listener, err := listenCfg.Listen(ctx, "tcp", cfg.GRPCAddr)
	exitOnError(err, "while listening")

	srv := grpc.NewServer()
	storage_backend.RegisterContextStorageBackendServer(srv, handler)

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
