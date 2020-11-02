package graphqlutil

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"projectvoltron.dev/voltron/pkg/httputil"
)

func NewHTTPServer(log *zap.Logger, execSchema graphql.ExecutableSchema, addr, name string) httputil.StartableServer {
	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", playground.Handler(name, "/graphql")).Methods(http.MethodGet)
	mainRouter.Handle("/graphql", handler.NewDefaultServer(execSchema)).Methods(http.MethodPost)

	return httputil.NewStartableServer(
		log.Named(name).With(zap.String("server", "graphql")),
		addr,
		mainRouter,
	)
}
