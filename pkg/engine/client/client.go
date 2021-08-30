package client

import (
	"context"
	"fmt"

	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	enginegraphql "capact.io/capact/pkg/engine/api/graphql"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// Client used to communicate with the Capact Engine GraphQL API
type Client struct {
	client *graphql.Client
}

// New returns a new Client instance.
func New(gqlClient *graphql.Client) *Client {
	return &Client{
		client: gqlClient,
	}
}

// CreateAction creates Action in the Namespace extracted from a given ctx.
func (c *Client) CreateAction(ctx context.Context, in *enginegraphql.ActionDetailsInput) (*enginegraphql.Action, error) {
	req := graphql.NewRequest(fmt.Sprintf(`mutation($in: ActionDetailsInput) {
		createAction(
			in: $in
		) {
			%s
		}
	}`, actionFields))

	c.enrichWithNamespace(ctx, req)
	req.Var("in", in)

	var resp struct {
		Action enginegraphql.Action `json:"createAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create Action")
	}

	return &resp.Action, nil
}

// GetAction returns Action with a given name from Namespace extracted from a given ctx.
func (c *Client) GetAction(ctx context.Context, name string) (*enginegraphql.Action, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query($name: String!) {
		action(name: $name) {
			%s
		}
	}`, actionFields))

	c.enrichWithNamespace(ctx, req)
	req.Var("name", name)

	var resp struct {
		Action *enginegraphql.Action `json:"action"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to get Action")
	}

	return resp.Action, nil
}

// ListActions returns all Actions which meet filter criteria.
// Namespace extracted from a given ctx.
func (c *Client) ListActions(ctx context.Context, filter *enginegraphql.ActionFilter) ([]*enginegraphql.Action, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query($filter: ActionFilter) {
		actions(filter: $filter) {
			%s
		}
	}`, actionFields))

	c.enrichWithNamespace(ctx, req)
	req.Var("filter", filter)

	var resp struct {
		Actions []*enginegraphql.Action `json:"actions"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to get Action")
	}

	return resp.Actions, nil
}

// RunAction executes a given Action.
func (c *Client) RunAction(ctx context.Context, name string) error {
	req := graphql.NewRequest(fmt.Sprintf(`mutation($name: String!) {
		runAction(
			name: $name
		) {
			%s
		}
	}`, actionFields))

	c.enrichWithNamespace(ctx, req)
	req.Var("name", name)

	var resp struct {
		Action enginegraphql.Action `json:"runAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing mutation to run Action")
	}

	return nil
}

// DeleteAction deletes a given Action.
func (c *Client) DeleteAction(ctx context.Context, name string) error {
	req := graphql.NewRequest(fmt.Sprintf(`mutation($name: String!) {
		deleteAction(
			name: $name
		) {
			%s
		}
	}`, actionFields))

	c.enrichWithNamespace(ctx, req)
	req.Var("name", name)

	var resp struct {
		Action enginegraphql.Action `json:"deleteAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing mutation to delete Action")
	}

	return nil
}

// UpdatePolicy updates Capact Policy on cluster side.
func (c *Client) UpdatePolicy(ctx context.Context, policy *enginegraphql.PolicyInput) (*enginegraphql.Policy, error) {
	req := graphql.NewRequest(fmt.Sprintf(`mutation($in: PolicyInput!) {
		updatePolicy(
			in: $in
		) {
			%s
		}
	}`, policyFields))
	req.Var("in", policy)

	var resp struct {
		Policy *enginegraphql.Policy `json:"updatePolicy"`
	}
	if err := c.client.Run(ctx, req, nil); err != nil {
		return nil, errors.Wrap(err, "while executing mutation to update Policy")
	}

	return resp.Policy, nil
}

// GetPolicy returns current Capact Policy.
func (c *Client) GetPolicy(ctx context.Context) (*enginegraphql.Policy, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query {
		policy{
			%s
		}
	}`, policyFields))

	var resp struct {
		Policy *enginegraphql.Policy `json:"policy"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to get Policy")
	}

	return resp.Policy, nil
}

func (c *Client) enrichWithNamespace(ctx context.Context, req *graphql.Request) {
	ns, err := namespace.FromContext(ctx)
	if err != nil {
		return
	}
	req.Header.Add(namespace.NamespaceHeaderName, ns)
}
