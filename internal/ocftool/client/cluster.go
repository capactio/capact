package client

import (
	"context"
	"fmt"
	"time"

	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	enginegraphql "projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/client"
	"projectvoltron.dev/voltron/pkg/httputil"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"projectvoltron.dev/voltron/pkg/och/client/local"

	"github.com/machinebox/graphql"
)

type ClusterClient interface {
	TypeInstanceClient
	ActionClient
}

type ActionClient interface {
	CreateAction(ctx context.Context, in *enginegraphql.ActionDetailsInput) (*enginegraphql.Action, error)
	GetAction(ctx context.Context, name string) (*enginegraphql.Action, error)
	ListActions(ctx context.Context) ([]*enginegraphql.Action, error)
	RunAction(ctx context.Context, name string) error
}

type TypeInstanceClient interface {
	FindTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
}

type clusterClient struct {
	TypeInstanceClient
	ActionClient
}

func NewCluster(server string) (ClusterClient, error) {
	creds, err := credstore.GetHub(server)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/graphql", server)

	httpClient := httputil.NewClient(30*time.Second,
		httputil.WithBasicAuth(creds.Username, creds.Secret))

	gqlClient := graphql.NewClient(endpoint, graphql.WithHTTPClient(httpClient))

	return &clusterClient{
		TypeInstanceClient: local.NewClient(gqlClient),
		ActionClient:       client.New(endpoint, httpClient),
	}, nil
}
