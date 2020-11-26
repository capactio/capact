package gateway

import (
	"context"

	"github.com/machinebox/graphql"
	errs "github.com/pkg/errors"
	ochgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type Client struct {
	client *graphql.Client
}

func NewClient(endpoint string) *Client {
	return &Client{
		client: graphql.NewClient(endpoint),
	}
}

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
		// TODO improve error handling
		return nil, errs.Wrap(err, "failed to query gateway endpoint")
	}

	return resp.Implementation.LatestRevision, nil
}
