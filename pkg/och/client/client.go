package client

import (
	"context"
	"net/http"

	ochpublicgraphql "capact.io/capact/pkg/och/api/graphql/public"

	ochlocalgraphql "capact.io/capact/pkg/och/api/graphql/local"
	"capact.io/capact/pkg/och/client/local"
	"capact.io/capact/pkg/och/client/public"
	"github.com/machinebox/graphql"
)

// Client used to communicate with the Capact OCH GraphQL APIs
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
	ListInterfacesWithLatestRevision(ctx context.Context, filter ochpublicgraphql.InterfaceFilter) ([]*ochpublicgraphql.Interface, error)
	ListImplementationRevisions(ctx context.Context, filter *ochpublicgraphql.ImplementationRevisionFilter) ([]*ochpublicgraphql.ImplementationRevision, error)
}

func New(endpoint string, httpClient *http.Client) *Client {
	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(endpoint, clientOpt)

	return &Client{
		Local:  local.NewClient(client),
		Public: public.NewClient(client),
	}
}
