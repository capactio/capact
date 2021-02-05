package client

import (
	"context"
	"fmt"
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
	req := graphql.NewRequest(fmt.Sprintf(`mutation($in: ActionDetailsInput) {
		createAction(
			in: $in
		) {
			%s
		}
	}`, actionFields))

	req.Var("in", in)
	var resp struct {
		Action enginegraphql.Action `json:"createAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing mutation to create Action")
	}

	return &resp.Action, nil
}

func (c *Client) GetAction(ctx context.Context, name string) (*enginegraphql.Action, error) {
	req := graphql.NewRequest(fmt.Sprintf(`query($name: String!) {
		action(name: $name) {
			%s
		}
	}`, actionFields))

	req.Var("name", name)
	var resp struct {
		Action enginegraphql.Action `json:"action"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errors.Wrap(err, "while executing query to get Action")
	}

	return &resp.Action, nil
}

func (c *Client) RunAction(ctx context.Context, name string) error {
	req := graphql.NewRequest(fmt.Sprintf(`mutation($name: String!) {
		runAction(
			name: $name
		) {
			%s
		}
	}`, actionFields))

	req.Var("name", name)
	var resp struct {
		Action enginegraphql.Action `json:"runAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing mutation to run Action")
	}

	return nil
}

func (c *Client) DeleteAction(ctx context.Context, name string) error {
	req := graphql.NewRequest(fmt.Sprintf(`mutation($name: String!) {
		deleteAction(
			name: $name
		) {
			%s
		}
	}`, actionFields))

	req.Var("name", name)
	var resp struct {
		Action enginegraphql.Action `json:"deleteAction"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return errors.Wrap(err, "while executing mutation to delete Action")
	}

	return nil
}

const actionFields = `
    name
    createdAt
    input {
        parameters
        typeInstances {
            id
            name
            optional
            typeRef {
                path
                revision
            }
        }
    }
    output {
        typeInstances {
            name
            typeRef {
                path
                revision
            }
            id
            name
        }
    }
    actionRef {
        path
        revision
    }
    cancel
    run
    dryRun
    renderedAction
    renderingAdvancedMode {
        enabled
        typeInstancesForRenderingIteration {
            name
            typeRef {
                path
                revision
            }
        }
    }
    renderedActionOverride
    status {
        phase
        timestamp
        message
        runner {
            status
        }
        canceledBy {
            username
            groups
            extra
        }
        runBy {
            username
            groups
            extra
        }
        createdBy {
            username
            groups
            extra
        }
    }
`
