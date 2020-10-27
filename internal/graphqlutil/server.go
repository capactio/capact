package graphqlutil

import (
	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"projectvoltron.dev/voltron/pkg/httputil"
)

func NewHTTPServer(log *zap.Logger, execSchema graphql.ExecutableSchema, addr, name string) httputil.StartableServer {
	mainRouter := mux.NewRouter()
	mainRouter.HandleFunc("/", playground.Handler(name, "/graphql")).Methods("GET")
	mainRouter.Handle("/graphql", handler.NewDefaultServer(execSchema)).Methods("POST")

	return httputil.NewStartableServer(
		log.Named(name).With(zap.String("server", "graphql")),
		addr,
		mainRouter,
	)
}
