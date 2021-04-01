package client

import (
	"context"
	"net/http"

	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	"github.com/machinebox/graphql"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
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
	CreateTypeInstances(ctx context.Context, in *ochlocalgraphql.CreateTypeInstancesInput) ([]ochlocalgraphql.CreateTypeInstanceOutput, error)
	FindTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
	ListTypeInstances(ctx context.Context, filter *ochlocalgraphql.TypeInstanceFilter) ([]ochlocalgraphql.TypeInstance, error)
	ListTypeInstancesTypeRef(ctx context.Context) ([]ochlocalgraphql.TypeInstanceTypeReference, error)
	DeleteTypeInstance(ctx context.Context, id string) error
	LockTypeInstances(ctx context.Context, in *ochlocalgraphql.LockTypeInstancesInput) error
	UnlockTypeInstances(ctx context.Context, in *ochlocalgraphql.UnlockTypeInstancesInput) error
	UpdateTypeInstances(ctx context.Context, in []ochlocalgraphql.UpdateTypeInstancesInput) ([]ochlocalgraphql.TypeInstance, error)
}

type Public interface {
	ListInterfacesMetadata(ctx context.Context) ([]ochpublicgraphql.Interface, error)
	GetInterfaceLatestRevisionString(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (string, error)
	FindInterfaceRevision(ctx context.Context, ref ochpublicgraphql.InterfaceReference) (*ochpublicgraphql.InterfaceRevision, error)
	ListImplementationRevisionsForInterface(ctx context.Context, ref ochpublicgraphql.InterfaceReference, opts ...public.GetImplementationOption) ([]ochpublicgraphql.ImplementationRevision, error)
	ListInterfacesWithLatest(ctx context.Context, filter ochpublicgraphql.InterfaceFilter) ([]*ochpublicgraphql.Interface, error)
}

func New(endpoint string, httpClient *http.Client) *Client {
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return &Client{
		Local:  local.NewClient(client),
		Public: public.NewClient(client),
	}
}
