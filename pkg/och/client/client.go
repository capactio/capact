package client

import (
	"context"
	"net/http"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	ochgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
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

// TODO simple implementation for demo, does return only some fields
// TODO: add support for not found errors
func (c *Client) GetLatestRevisionOfImplementationForInterface(ctx context.Context, path string) (*ochgraphql.ImplementationRevision, error) {
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
		Interface ochgraphql.Interface `json:"interface"`
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
