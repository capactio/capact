package graphqlutil

import (
	"net/http"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gorilla/mux"
)

// NewGraphQLRouter returns gorilla router for a GraphQL server.
func NewGraphQLRouter(execSchema graphql.ExecutableSchema, name string) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/", playground.Handler(name, "/graphql")).Methods(http.MethodGet)
	r.Handle("/graphql", handler.NewDefaultServer(execSchema)).Methods(http.MethodPost)

	return r
}
