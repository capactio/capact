package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"capact.io/capact/internal/gateway/header"
	"capact.io/capact/internal/healthz"
	"capact.io/capact/internal/logger"
	"capact.io/capact/pkg/httputil"

	"github.com/avast/retry-go"
	"github.com/gorilla/mux"
	"github.com/nautilus/gateway"
	"github.com/nautilus/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"
)

// Config holds application related configuration.
type Config struct {
	// GraphQLAddr is the TCP address the GraphQL endpoint binds to.
	GraphQLAddr string `envconfig:"default=:8080"`

	// HealthzAddr is the TCP address the health probes endpoint binds to.
	HealthzAddr string `envconfig:"default=:8082"`

	Logger logger.Config

	// Introspection holds configuration parameters related to GraphQL schema introspection.
	Introspection IntrospectionConfig

	// Auth holds configuration parameters for user authentication
	Auth BasicAuth
}

// BasicAuth holds the credentials for HTTP basic access authentication.
type BasicAuth struct {
	Username string `envconfig:"default=graphql"`
	Password string
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
	unnamedLogger, err := logger.New(cfg.Logger)
	exitOnError(err, "while creating zap logger")

	logger := unnamedLogger.Named(appName)

	parallelServers := new(errgroup.Group)

	// healthz server
	healthzServer := healthz.NewHTTPServer(logger, cfg.HealthzAddr, appName)
	parallelServers.Go(func() error { return healthzServer.Start(stop) })

	// graphql server
	schemas, err := introspectGraphQLSchemas(logger, cfg.Introspection)
	exitOnError(err, "while introspecting GraphQL schemas")

	gqlServer, err := setupGatewayServerFromSchemas(logger, schemas, cfg.Auth, cfg.GraphQLAddr)
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

func setupGatewayServerFromSchemas(log *zap.Logger, schemas []*graphql.RemoteSchema, authCfg BasicAuth, addr string) (httputil.StartableServer, error) {
	log.Info("Setting up gateway GraphQL server")

	headerMiddleware := header.Middleware{}

	middlewares := []gateway.Middleware{
		gateway.RequestMiddleware(
			headerMiddleware.RestoreFromCtx(),
		)}
	gw, err := gateway.New(schemas, gateway.WithMiddlewares(middlewares...))
	if err != nil {
		return nil, errors.Wrap(err, "while creating gateway")
	}

	router := mux.NewRouter()
	// TODO: Remove redirect after https://github.com/nautilus/gateway/issues/120
	router.Handle("/", http.RedirectHandler("/graphql", http.StatusTemporaryRedirect)).Methods(http.MethodGet)
	// TODO: Replace with proper authentication mechanism
	gatewayHandler := withBasicAuth(log, authCfg,
		headerMiddleware.StoreInCtx(
			http.HandlerFunc(gw.PlaygroundHandler),
		),
	)
	router.HandleFunc("/graphql", gatewayHandler).Methods(http.MethodGet, http.MethodPost)

	gqlServer := httputil.NewStartableServer(
		log.With(zap.String("server", "graphql")),
		addr,
		router,
	)

	return gqlServer, nil
}

func withBasicAuth(log *zap.Logger, cfg BasicAuth, handler http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handler.ServeHTTP(w, r)
			return
		}

		username, password, ok := r.BasicAuth()

		if !ok {
			if err := writeJSONError(w, "missing credentials", http.StatusOK); err != nil {
				log.Info("failed to write response")
			}
			return
		}

		if username != cfg.Username || password != cfg.Password {
			if err := writeJSONError(w, "wrong credentials", http.StatusOK); err != nil {
				log.Info("failed to write response")
			}
			return
		}

		handler.ServeHTTP(w, r)
	}
}

func writeJSONError(w http.ResponseWriter, message string, statusCode int) error {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(statusCode)
	return json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []map[string]interface{}{
			{
				"message": message,
			},
		},
	})
}

func exitOnError(err error, context string) {
	if err != nil {
		log.Fatalf("%s: %v", context, err)
	}
}
