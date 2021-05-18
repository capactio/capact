package policy

import (
	"context"

	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/clusterpolicy"
	"github.com/pkg/errors"
)

type Service interface {
	Update(ctx context.Context, in clusterpolicy.ClusterPolicy) (clusterpolicy.ClusterPolicy, error)
	Get(ctx context.Context) (clusterpolicy.ClusterPolicy, error)
}

type policyConverter interface {
	FromGraphQLInput(in graphql.PolicyInput) clusterpolicy.ClusterPolicy
	ToGraphQL(in clusterpolicy.ClusterPolicy) graphql.Policy
}

type Resolver struct {
	svc  Service
	conv policyConverter
}

func NewResolver(svc Service, conv policyConverter) *Resolver {
	return &Resolver{
		svc:  svc,
		conv: conv,
	}
}

func (r *Resolver) UpdatePolicy(ctx context.Context, in graphql.PolicyInput) (*graphql.Policy, error) {
	policy := r.conv.FromGraphQLInput(in)

	policy, err := r.svc.Update(ctx, policy)
	if err != nil {
		return nil, errors.Wrap(err, "while updating Policy")
	}

	gqlPolicy := r.conv.ToGraphQL(policy)
	return &gqlPolicy, nil
}

func (r *Resolver) Policy(ctx context.Context) (*graphql.Policy, error) {
	policy, err := r.svc.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Policy")
	}

	gqlPolicy := r.conv.ToGraphQL(policy)
	return &gqlPolicy, nil
}
