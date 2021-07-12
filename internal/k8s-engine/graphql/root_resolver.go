package graphql

import (
	"capact.io/capact/internal/k8s-engine/graphql/domain/action"
	"capact.io/capact/internal/k8s-engine/graphql/domain/policy"
	"capact.io/capact/pkg/engine/api/graphql"
	"go.uber.org/zap"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var _ graphql.ResolverRoot = &RootResolver{}

// RootResolver aggregates all query and mutation resolver for Capact Engine domain.
type RootResolver struct {
	combinedResolver combinedResolver
}

// NewRootResolver returns a new RootResolver instance.
func NewRootResolver(log *zap.Logger, k8sCli client.Client, policyService policy.Service) *RootResolver {
	actionConverter := action.NewConverter()
	actionService := action.NewService(log, k8sCli)
	actionResolver := action.NewResolver(actionService, actionConverter)

	policyConverter := policy.NewConverter()
	policyResolver := policy.NewResolver(policyService, policyConverter)

	return &RootResolver{
		combinedResolver{
			actionResolver: actionResolver,
			policyResolver: policyResolver,
		},
	}
}

// Mutation returns Capact Engine mutation resolvers.
func (r RootResolver) Mutation() graphql.MutationResolver {
	return r.combinedResolver
}

// Query returns Capact Engine query resolvers.
func (r RootResolver) Query() graphql.QueryResolver {
	return r.combinedResolver
}

type actionResolver = action.Resolver
type policyResolver = policy.Resolver

type combinedResolver struct {
	*actionResolver
	*policyResolver
}
