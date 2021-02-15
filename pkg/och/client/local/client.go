package local

import (
	"context"
	"fmt"

	"github.com/avast/retry-go"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
)

const retryAttempts = 1

// Client used to communicate with the Voltron Local OCH GraphQL APIs
type Client struct {
	client *graphql.Client
}

func NewClient(cli *graphql.Client) *Client {
	return &Client{client: cli}
}

func (c *Client) CreateTypeInstance(ctx context.Context, in *ochlocalgraphql.CreateTypeInstanceInput) (*ochlocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`mutation($in: CreateTypeInstanceInput!) {
		createTypeInstance(
			in: $in
		) {
			%s
		}
	}`, typeInstanceFields)

	req := graphql.NewRequest(query)
	req.Var("in", in)

	var resp struct {
		TypeInstance ochlocalgraphql.TypeInstance `json:"createTypeInstance"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create TypeInstance")
	}

	return &resp.TypeInstance, nil
}

func (c *Client) CreateTypeInstances(ctx context.Context, in *ochlocalgraphql.CreateTypeInstancesInput) ([]ochlocalgraphql.CreatedTypeInstanceID, error) {
	query := `mutation($in: CreateTypeInstancesInput!) {
		createTypeInstances(
			in: $in
		) {
			alias
			id
		}
	}`

	req := graphql.NewRequest(query)
	req.Var("in", in)

	var resp struct {
		CreatedTypeInstances []ochlocalgraphql.CreatedTypeInstanceID `json:"createTypeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create TypeInstances")
	}

	return resp.CreatedTypeInstances, nil
}

func (c *Client) GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`query($id: ID!) {
		typeInstance(id: $id) {
			%s	
		}
	}`, typeInstanceFields)

	req := graphql.NewRequest(query)
	req.Var("id", id)

	var resp struct {
		TypeInstance ochlocalgraphql.TypeInstance `json:"typeInstance"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to get TypeInstance")
	}

	return &resp.TypeInstance, nil
}

func (c *Client) DeleteTypeInstance(ctx context.Context, id string) error {
	req := graphql.NewRequest(`mutation ($id: ID!) {
	  deleteTypeInstance(
	    id: $id
	  )
	}`)
	req.Var("id", id)

	var resp struct{}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return errors.Wrap(err, "while executing query to get TypeInstance")
	}

	return nil
}
