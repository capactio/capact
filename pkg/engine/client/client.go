package client

import (
	"context"
	"net/http"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	enginegraphql "projectvoltron.dev/voltron/pkg/engine/api/graphql"
)

// Client used to communicate with the Voltron Engine GraphQL API
type Client struct {
	client *graphql.Client
}

func New(endpoint string, httpClient *http.Client) *Client {
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return &Client{
		client: client,
	}
}

func (c *Client) CreateAction(ctx context.Context, in *enginegraphql.ActionDetailsInput) (*enginegraphql.Action, error) {
	req := graphql.NewRequest(`mutation($in: ActionDetailsInput) {
		createAction(
			in: $in
		) {
    	name
    	createdAt
    	input {
    	  parameters
    	  typeInstances {
    	    id
    	    name
    	    typeRef {
    	      path
    	      revision
    	    }
    	  }
    	}
    	output {
    	  typeInstances {
    	    id
    	    name
    	    typeRef {
    	      path
    	      revision
    	    }
    	  }
    	}
		}
	}`)

	req.Var("in", in)
	var resp struct {
		Action enginegraphql.Action `json:"createAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to create Action")
	}

	return &resp.Action, nil
}

func (c *Client) GetAction(ctx context.Context, name string) (*enginegraphql.Action, error) {
	req := graphql.NewRequest(`query($name: String!) {
		action(name: $name) {
    	name
    	createdAt
    	input {
    	  parameters
    	  typeInstances {
    	    id
    	    name
    	    typeRef {
    	      path
    	      revision
    	    }
    	  }
    	}
    	output {
    	  typeInstances {
    	    id
    	    name
    	    typeRef {
    	      path
    	      revision
    	    }
    	  }
    	}
    	actionRef {
    	  path
    	  revision
    	}
    	run
    	cancel
    	dryRun
			renderedAction
			status {
				condition
			}
		}
	}`)

	req.Var("name", name)
	var resp struct {
		Action enginegraphql.Action `json:"action"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to create Action")
	}

	return &resp.Action, nil
}

func (c *Client) RunAction(ctx context.Context, name string) error {
	req := graphql.NewRequest(`mutation($name: String!) {
		runAction(
			name: $name
		) {
    	name
		}
	}`)

	req.Var("name", name)
	var resp struct {
		Action enginegraphql.Action `json:"runAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing query to create Action")
	}

	return nil
}

func (c *Client) DeleteAction(ctx context.Context, name string) error {
	req := graphql.NewRequest(`mutation($name: String!) {
		deleteAction(
			name: $name
		) {
    	name
		}
	}`)

	req.Var("name", name)
	var resp struct {
		Action enginegraphql.Action `json:"deleteAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing query to create Action")
	}

	return nil
}
