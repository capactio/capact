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

type ImplementationRevisionResolver struct {
}

func NewRevisionResolver() *ImplementationRevisionResolver {
	return &ImplementationRevisionResolver{}
}

func (i *ImplementationResolver) Implementations(ctx context.Context, filter *gqlpublicapi.ImplementationFilter) ([]*gqlpublicapi.Implementation, error) {
	implementation, err := mockedresolver.MockedImplementation()
	if err != nil {
		return []*gqlpublicapi.Implementation{}, err
	}
	return []*gqlpublicapi.Implementation{implementation}, nil
}

func (i ImplementationResolver) Implementation(ctx context.Context, path string) (*gqlpublicapi.Implementation, error) {
	implementation, err := mockedresolver.MockedImplementation()
	if err != nil {
		return &gqlpublicapi.Implementation{}, err
	}
	return implementation, nil
}

func (i *ImplementationResolver) Revision(ctx context.Context, obj *gqlpublicapi.Implementation, revision string) (*gqlpublicapi.ImplementationRevision, error) {
	implementation, err := mockedresolver.MockedImplementation()
	if err != nil {
		return &gqlpublicapi.ImplementationRevision{}, err
	}
	for _, im := range implementation.Revisions {
		if im.Revision == revision {
			return im, nil
		}
	}
	return &gqlpublicapi.ImplementationRevision{}, fmt.Errorf("No Implementation with revision %s", revision)
}

func (i *ImplementationRevisionResolver) Interfaces(ctx context.Context, obj *gqlpublicapi.ImplementationRevision) ([]*gqlpublicapi.Interface, error) {
	iface, err := mockedresolver.MockedInterface()
	if err != nil {
		return []*gqlpublicapi.Interface{}, err
	}

	return []*gqlpublicapi.Interface{iface}, nil
}
