package local

import (
	"bytes"
	"context"
	"fmt"

	"capact.io/capact/pkg/httputil"

	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"github.com/avast/retry-go"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

const retryAttempts = 1

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
	httpClient := httputil.NewClient(opts...)
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return NewClient(client)
}

// CreateTypeInstance creates a new TypeInstances in the local Hub.
func (c *Client) CreateTypeInstance(ctx context.Context, in *hublocalgraphql.CreateTypeInstanceInput) (string, error) {
	req := graphql.NewRequest(`mutation CreateTypeInstance($in: CreateTypeInstanceInput!) {
		createTypeInstance(
			in: $in
		)
	}`)
	req.Var("in", in)

	var resp struct {
		CreatedTypeInstance string `json:"createTypeInstance"`
	}
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return "", errors.Wrap(err, "while executing mutation to create TypeInstance")
	}

	return resp.CreatedTypeInstance, nil
}

// CreateTypeInstances creates new TypeInstances and allows to define "uses" relationships between them.
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

// UpdateTypeInstances updates multiple TypeInstances in the local Hub.
func (c *Client) UpdateTypeInstances(ctx context.Context, in []hublocalgraphql.UpdateTypeInstancesInput, opts ...TypeInstancesOption) ([]hublocalgraphql.TypeInstance, error) {
	tiOpts := newTypeInstancesOptions(TypeInstanceAllFields)
	tiOpts.Apply(opts...)

	query := fmt.Sprintf(`mutation UpdateTypeInstances($in: [UpdateTypeInstancesInput]!) {
		updateTypeInstances(
			in: $in
		) {
			%s
		}
	}`, tiOpts.fields)

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

// FindTypeInstance finds a TypeInstance with the given ID. If no TypeInstance is found, it returns nil.
func (c *Client) FindTypeInstance(ctx context.Context, id string, opts ...TypeInstancesOption) (*hublocalgraphql.TypeInstance, error) {
	tiOpts := newTypeInstancesOptions(TypeInstanceAllFieldsWithRelations)
	tiOpts.Apply(opts...)

	query := fmt.Sprintf(`query FindTypeInstance($id: ID!) {
		typeInstance(id: $id) {
			%s
		}
	}`, tiOpts.fields)

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

// FindTypeInstancesTypeRef finds TypeRef for all specified TypeInstance IDs.
// If no TypeInstances are found, it returns nil.
func (c *Client) FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]hublocalgraphql.TypeInstanceTypeReference, error) {
	if len(ids) == 0 {
		return map[string]hublocalgraphql.TypeInstanceTypeReference{}, nil
	}

	body := bytes.Buffer{}
	for idx, id := range ids {
		body.WriteString(fmt.Sprintf(`
		id_%d:typeInstance(id: %q) {
			id
			typeRef {
			  path
			  revision
			}
		}`, idx, id))
	}

	req := graphql.NewRequest(fmt.Sprintf(`query FindTypeInstancesTypeRef {
		%s
	}`, body.String()))

	var resp map[string]*hublocalgraphql.TypeInstance
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to get TypeInstances TypeRefs")
	}

	out := map[string]hublocalgraphql.TypeInstanceTypeReference{}
	for _, ti := range resp {
		if ti == nil || ti.TypeRef == nil {
			continue
		}
		out[ti.ID] = *ti.TypeRef
	}

	return out, nil
}

// FindTypeInstances finds TypeInstance based on IDs.
// If no TypeInstances are found, it returns nil.
func (c *Client) FindTypeInstances(ctx context.Context, ids []string, opts ...TypeInstancesOption) (map[string]hublocalgraphql.TypeInstance, error) {
	tiOpts := newTypeInstancesOptions(TypeInstanceAllFieldsWithRelations)
	tiOpts.Apply(opts...)

	if len(ids) == 0 {
		return nil, nil
	}

	body := bytes.Buffer{}
	for idx, id := range ids {
		body.WriteString(fmt.Sprintf(`
		id_%d:typeInstance(id: %q) {
			%s
		}`, idx, id, tiOpts.fields))
	}

	req := graphql.NewRequest(fmt.Sprintf(`query FindTypeInstances {
		%s
	}`, body.String()))

	var resp map[string]*hublocalgraphql.TypeInstance
	err := retry.Do(func() error {
		return c.client.Run(ctx, req, &resp)
	}, retry.Attempts(retryAttempts))
	if err != nil {
		return nil, errors.Wrap(err, "while executing query to get TypeInstances based on IDs")
	}

	out := map[string]hublocalgraphql.TypeInstance{}
	for _, ti := range resp {
		if ti == nil {
			continue
		}
		out[ti.ID] = *ti
	}

	return out, nil
}

// ListTypeInstances lists the TypeInstances in the local Hub. You can pass a filter limit the list of returned TypeInstances.
func (c *Client) ListTypeInstances(ctx context.Context, filter *hublocalgraphql.TypeInstanceFilter, opts ...TypeInstancesOption) ([]hublocalgraphql.TypeInstance, error) {
	tiOpts := newTypeInstancesOptions(TypeInstanceAllFieldsWithRelations)
	tiOpts.Apply(opts...)

	query := fmt.Sprintf(`query ListTypeInstances($filter: TypeInstanceFilter) {
		typeInstances(filter: $filter) {
			%s
		}
	}`, tiOpts.fields)

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

// ListTypeInstancesTypeRef lists TypeInstances with only the TypeReference fields filled. It can be used to determine,
// if a TypeInstance of a given TypeReference exists in the local Hub.
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

// DeleteTypeInstance deletes a TypeInstances from the local Hub.
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

// LockTypeInstances locks the given TypeInstances. It will return an error, if the TypeInstance is already locked by an another owner.
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

// UnlockTypeInstances unlocks the given TypeInstances. It will return an error, if the TypeInstances are locked by a different owner.
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
