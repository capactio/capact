package client

import (
	"context"
	"fmt"
	"time"

	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	enginegraphql "projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/client"
	"projectvoltron.dev/voltron/pkg/httputil"
)

type ClusterClient interface {
	CreateAction(ctx context.Context, in *enginegraphql.ActionDetailsInput) (*enginegraphql.Action, error)
	GetAction(ctx context.Context, name string) (*enginegraphql.Action, error)
	RunAction(ctx context.Context, name string) error
}

func NewCluster(server string) (ClusterClient, error) {
	store := credstore.NewOCH()
	user, pass, err := store.Get(server)
	if err != nil {
		return nil, err
	}

	httpClient := httputil.NewClient(30*time.Second, false,
		httputil.WithBasicAuth(user, pass))

	return client.New(fmt.Sprintf("%s/graphql", server), httpClient), nil
}
