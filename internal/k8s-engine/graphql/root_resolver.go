package graphql

import (
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/domain/action"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ graphql.ResolverRoot = &RootResolver{}

type RootResolver struct {
	mutationResolver graphql.MutationResolver
	queryResolver    graphql.QueryResolver
}

func NewRootResolver(k8sCli client.Client) *RootResolver {
	actionConverter := action.NewConverter()
	actionService := action.NewService(k8sCli)
	actionResolver := action.NewResolver(actionService, actionConverter)

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
