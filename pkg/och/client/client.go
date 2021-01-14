package client

import (
	"context"
	"net/http"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

// Client used to communicate with the Voltron OCH GraphQL APIs
// TODO this should be split into public and local OCH clients and composed together here
type Client struct {
	client *graphql.Client
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return &Client{
		client: client,
	}
}

func (c *Client) GetInterfaces(ctx context.Context) ([]ochpublicgraphql.Interface, error) {
	req := graphql.NewRequest(`query {
		interfaces {
			name
			prefix
			path
		}		
	}`)

	var resp struct {
		Interfaces []ochpublicgraphql.Interface `json:"interfaces"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch OCH Implementation")
	}

	return resp.Interfaces, nil
}

// TODO simple implementation for demo, does return only some fields
// TODO: add support for not found errors
func (c *Client) GetLatestRevisionOfImplementationForInterface(ctx context.Context, path string) (*ochpublicgraphql.ImplementationRevision, error) {
	req := graphql.NewRequest(`query($interfacePath: NodePath!) {
		  interface(path: $interfacePath) {
			latestRevision {
			  implementations {
				latestRevision {
				  spec {
					action {
					  runnerInterface
					  args
					}
				  }
				}
			  }
			}
		  }
		}`)

	req.Var("interfacePath", path)
	var resp struct {
		Interface ochpublicgraphql.Interface `json:"interface"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to fetch OCH Implementation")
	}
	if resp.Interface.LatestRevision == nil {
		return nil, errors.New("Interface.LatestRevision cannot be nil")
	}
	if len(resp.Interface.LatestRevision.Implementations) == 0 {
		return nil, errors.New("Interface.LatestRevision.Implementations cannot be nil")
	}

	return resp.Interface.LatestRevision.Implementations[0].LatestRevision, nil
}

func (c *Client) CreateTypeInstance(ctx context.Context, in *ochlocalgraphql.CreateTypeInstanceInput) (*ochlocalgraphql.TypeInstance, error) {
	req := graphql.NewRequest(`mutation($in: CreateTypeInstanceInput!) {
	  createTypeInstance(
	    in: $in
	  ) {
	    resourceVersion
	    metadata {
	      id
	      attributes {
	        path
	        revision
	      }
	    }
	    spec {
	      typeRef {
	        path
	        revision
	      }
	      value
	    }
	  }
	}`)
	req.Var("in", in)

	var resp struct {
		TypeInstance ochlocalgraphql.TypeInstance `json:"createTypeInstance"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to create TypeInstance")
	}

	return &resp.TypeInstance, nil
}

func (c *Client) GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error) {
	req := graphql.NewRequest(`query($id: ID!) {
	  typeInstance(id: $id) {
	    resourceVersion
	    metadata {
	      id
	      attributes {
	        path
	        revision
	      }
	    }
	    spec {
	      typeRef {
	        path
	        revision
	      }
	      value
	    }
	  }
	}`)
	req.Var("id", id)

	var resp struct {
		TypeInstance ochlocalgraphql.TypeInstance `json:"typeInstance"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
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
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing query to get TypeInstance")
	}

	return nil
}
