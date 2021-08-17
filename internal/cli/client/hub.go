package client

import (
	"context"
	"fmt"

	"capact.io/capact/pkg/hub/client/local"
	"capact.io/capact/pkg/hub/client/public"

	"capact.io/capact/internal/cli/credstore"
	"capact.io/capact/pkg/httputil"
	gqllocalapi "capact.io/capact/pkg/hub/api/graphql/local"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client"
)

// Hub aggregates operation executed by Capact CLI against Capact Hub server.
type Hub interface {
	ListInterfaces(ctx context.Context, opts ...public.InterfaceOption) ([]*gqlpublicapi.Interface, error)
	ListTypeInstances(ctx context.Context, filter *gqllocalapi.TypeInstanceFilter, opts ...local.TypeInstancesOption) ([]gqllocalapi.TypeInstance, error)
	ListImplementationRevisions(ctx context.Context, opts ...public.ListImplementationRevisionsOption) ([]*gqlpublicapi.ImplementationRevision, error)
	FindTypeInstance(ctx context.Context, id string, opts ...local.TypeInstancesOption) (*gqllocalapi.TypeInstance, error)
	CreateTypeInstance(ctx context.Context, in *gqllocalapi.CreateTypeInstanceInput, opts ...local.TypeInstancesOption) (*gqllocalapi.TypeInstance, error)
	CreateTypeInstances(ctx context.Context, in *gqllocalapi.CreateTypeInstancesInput) ([]gqllocalapi.CreateTypeInstanceOutput, error)
	UpdateTypeInstances(ctx context.Context, in []gqllocalapi.UpdateTypeInstancesInput, opts ...local.TypeInstancesOption) ([]gqllocalapi.TypeInstance, error)
	DeleteTypeInstance(ctx context.Context, id string) error
	ListTypeRefRevisionsJSONSchemas(ctx context.Context, filter gqlpublicapi.TypeFilter) ([]*gqlpublicapi.TypeRevision, error)
	FindInterfaceRevision(ctx context.Context, ref gqlpublicapi.InterfaceReference, opts ...public.InterfaceRevisionOption) (*gqlpublicapi.InterfaceRevision, error)
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]gqllocalapi.TypeInstanceTypeReference, error)
	CheckManifestRevisionsExist(ctx context.Context, manifestRefs []gqlpublicapi.ManifestReference) (map[gqlpublicapi.ManifestReference]bool, error)
}

// NewHub returns client for Capact Hub configured with saved credentials for a given server URL.
func NewHub(server string) (Hub, error) {
	creds, err := credstore.GetHub(server)
	if err != nil {
		return nil, err
	}

	httpClient := httputil.NewClient(
		httputil.WithBasicAuth(creds.Username, creds.Secret),
		httputil.WithTimeout(timeout))

	return client.New(fmt.Sprintf("%s/graphql", server), httpClient), nil
}
