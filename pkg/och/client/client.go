package client

import (
	"context"
	"net/http"

	"github.com/machinebox/graphql"
	errs "github.com/pkg/errors"
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
func (c *Client) GetImplementationLatestRevision(ctx context.Context, path string) (*ochgraphql.ImplementationRevision, error) {
	req := graphql.NewRequest(`query($implementationPath: NodePath!) {
	  implementation(path: $implementationPath) {
	    latestRevision {
	      spec {
	        action {
	          runnerInterface
	          args
	        }
	      }
	    }
	  }
	}`)
	req.Var("implementationPath", path)

	var resp struct {
		Implementation ochgraphql.Implementation `json:"implementation"`
	}
	if err := c.client.Run(ctx, req, &resp); err != nil {
		return nil, errs.Wrap(err, "while executing query to fetch OCH Implementation")
	}

	return resp.Implementation.LatestRevision, nil
}
