package graphqlutil

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"capact.io/capact/pkg/httputil"
)

func NewHTTPServer(log *zap.Logger, execSchema graphql.ExecutableSchema, addr, name string) httputil.StartableServer {
	return httputil.NewStartableServer(
		log.Named(name).With(zap.String("server", "graphql")),
		addr,
		NewGraphQLRouter(execSchema, name),
	)
}

func NewGraphQLRouter(execSchema graphql.ExecutableSchema, name string) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", playground.Handler(name, "/graphql")).Methods(http.MethodGet)
	r.Handle("/graphql", handler.NewDefaultServer(execSchema)).Methods(http.MethodPost)

	return r
}
