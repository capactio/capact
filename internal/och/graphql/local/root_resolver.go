package local

import (
	"projectvoltron.dev/voltron/internal/och/graphql/local/resolver/typeinstance"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
)

var _ gqllocalapi.ResolverRoot = &RootResolver{}

type RootResolver struct {
	mutationResolver gqllocalapi.MutationResolver
	queryResolver    gqllocalapi.QueryResolver
}

func NewRootResolver() *RootResolver {
	instanceResolver := typeinstance.NewResolver()
	return &RootResolver{
		mutationResolver: instanceResolver,
		queryResolver:    instanceResolver,
	}
}

func (r RootResolver) Mutation() gqllocalapi.MutationResolver {
	return r.mutationResolver
}

func (r RootResolver) Query() gqllocalapi.QueryResolver {
	return r.queryResolver
}
