package graphql

import (
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/resolver/action"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
)

var _ graphql.ResolverRoot = &RootResolver{}

type RootResolver struct {
	mutationResolver graphql.MutationResolver
	queryResolver    graphql.QueryResolver
}

func NewRootResolver() *RootResolver {
	actionResolver := action.NewResolver()
	return &RootResolver{
		mutationResolver: actionResolver,
		queryResolver:    actionResolver,
	}
}

func (r RootResolver) Mutation() graphql.MutationResolver {
	return r.mutationResolver
}

func (r RootResolver) Query() graphql.QueryResolver {
	return r.queryResolver
}
