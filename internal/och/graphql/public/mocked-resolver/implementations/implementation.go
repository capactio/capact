package implementations

import (
	"context"
	"fmt"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type ImplementationResolver struct {
}

func NewResolver() *ImplementationResolver {
	return &ImplementationResolver{}
}
func (i *ImplementationResolver) Implementations(ctx context.Context, filter *gqlpublicapi.ImplementationFilter) ([]*gqlpublicapi.Implementation, error) {
	iface, err := getInterface()
	if err != nil {
		return []*gqlpublicapi.Implementation{}, err
	}
	return iface.LatestRevision.Implementations, nil
}

func (i ImplementationResolver) Implementation(ctx context.Context, path string) (*gqlpublicapi.Implementation, error) {
	iface, err := getInterface()
	if err != nil {
		return &gqlpublicapi.Implementation{}, err
	}
	for _, implementation := range iface.LatestRevision.Implementations {
		if implementation.Path == path {
			return implementation, nil
		}
	}
	return &gqlpublicapi.Implementation{}, fmt.Errorf("Implementation with path %s not found", path)
}

func (r *ImplementationResolver) Revision(ctx context.Context, obj *gqlpublicapi.Implementation, revision string) (*gqlpublicapi.ImplementationRevision, error) {
	iface, err := getInterface()
	if err != nil {
		return &gqlpublicapi.ImplementationRevision{}, err
	}
	for _, implementation := range iface.LatestRevision.Implementations {
		if implementation.Revision.Revision == revision {
			return implementation.Revision, nil
		}
	}
	return &gqlpublicapi.ImplementationRevision{}, fmt.Errorf("No Implementation with revision %s", revision)
}

func getInterface() (*gqlpublicapi.Interface, error) {
	iface, err := mockedresolver.MockedInterface()
	if err != nil {
		return &gqlpublicapi.Interface{}, err
	}
	for _, implementation := range iface.LatestRevision.Implementations {
		implementation.LatestRevision = implementation.Revisions[0]
	}
	return iface, nil
}
