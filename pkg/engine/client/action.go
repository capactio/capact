package client

import (
	"context"
	"fmt"

	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	gqlengine "capact.io/capact/pkg/engine/api/graphql"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

// Action knows how to execute GraphQL queries and mutations for Action type.
type Action struct {
	client *graphql.Client
}

// CreateAction creates Action in the Namespace extracted from a given ctx.
func (c *Action) CreateAction(ctx context.Context, in *gqlengine.ActionDetailsInput) (*gqlengine.Action, error) {
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
		Action gqlengine.Action `json:"createAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create Action")
	}

	return &resp.Action, nil
}

// GetAction returns Action with a given name from Namespace extracted from a given ctx.
func (c *Action) GetAction(ctx context.Context, name string) (*gqlengine.Action, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query($name: String!) {
		action(name: $name) {
			%s
		}
	}`, actionFields))

	c.enrichWithNamespace(ctx, req)
	req.Var("name", name)

	var resp struct {
		Action *gqlengine.Action `json:"action"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to get Action")
	}

	return resp.Action, nil
}

// ListActions returns all Actions which meet filter criteria.
// Namespace extracted from a given ctx.
func (c *Action) ListActions(ctx context.Context, filter *gqlengine.ActionFilter) ([]*gqlengine.Action, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query($filter: ActionFilter) {
		actions(filter: $filter) {
			%s
		}
	}`, actionFields))

	c.enrichWithNamespace(ctx, req)
	req.Var("filter", filter)

	var resp struct {
		Actions []*gqlengine.Action `json:"actions"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to get Action")
	}

	return resp.Actions, nil
}

// RunAction executes a given Action.
func (c *Action) RunAction(ctx context.Context, name string) error {
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
		Action gqlengine.Action `json:"runAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing mutation to run Action")
	}

	return nil
}

// DeleteAction deletes a given Action.
func (c *Action) DeleteAction(ctx context.Context, name string) error {
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
		Action gqlengine.Action `json:"deleteAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing mutation to delete Action")
	}

	return nil
}

func (c *Action) enrichWithNamespace(ctx context.Context, req *graphql.Request) {
	ns, err := namespace.FromContext(ctx)
	if err != nil {
		return
	}
	req.Header.Add(namespace.NamespaceHeaderName, ns)
}
