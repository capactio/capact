package main

import (
	"log"
	"time"

	"capact.io/capact/internal/k8s-engine/policy"

	"capact.io/capact/internal/graphqlutil"
	"capact.io/capact/internal/k8s-engine/controller"
	domaingraphql "capact.io/capact/internal/k8s-engine/graphql"
	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	"capact.io/capact/internal/k8s-engine/validate"
	"capact.io/capact/internal/logger"
	"capact.io/capact/pkg/engine/api/graphql"
	corev1alpha1 "capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"capact.io/capact/pkg/httputil"
	ochclient "capact.io/capact/pkg/och/client"
	"capact.io/capact/pkg/sdk/renderer"
	"capact.io/capact/pkg/sdk/renderer/argo"
	gqlgen_graphql "github.com/99designs/gqlgen/graphql"
	wfclientset "github.com/argoproj/argo/v2/pkg/client/clientset/versioned"
	"github.com/go-logr/zapr"
	"github.com/vrischmann/envconfig"
	uber_zap "go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

const (
	GraphQLServerName = "engine-graphql"
	PolicyServiceName = "policy-svc"
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
	// MaxConcurrentReconciles is the maximum number of concurrent Reconciles which can be run.
	MaxConcurrentReconciles int `envconfig:"default=1"`
	// MaxRetryForFailedAction is the maximum number of concurrent Reconciles which can be run.
	MaxRetryForFailedAction int `envconfig:"default=15"`

	Logger logger.Config

	GraphQLGateway struct {
		Endpoint string `envconfig:"default=http://capact-gateway/graphql"`
		Username string
		Password string
	}

	BuiltinRunner controller.BuiltinRunnerConfig

	Policy policy.Config

	Renderer        renderer.Config
	OCHActionsImage string
}

func main() {
	// init configuration
	var cfg Config
	err := envconfig.InitWithPrefix(&cfg, "APP")
	exitOnError(err, "while loading configuration")

	logger, err := logger.New(cfg.Logger)
	exitOnError(err, "while creating zap logger")

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
		LeaderElectionID:        "152f0254.capact.io",
		MetricsBindAddress:      cfg.MetricsAddr,
		HealthProbeBindAddress:  cfg.HealthzAddr,
	})
	exitOnError(err, "while creating manager")

	ochClient := getOCHClient(&cfg)
	policyEnforcedClient := ochclient.NewPolicyEnforcedClient(ochClient)
	typeInstanceHandler := argo.NewTypeInstanceHandler(cfg.OCHActionsImage)
	argoRenderer := argo.NewRenderer(cfg.Renderer, policyEnforcedClient, typeInstanceHandler)

	wfCli, err := wfclientset.NewForConfig(k8sCfg)
	exitOnError(err, "while creating Argo client")
	actionValidator := validate.NewActionValidator(wfCli)

	policySvcLogger := logger.Named(PolicyServiceName)
	policyService := policy.NewService(policySvcLogger, mgr.GetClient(), cfg.Policy)

	actionSvc := controller.NewActionService(
		logger,
		mgr.GetClient(),
		argoRenderer,
		actionValidator,
		policyService,
		ochClient,
		ochClient,
		controller.Config{
			BuiltinRunner: cfg.BuiltinRunner,
		},
	)

	actionCtrl := controller.NewActionReconciler(ctrl.Log, actionSvc, cfg.MaxRetryForFailedAction)
	err = actionCtrl.SetupWithManager(mgr, cfg.MaxConcurrentReconciles)
	exitOnError(err, "while creating controller")

	// setup instrumentation
	err = mgr.AddHealthzCheck("ping", healthz.Ping)
	exitOnError(err, "while adding healthz check")

	// setup GraphQL server
	k8sCli, err := client.New(k8sCfg, client.Options{Scheme: scheme})
	exitOnError(err, "while creating K8s client")

	gqlLogger := logger.Named(GraphQLServerName)

	execSchema := graphql.NewExecutableSchema(graphql.Config{
		Resolvers: domaingraphql.NewRootResolver(gqlLogger, k8sCli, policyService),
	})
	gqlSrv := gqlServer(gqlLogger, execSchema, cfg.GraphQLAddr, GraphQLServerName)

	err = mgr.Add(gqlSrv)
	exitOnError(err, "while adding GraphQL server")

	// start
	setupLog.Info("starting manager")
	err = mgr.Start(ctrl.SetupSignalHandler())
	exitOnError(err, "while running manager")
}

func getOCHClient(cfg *Config) *ochclient.Client {
	httpClient := httputil.NewClient(30*time.Second,
		httputil.WithBasicAuth(cfg.GraphQLGateway.Username, cfg.GraphQLGateway.Password))
	return ochclient.New(cfg.GraphQLGateway.Endpoint, httpClient)
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
