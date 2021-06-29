package local

import (
	"context"
	"fmt"
	"time"

	"capact.io/capact/pkg/httputil"

	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"github.com/avast/retry-go"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const (
	retryAttempts      = 1
	httpRequestTimeout = 30 * time.Second
)

// Client used to communicate with the Capact Local Hub GraphQL APIs
type Client struct {
	client *graphql.Client
}

// NewClient creates a local client with a given graphql custom client instance.
func NewClient(cli *graphql.Client) *Client {
	return &Client{client: cli}
}

// NewDefaultClient creates ready to use client with default values.
func NewDefaultClient(endpoint string, opts ...httputil.ClientOption) *Client {
	httpClient := httputil.NewClient(
		httpRequestTimeout,
	)
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return NewClient(client)
}

func (c *Client) CreateTypeInstance(ctx context.Context, in *hublocalgraphql.CreateTypeInstanceInput) (*hublocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`mutation CreateTypeInstance($in: CreateTypeInstanceInput!) {
		createTypeInstance(
			in: $in
		) {
			%s
		}
	}`, typeInstanceWithUsesFields)

	req := graphql.NewRequest(query)
	req.Var("in", in)

	var resp struct {
		TypeInstance *hublocalgraphql.TypeInstance `json:"createTypeInstance"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create TypeInstance")
	}

	return resp.TypeInstance, nil
}

func (c *Client) CreateTypeInstances(ctx context.Context, in *hublocalgraphql.CreateTypeInstancesInput) ([]hublocalgraphql.CreateTypeInstanceOutput, error) {
	query := `mutation CreateTypeInstances($in: CreateTypeInstancesInput!) {
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
		CreatedTypeInstances []hublocalgraphql.CreateTypeInstanceOutput `json:"createTypeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create TypeInstances")
	}

	return resp.CreatedTypeInstances, nil
}

func (c *Client) UpdateTypeInstances(ctx context.Context, in []hublocalgraphql.UpdateTypeInstancesInput) ([]hublocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`mutation UpdateTypeInstances($in: [UpdateTypeInstancesInput]!) {
		updateTypeInstances(
			in: $in
		) {
			%s
		}
	}`, typeInstanceFields)

	req := graphql.NewRequest(query)
	req.Var("in", in)

	var resp struct {
		TypeInstances []hublocalgraphql.TypeInstance `json:"updateTypeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing mutation to update TypeInstances")
	}

	return resp.TypeInstances, nil
}

func (c *Client) FindTypeInstance(ctx context.Context, id string) (*hublocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`query FindTypeInstance($id: ID!) {
		typeInstance(id: $id) {
			%s	
		}
	}`, typeInstanceWithUsesFields)

	req := graphql.NewRequest(query)
	req.Var("id", id)

	var resp struct {
		TypeInstance *hublocalgraphql.TypeInstance `json:"typeInstance"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to get TypeInstance")
	}

	return resp.TypeInstance, nil
}

func (c *Client) ListTypeInstances(ctx context.Context, filter *hublocalgraphql.TypeInstanceFilter) ([]hublocalgraphql.TypeInstance, error) {
	query := fmt.Sprintf(`query ListTypeInstances($filter: TypeInstanceFilter) {
		typeInstances(filter: $filter) {
			%s	
		}
	}`, typeInstanceWithUsesFields)

	req := graphql.NewRequest(query)
	req.Var("filter", filter)

	var resp struct {
		TypeInstances []hublocalgraphql.TypeInstance `json:"typeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to list TypeInstances")
	}

	return resp.TypeInstances, nil
}

func (c *Client) ListTypeInstancesTypeRef(ctx context.Context) ([]hublocalgraphql.TypeInstanceTypeReference, error) {
	query := `query ListTypeInstancesTypeRef {
	  typeInstances {
		  typeRef {
			path
			revision
		  }
	  }
	}`

	req := graphql.NewRequest(query)

	var resp struct {
		TypeInstances []hublocalgraphql.TypeInstance `json:"typeInstances"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to list TypeRef for TypeInstances")
	}

	var typeRefs []hublocalgraphql.TypeInstanceTypeReference
	for _, ti := range resp.TypeInstances {
		if ti.TypeRef == nil {
			continue
		}

		typeRefs = append(typeRefs, *ti.TypeRef)
	}

	return typeRefs, nil
}

func (c *Client) DeleteTypeInstance(ctx context.Context, id string) error {
	req := graphql.NewRequest(`mutation DeleteTypeInstance($id: ID!) {
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
		return errors.Wrap(err, "while executing mutation to delete TypeInstance")
	}

	return nil
}

func (c *Client) LockTypeInstances(ctx context.Context, in *hublocalgraphql.LockTypeInstancesInput) error {
	query := `mutation LockTypeInstances($in: LockTypeInstancesInput!) {
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

func (c *Client) UnlockTypeInstances(ctx context.Context, in *hublocalgraphql.UnlockTypeInstancesInput) error {
	query := `mutation UnlockTypeInstances($in: UnlockTypeInstancesInput!) {
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
