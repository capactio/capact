package main

import (
	"log"
	"time"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
	"projectvoltron.dev/voltron/pkg/httputil"

	gqlgen_graphql "github.com/99designs/gqlgen/graphql"
	"github.com/go-logr/zapr"
	"github.com/vrischmann/envconfig"
	uber_zap "go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"projectvoltron.dev/voltron/internal/graphqlutil"
	"projectvoltron.dev/voltron/internal/k8s-engine/controller"
	domaingraphql "projectvoltron.dev/voltron/internal/k8s-engine/graphql"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	ochclient "projectvoltron.dev/voltron/pkg/och/client"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	GraphQLServerName = "engine-graphql"
)

// Config holds application related configuration
type Config struct {
	// LeaderElectionNamespace determines the namespace in which the leader election configmap will be created.
	LeaderElectionNamespace string `envconfig:"optional"`
	// GraphQLAddr is the TCP address the GraphQL endpoint binds to.
	GraphQLAddr string `envconfig:"default=:8080"`
	// MetricsAddr is the TCP address the metric endpoint binds to.
	MetricsAddr string `envconfig:"default=:8081"`
	// HealthzAddr is the TCP address the health probes endpoint binds to.
	HealthzAddr string `envconfig:"default=:8082"`
	// EnableLeaderElection for controller manager. Enabling this will ensure there is only one active controller manager.
	EnableLeaderElection bool `envconfig:"default=false"`
	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles which can be run. Defaults to 1.
	MaxConcurrentReconciles int `envconfig:"default=1"`
	// LoggerDevMode sets the logger to use (or not use) development mode (more human-readable output, extra stack traces
	// and logging information, etc).
	LoggerDevMode bool `envconfig:"default=false"`
	// MockGraphQL sets the grapql servers to use mocked data
	MockGraphQL bool `envconfig:"default=false"`

	GraphQLGateway GraphQLGateway
}

type GraphQLGateway struct {
	Endpoint string `envconfig:"default=http://voltron-gateway/graphql"`
	Username string
	Password string
}

func main() {
	// init configuration
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	logger := zap.NewRaw(zap.UseDevMode(cfg.LoggerDevMode))

	// setup controller
	ctrl.SetLogger(zapr.NewLogger(logger))

	err = clientgoscheme.AddToScheme(scheme)
	exitOnError(err, "while adding k8s scheme")
	err = corev1alpha1.AddToScheme(scheme)
	exitOnError(err, "while adding core Action scheme")

	k8sCfg := ctrl.GetConfigOrDie()
	mgr, err := ctrl.NewManager(k8sCfg, ctrl.Options{
		Scheme:                  scheme,
		LeaderElection:          cfg.EnableLeaderElection,
		LeaderElectionNamespace: cfg.LeaderElectionNamespace,
		LeaderElectionID:        "152f0254.projectvoltron.dev",
		MetricsBindAddress:      cfg.MetricsAddr,
		HealthProbeBindAddress:  cfg.HealthzAddr,
	})
	exitOnError(err, "while creating manager")

	ochClient := getOCHClient(&cfg)
	actionSvc := controller.NewActionService(mgr.GetClient(), time.Minute)

	actionCtrl := controller.NewActionReconciler(ctrl.Log, mgr.GetClient(), mgr.GetEventRecorderFor("action-controller"), ochClient, actionSvc)
	err = actionCtrl.SetupWithManager(mgr, cfg.MaxConcurrentReconciles)
	exitOnError(err, "while creating controller")

	// setup instrumentation
	err = mgr.AddHealthzCheck("ping", healthz.Ping)
	exitOnError(err, "while adding healthz check")

	// setup GraphQL server
	k8sCli, err := client.New(k8sCfg, client.Options{Scheme: scheme})
	exitOnError(err, "while creating K8s client")

	gqlLogger := logger.Named(GraphQLServerName)

	var execSchema gqlgen_graphql.ExecutableSchema
	if cfg.MockGraphQL {
		logger.Info("Using mocked version of engine API")
		execSchema = graphql.NewExecutableSchema(graphql.Config{
			Resolvers: domaingraphql.NewMockedRootResolver(),
		})
	} else {
		execSchema = graphql.NewExecutableSchema(graphql.Config{
			Resolvers: domaingraphql.NewRootResolver(gqlLogger, k8sCli),
		})
	}
	gqlSrv := gqlServer(gqlLogger, execSchema, cfg.GraphQLAddr, GraphQLServerName)

	err = mgr.Add(gqlSrv)
	exitOnError(err, "while adding GraphQL server")

	// start
	setupLog.Info("starting manager")
	err = mgr.Start(ctrl.SetupSignalHandler())
	exitOnError(err, "while running manager")
}

func getOCHClient(cfg *Config) *ochclient.Client {
	httpClient := httputil.NewClient(30*time.Second, false,
		httputil.WithBasicAuth(cfg.GraphQLGateway.Username, cfg.GraphQLGateway.Password))
	return ochclient.NewClient(cfg.GraphQLGateway.Endpoint, httpClient)
}

func gqlServer(log *uber_zap.Logger, execSchema gqlgen_graphql.ExecutableSchema, addr, name string) httputil.StartableServer {
	nsMiddleware := namespace.NewMiddleware()

	gqlRouter := graphqlutil.NewGraphQLRouter(execSchema, name)
	gqlRouter.Use(nsMiddleware.Handle)

	return httputil.NewStartableServer(
		log.With(uber_zap.String("server", "graphql")),
		addr,
		gqlRouter,
	)
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
