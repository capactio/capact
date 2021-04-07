package client

import (
	"context"
	"fmt"
	"time"

	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	"projectvoltron.dev/voltron/pkg/httputil"
	gqllocalapi "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/och/client"
	"projectvoltron.dev/voltron/pkg/och/client/public"
)

type Hub interface {
	ListInterfacesWithLatest(ctx context.Context, filter gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error)
	ListTypeInstances(ctx context.Context, filter *gqllocalapi.TypeInstanceFilter) ([]gqllocalapi.TypeInstance, error)
	ListImplementationRevisionsForInterface(ctx context.Context, ref gqlpublicapi.InterfaceReference, opts ...public.GetImplementationOption) ([]gqlpublicapi.ImplementationRevision, error)
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
