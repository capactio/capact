package client

import (
	"context"
	"fmt"

	"capact.io/capact/internal/cli/credstore"
	enginegraphql "capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/client"
	"capact.io/capact/pkg/httputil"
	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/hub/client/local"

	"github.com/machinebox/graphql"
)

// ClusterClient groups GraphQL operation that can be executed against Capact cluster.
type ClusterClient interface {
	TypeInstanceClient
	EngineClient
}

// EngineClient aggregates operations that are executed against Capact Engine by Capact CLI.
type EngineClient interface {
	CreateAction(ctx context.Context, in *enginegraphql.ActionDetailsInput) (*enginegraphql.Action, error)
	GetAction(ctx context.Context, name string) (*enginegraphql.Action, error)
	ListActions(ctx context.Context, filter *enginegraphql.ActionFilter) ([]*enginegraphql.Action, error)
	RunAction(ctx context.Context, name string) error
	DeleteAction(ctx context.Context, name string) error
	UpdatePolicy(ctx context.Context, policy *enginegraphql.PolicyInput) (*enginegraphql.Policy, error)
	GetPolicy(ctx context.Context) (*enginegraphql.Policy, error)
}

// TypeInstanceClient aggregates operations that are executed against Local Hub by Capact CLI.
type TypeInstanceClient interface {
	FindTypeInstance(ctx context.Context, id string, opts ...local.TypeInstancesOption) (*hublocalgraphql.TypeInstance, error)
}

type clusterClient struct {
	TypeInstanceClient
	EngineClient
}

// NewCluster returns client for Capact cluster configured with saved credentials for a given server URL.
func NewCluster(serverURL string) (ClusterClient, error) {
	creds, err := credstore.GetHub(serverURL)
	if err != nil {
		return nil, err
	}

	return NewClusterWithCreds(serverURL, creds)
}

// NewClusterWithCreds returns client for Capact cluster with custom credentials.
func NewClusterWithCreds(server string, creds *credstore.Credentials) (ClusterClient, error) {
	endpoint := fmt.Sprintf("%s/graphql", server)

	httpClient := httputil.NewClient(
		httputil.WithBasicAuth(creds.Username, creds.Secret))

	gqlClient := graphql.NewClient(endpoint, graphql.WithHTTPClient(httpClient))

	return &clusterClient{
		TypeInstanceClient: local.NewClient(gqlClient),
		EngineClient:       client.New(endpoint, httpClient),
	}, nil
}
