package och

import (
	"github.com/99designs/gqlgen/graphql"

	gqllocaldomain "projectvoltron.dev/voltron/internal/och/graphql/local"
	gqlpublicdomain "projectvoltron.dev/voltron/internal/och/graphql/public"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

// GraphQLSchema returns the proper GraphQL schema based on the given mode
func GraphQLSchema(mode Mode) graphql.ExecutableSchema {
	switch mode {
	case PublicMode:
		cfg := gqlpublicapi.Config{
			Resolvers: gqlpublicdomain.NewRootResolver(),
		}
		return gqlpublicapi.NewExecutableSchema(cfg)
	case LocalMode:
		cfg := gqllocalapi.Config{
			Resolvers: gqllocaldomain.NewRootResolver(),
		}
		return gqllocalapi.NewExecutableSchema(cfg)
	}

	return nil
}
