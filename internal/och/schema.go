package och

import (
	"github.com/99designs/gqlgen/graphql"

	gqllocaldomain "projectvoltron.dev/voltron/internal/och/graphql/local"
	gqlpublicdomain "projectvoltron.dev/voltron/internal/och/graphql/public"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

// GraphQLSchema returns the proper GraphQL schema based on the given mode
func GraphQLSchema(mode Mode, useMockedResolver bool) graphql.ExecutableSchema {
	switch mode {
	case PublicMode:
		var rootResolver gqlpublicapi.ResolverRoot
		if useMockedResolver {
			rootResolver = gqlpublicdomain.NewMockedRootResolver()
		} else {
			rootResolver = gqlpublicdomain.NewRootResolver()
		}
		cfg := gqlpublicapi.Config{
			Resolvers: rootResolver,
		}
		return gqlpublicapi.NewExecutableSchema(cfg)
	case LocalMode:
		var rootResolver gqllocalapi.ResolverRoot
		if useMockedResolver {
			rootResolver = gqllocaldomain.NewMockedRootResolver()
		} else {
			rootResolver = gqllocaldomain.NewRootResolver()
		}
		cfg := gqllocalapi.Config{
			Resolvers: rootResolver,
		}
		return gqllocalapi.NewExecutableSchema(cfg)
	}

	return nil
}
