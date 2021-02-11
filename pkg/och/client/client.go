package client

import (
	"context"
	"net/http"

	"github.com/machinebox/graphql"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/och/client/local"
	"projectvoltron.dev/voltron/pkg/och/client/public"
)

// Client used to communicate with the Voltron OCH GraphQL APIs
type Client struct {
	Local
	Public
}

type Local interface {
	CreateTypeInstance(ctx context.Context, in *ochlocalgraphql.CreateTypeInstanceInput) (*ochlocalgraphql.TypeInstance, error)
	CreateTypeInstances(ctx context.Context, in *ochlocalgraphql.CreateTypeInstancesInput) ([]string, error)
	ListTypeInstances(ctx context.Context, filter ochlocalgraphql.TypeInstanceFilter) ([]ochlocalgraphql.TypeInstance, error)
	GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
	DeleteTypeInstance(ctx context.Context, id string) error
}

type Public interface {
	ListInterfacesMetadata(ctx context.Context) ([]ochpublicgraphql.Interface, error)
	GetImplementationRevisionsForInterface(ctx context.Context, ref ochpublicgraphql.InterfaceReference, opts ...public.GetImplementationOption) ([]ochpublicgraphql.ImplementationRevision, error)
}

func NewClient(endpoint string, httpClient *http.Client) *Client {
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return &Client{
		Local:  local.NewClient(client),
		Public: public.NewClient(client),
	}
}
