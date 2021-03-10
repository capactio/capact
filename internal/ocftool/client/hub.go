package client

import (
	"fmt"
	"time"

	"projectvoltron.dev/voltron/internal/ocftool/credstore"
	"projectvoltron.dev/voltron/pkg/httputil"
	generated "projectvoltron.dev/voltron/pkg/och/client/public/generated"
)

func NewHub(server string) (*generated.Client, error) {
	store := credstore.NewOCH()
	user, pass, err := store.Get(server)
	if err != nil {
		return nil, err
	}

	httpClient := httputil.NewClient(30*time.Second, false,
		httputil.WithBasicAuth(user, pass))

	return generated.NewClient(httpClient, fmt.Sprintf("%s/graphql", server)), nil
}
