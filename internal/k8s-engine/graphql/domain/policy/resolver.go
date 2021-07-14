package policy

import (
	"context"

	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/policy"
	"github.com/pkg/errors"
)

// Service allows to get and update Capact Policy.
type Service interface {
	Update(ctx context.Context, in policy.Policy) (policy.Policy, error)
	Get(ctx context.Context) (policy.Policy, error)
}

type policyConverter interface {
	FromGraphQLInput(in graphql.PolicyInput) (policy.Policy, error)
	ToGraphQL(in policy.Policy) graphql.Policy
}

// Resolver provides functionality to manage Capact Policy via GraphQL.
type Resolver struct {
	svc  Service
	conv policyConverter
}

// NewResolver returns a new Resolver instance.
func NewResolver(svc Service, conv policyConverter) *Resolver {
	return &Resolver{
		svc:  svc,
		conv: conv,
	}
}

// UpdatePolicy updates Capact Policy on cluster side.
func (r *Resolver) UpdatePolicy(ctx context.Context, in graphql.PolicyInput) (*graphql.Policy, error) {
	p, err := r.conv.FromGraphQLInput(in)
	if err != nil {
		return nil, errors.Wrap(err, "while getting policy from GraphQL input")
	}

	p, err = r.svc.Update(ctx, p)
	if err != nil {
		return nil, errors.Wrap(err, "while updating Policy")
	}

	gqlPolicy := r.conv.ToGraphQL(p)
	return &gqlPolicy, nil
}

// Policy returns Capact Policy.
func (r *Resolver) Policy(ctx context.Context) (*graphql.Policy, error) {
	currentPolicy, err := r.svc.Get(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Policy")
	}

	gqlPolicy := r.conv.ToGraphQL(currentPolicy)
	return &gqlPolicy, nil
}
