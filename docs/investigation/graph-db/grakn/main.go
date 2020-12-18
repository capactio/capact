package main

import (
	"log"
	"net/http"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"projectvoltron.dev/voltron/poc/graph-db/grakn/graphql"
)

func main() {

	gqlCfg := graphql.Config{
		Resolvers: NewRootResolver(),
	}

	executableSchema := graphql.NewExecutableSchema(gqlCfg)
	srv := handler.NewDefaultServer(executableSchema)

	http.Handle("/", playground.Handler("grakn PoC", "/graphql"))
	http.Handle("/graphql", srv)

	log.Println("Server started")
	log.Fatal(http.ListenAndServe(":3000", nil))
}
