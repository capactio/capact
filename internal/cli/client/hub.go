package client

import (
	"context"
	"fmt"
	"time"

	"capact.io/capact/internal/cli/credstore"
	"capact.io/capact/pkg/httputil"
	gqllocalapi "capact.io/capact/pkg/och/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/och/api/graphql/public"
	"capact.io/capact/pkg/och/client"
	"capact.io/capact/pkg/och/client/public"
)

type Hub interface {
	ListInterfacesWithLatestRevision(ctx context.Context, filter gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error)
	ListTypeInstances(ctx context.Context, filter *gqllocalapi.TypeInstanceFilter) ([]gqllocalapi.TypeInstance, error)
	ListImplementationRevisions(ctx context.Context, filter *gqlpublicapi.ImplementationRevisionFilter) ([]*gqlpublicapi.ImplementationRevision, error)
	FindTypeInstance(ctx context.Context, id string) (*gqllocalapi.TypeInstance, error)
	CreateTypeInstances(ctx context.Context, in *gqllocalapi.CreateTypeInstancesInput) ([]gqllocalapi.CreateTypeInstanceOutput, error)
	UpdateTypeInstances(ctx context.Context, in []gqllocalapi.UpdateTypeInstancesInput) ([]gqllocalapi.TypeInstance, error)
	DeleteTypeInstance(ctx context.Context, id string) error
}

func NewHub(server string) (Hub, error) {
	creds, err := credstore.GetHub(server)
	if err != nil {
		return nil, err
	}

	httpClient := httputil.NewClient(30*time.Second,
		httputil.WithBasicAuth(creds.Username, creds.Secret))

	return client.New(fmt.Sprintf("%s/graphql", server), httpClient), nil
}
