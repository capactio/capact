package main

import (
	"log"

	"github.com/go-logr/zapr"
	"github.com/vrischmann/envconfig"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"projectvoltron.dev/voltron/internal/graphqlutil"
	"projectvoltron.dev/voltron/internal/k8s-engine/controller"
	domaingraphql "projectvoltron.dev/voltron/internal/k8s-engine/graphql"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	corev1alpha1 "projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
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

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                  scheme,
		LeaderElection:          cfg.EnableLeaderElection,
		LeaderElectionNamespace: cfg.LeaderElectionNamespace,
		LeaderElectionID:        "152f0254.projectvoltron.dev",
		MetricsBindAddress:      cfg.MetricsAddr,
		HealthProbeBindAddress:  cfg.HealthzAddr,
	})
	exitOnError(err, "while creating manager")

	actionCtrl := controller.NewActionReconciler(mgr.GetClient(), ctrl.Log.WithName("controllers").WithName("Action"))
	err = actionCtrl.SetupWithManager(mgr, cfg.MaxConcurrentReconciles)
	exitOnError(err, "while creating controller")

	// setup instrumentation
	err = mgr.AddHealthzCheck("ping", healthz.Ping)
	exitOnError(err, "while adding healthz check")

	// setup graphql server
	execSchema := graphql.NewExecutableSchema(graphql.Config{
		Resolvers: domaingraphql.NewRootResolver(),
	})

	gsvr := graphqlutil.NewHTTPServer(logger, execSchema, cfg.GraphQLAddr, "Engine GraphQL API")
	err = mgr.Add(gsvr)
	exitOnError(err, "while adding GraphQL server")

	// start
	setupLog.Info("starting manager")
	err = mgr.Start(ctrl.SetupSignalHandler())
	exitOnError(err, "while running manager")
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
