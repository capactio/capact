package client

import (
	"context"
	"fmt"

	gqlengine "capact.io/capact/pkg/engine/api/graphql"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// Policy knows how to execute GraphQL queries and mutations for Policy type.
type Policy struct {
	client *graphql.Client
}

// UpdatePolicy updates Capact Policy on cluster side.
func (c *Policy) UpdatePolicy(ctx context.Context, policy *gqlengine.PolicyInput) (*gqlengine.Policy, error) {
	req := graphql.NewRequest(fmt.Sprintf(`mutation($in: PolicyInput!) {
		updatePolicy(
			in: $in
		) {
			%s
		}
	}`, policyFields))
	req.Var("in", policy)

	var resp struct {
		Policy *gqlengine.Policy `json:"updatePolicy"`
	}
	if err := c.client.Run(ctx, req, nil); err != nil {
		return nil, errors.Wrap(err, "while executing mutation to update Policy")
	}

	return resp.Policy, nil
}

// GetPolicy returns current Capact Policy.
func (c *Policy) GetPolicy(ctx context.Context) (*gqlengine.Policy, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query {
		policy{
			%s
		}
	}`, policyFields))

	var resp struct {
		Policy *gqlengine.Policy `json:"policy"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to get Policy")
	}

	return resp.Policy, nil
}
