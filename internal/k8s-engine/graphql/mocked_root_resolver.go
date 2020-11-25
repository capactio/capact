package graphql

import (
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/mocked-resolver/action"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
)

var _ graphql.ResolverRoot = &MockedRootResolver{}

type MockedRootResolver struct {
	mutationResolver graphql.MutationResolver
	queryResolver    graphql.QueryResolver
}

func NewMockedRootResolver() *MockedRootResolver {
	actionResolver := action.NewResolver()
	return &MockedRootResolver{
		mutationResolver: actionResolver,
		queryResolver:    actionResolver,
	}
}

func (r MockedRootResolver) Mutation() graphql.MutationResolver {
	return r.mutationResolver
}

func (r MockedRootResolver) Query() graphql.QueryResolver {
	return r.queryResolver
}
