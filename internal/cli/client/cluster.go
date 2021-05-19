package client

import (
	"context"
	"fmt"
	"strings"
	"time"

	"capact.io/capact/internal/cli/credstore"
	enginegraphql "capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/client"
	"capact.io/capact/pkg/httputil"
	ochlocalgraphql "capact.io/capact/pkg/och/api/graphql/local"
	"capact.io/capact/pkg/och/client/local"

	"github.com/machinebox/graphql"
)

type ClusterClient interface {
	TypeInstanceClient
	EngineClient
}

type EngineClient interface {
	CreateAction(ctx context.Context, in *enginegraphql.ActionDetailsInput) (*enginegraphql.Action, error)
	GetAction(ctx context.Context, name string) (*enginegraphql.Action, error)
	ListActions(ctx context.Context, filter *enginegraphql.ActionFilter) ([]*enginegraphql.Action, error)
	RunAction(ctx context.Context, name string) error
	DeleteAction(ctx context.Context, name string) error
	UpdatePolicy(ctx context.Context, policy *enginegraphql.PolicyInput) (*enginegraphql.Policy, error)
	GetPolicy(ctx context.Context) (*enginegraphql.Policy, error)
}

type TypeInstanceClient interface {
	FindTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
}

type clusterClient struct {
	TypeInstanceClient
	EngineClient
}

func NewCluster(server string) (ClusterClient, error) {
	creds, err := credstore.GetHub(server)
	if err != nil {
		return nil, err
	}

	return NewClusterWithCreds(server, creds)
}

func NewClusterWithCreds(server string, creds *credstore.Credentials) (ClusterClient, error) {
	endpoint := normalizeServerEndpoint(server)

	httpClient := httputil.NewClient(30*time.Second,
		httputil.WithBasicAuth(creds.Username, creds.Secret))

	gqlClient := graphql.NewClient(endpoint, graphql.WithHTTPClient(httpClient))

	return &clusterClient{
		TypeInstanceClient: local.NewClient(gqlClient),
		EngineClient:       client.New(endpoint, httpClient),
	}, nil
}

func normalizeServerEndpoint(server string) string {
	if strings.HasPrefix(server, "http://") || strings.HasPrefix(server, "https://") {
		return fmt.Sprintf("%s/graphql", server)
	}

	return fmt.Sprintf("https://%s/graphql", server)
}
