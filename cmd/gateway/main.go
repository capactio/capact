package main

import (
	"log"
	"net/http"
	"time"

	"github.com/avast/retry-go"
	"github.com/gorilla/mux"
	"github.com/nautilus/gateway"
	"github.com/nautilus/graphql"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"projectvoltron.dev/voltron/internal/healthz"
	"projectvoltron.dev/voltron/pkg/httputil"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/vrischmann/envconfig"
)

// Config holds application related configuration.
type Config struct {
	// GraphQLAddr is the TCP address the GraphQL endpoint binds to.
	GraphQLAddr string `envconfig:"default=:8080"`

	// HealthzAddr is the TCP address the health probes endpoint binds to.
	HealthzAddr string `envconfig:"default=:8082"`

	// LoggerDevMode sets the logger to use (or not use) development mode (more human-readable output, extra stack traces
	// and logging information, etc).
	LoggerDevMode bool `envconfig:"default=false"`

	// Introspection holds configuration parameters related to GraphQL schema introspection.
	Introspection IntrospectionConfig
}

// IntrospectionConfig holds configuration parameters related to GraphQL schema introspection.
type IntrospectionConfig struct {
	// GraphQLEndpoints contains list of remote GraphQL API endpoints to introspect and merge into one unified GraphQL endpoint.
	// Endpoints have to be separated by comma, e.g. `http://localhost:3000/graphql,http://localhost:3001/graphql`
	GraphQLEndpoints []string

	// Attempts specifies how many attempts are done to successfully introspect GraphQL schemas for provided endpoints.
	Attempts uint `envconfig:"default=120"`

	// RetryDelay defines how many time it should wait before new attempt to introspect schemas.
	RetryDelay time.Duration `envconfig:"default=1s"`
}

const appName = "gateway"

func main() {
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

	unnamedLogger, err := logCfg.Build()
	exitOnError(err, "while creating zap logger")

	logger := unnamedLogger.Named(appName)

	parallelServers := new(errgroup.Group)

	// healthz server
	healthzServer := healthz.NewHTTPServer(logger, cfg.HealthzAddr, appName)
	parallelServers.Go(func() error { return healthzServer.Start(stop) })

	// graphql server
	schemas, err := introspectGraphQLSchemas(logger, cfg.Introspection)
	exitOnError(err, "while introspecting GraphQL schemas")

	gqlServer, err := setupGatewayServerFromSchemas(logger, schemas, cfg.GraphQLAddr)
	exitOnError(err, "while gateway setup")

	parallelServers.Go(func() error { return gqlServer.Start(stop) })

	err = parallelServers.Wait()
	exitOnError(err, "while waiting for servers to finish gracefully")
}

func introspectGraphQLSchemas(log *zap.Logger, cfg IntrospectionConfig) ([]*graphql.RemoteSchema, error) {
	log.Info("Introspecting GraphQL schemas",
		zap.Strings("URLs", cfg.GraphQLEndpoints),
		zap.Uint("attempts", cfg.Attempts),
		zap.Duration("retry delay", cfg.RetryDelay),
	)

	var schemas []*graphql.RemoteSchema
	var err error

	err = retry.Do(
		func() error {
			schemas, err = graphql.IntrospectRemoteSchemas(cfg.GraphQLEndpoints...)
			return errors.Wrap(err, "while introspecting schemas")
		},
		retry.OnRetry(func(n uint, err error) {
			log.Debug("Retry attempt", zap.Uint("attempt no", n+1), zap.Error(err))
		}),
		retry.Attempts(cfg.Attempts),
		retry.Delay(cfg.RetryDelay),
		retry.DelayType(retry.FixedDelay),
		retry.LastErrorOnly(true),
	)
	if err != nil {
		return nil, errors.Wrap(err, "while introspecting schemas with retry")
	}

	return schemas, nil
}

func setupGatewayServerFromSchemas(log *zap.Logger, schemas []*graphql.RemoteSchema, addr string) (httputil.StartableServer, error) {
	log.Info("Setting up gateway GraphQL server")
	gw, err := gateway.New(schemas)
	if err != nil {
		return nil, errors.Wrap(err, "while creating gateway")
	}

	router := mux.NewRouter()
	// TODO: Remove redirect after https://github.com/nautilus/gateway/issues/120
	router.Handle("/", http.RedirectHandler("/graphql", http.StatusTemporaryRedirect)).Methods(http.MethodGet)
	router.HandleFunc("/graphql", gw.PlaygroundHandler).Methods(http.MethodGet, http.MethodPost)

	gqlServer := httputil.NewStartableServer(
		log.With(zap.String("server", "graphql")),
		addr,
		router,
	)

	return gqlServer, nil
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
