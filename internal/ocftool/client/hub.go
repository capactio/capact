package client

import (
	"context"
	"fmt"
	"time"

	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	"projectvoltron.dev/voltron/pkg/httputil"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/och/client/public"

	"github.com/machinebox/graphql"
)

type Hub interface {
	ListInterfacesWithLatest(ctx context.Context, filter gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error)
}

func NewHub(server string) (*public.Client, error) {
	creds, err := credstore.GetHub(server)
	if err != nil {
		return nil, err
	}

	httpClient := httputil.NewClient(30*time.Second, false,
		httputil.WithBasicAuth(creds.Username, creds.Secret))

	clientOpt := graphql.WithHTTPClient(httpClient)
	client := graphql.NewClient(fmt.Sprintf("%s/graphql", server), clientOpt)

	return public.NewClient(client), nil
}
