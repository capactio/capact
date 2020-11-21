package local

import (
	"projectvoltron.dev/voltron/internal/och/graphql/local/mocked-resolver/typeinstance"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
)

var _ gqllocalapi.ResolverRoot = &MockedRootResolver{}

type MockedRootResolver struct {
	mutationResolver gqllocalapi.MutationResolver
	queryResolver    gqllocalapi.QueryResolver
}

func NewMockedRootResolver() *RootResolver {
	instanceResolver := typeinstance.NewResolver()
	return &RootResolver{
		mutationResolver: instanceResolver,
		queryResolver:    instanceResolver,
	}
}

func (r MockedRootResolver) Mutation() gqllocalapi.MutationResolver {
	return r.mutationResolver
}

func (r MockedRootResolver) Query() gqllocalapi.QueryResolver {
	return r.queryResolver
}
