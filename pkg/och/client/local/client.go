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
	}`, typeInstanceWithUsesFields)

	req := graphql.NewRequest(query)
	req.Var("in", in)

	var resp struct {
		TypeInstance *ochlocalgraphql.TypeInstance `json:"createTypeInstance"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create TypeInstance")
	}

	return resp.TypeInstance, nil
}

func (c *Client) CreateTypeInstances(ctx context.Context, in *ochlocalgraphql.CreateTypeInstancesInput) ([]ochlocalgraphql.CreateTypeInstanceOutput, error) {
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
		CreatedTypeInstances []ochlocalgraphql.CreateTypeInstanceOutput `json:"createTypeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create TypeInstances")
	}

	return resp.CreatedTypeInstances, nil
}

func (c *Client) UpdateTypeInstances(ctx context.Context, in []ochlocalgraphql.UpdateTypeInstancesInput) ([]ochlocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`mutation($in: [UpdateTypeInstancesInput]!) {
		updateTypeInstances(
			in: $in
		) {
			%s
		}
	}`, typeInstanceFields)

	req := graphql.NewRequest(query)
	req.Var("in", in)

	var resp struct {
		TypeInstances []ochlocalgraphql.TypeInstance `json:"updateTypeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to update TypeInstances")
	}

	return resp.TypeInstances, nil
}

func (c *Client) FindTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`query($id: ID!) {
		typeInstance(id: $id) {
			%s	
		}
	}`, typeInstanceWithUsesFields)

	req := graphql.NewRequest(query)
	req.Var("id", id)

	var resp struct {
		TypeInstance *ochlocalgraphql.TypeInstance `json:"typeInstance"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to get TypeInstance")
	}

	return resp.TypeInstance, nil
}

func (c *Client) ListTypeInstances(ctx context.Context, filter *ochlocalgraphql.TypeInstanceFilter) ([]ochlocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`query($filter: TypeInstanceFilter) {
		typeInstances(filter: $filter) {
			%s	
		}
	}`, typeInstanceWithUsesFields)

	req := graphql.NewRequest(query)
	req.Var("filter", filter)

	var resp struct {
		TypeInstances []ochlocalgraphql.TypeInstance `json:"typeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to list TypeInstances")
	}

	return resp.TypeInstances, nil
}

func (c *Client) ListTypeInstancesTypeRef(ctx context.Context) ([]ochlocalgraphql.TypeInstanceTypeReference, error) {
	query := `query {
	  typeInstances {
		  typeRef {
			path
			revision
		  }
	  }
	}`

	req := graphql.NewRequest(query)

	var resp struct {
		TypeInstances []ochlocalgraphql.TypeInstance `json:"typeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to list TypeRef for TypeInstances")
	}

	var typeRefs []ochlocalgraphql.TypeInstanceTypeReference
	for _, ti := range resp.TypeInstances {
		if ti.TypeRef == nil {
			continue
		}

		typeRefs = append(typeRefs, *ti.TypeRef)
	}

	return typeRefs, nil
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

func (c *Client) LockTypeInstances(ctx context.Context, in *ochlocalgraphql.LockTypeInstanceInput) error {
	query := `mutation($in: LockTypeInstanceInput!) {
		lockTypeInstances(in: $in)
	}`

	req := graphql.NewRequest(query)
	req.Var("in", in)

	err := retry.Do(func() error {
		return c.client.Run(ctx, req, nil)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return errors.Wrap(err, "while executing mutation to lock TypeInstances")
	}

	return nil
}

func (c *Client) UnlockTypeInstances(ctx context.Context, in *ochlocalgraphql.UnlockTypeInstanceInput) error {
	query := `mutation($in: UnlockTypeInstanceInput!) {
		unlockTypeInstances(in: $in)
	}`

	req := graphql.NewRequest(query)
	req.Var("in", in)

	err := retry.Do(func() error {
		return c.client.Run(ctx, req, nil)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return errors.Wrap(err, "while executing mutation to unlock TypeInstances")
	}

	return nil
}

