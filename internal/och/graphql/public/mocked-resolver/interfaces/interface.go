package interfaces

import (
	"context"
	"fmt"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type InterfaceResolver struct{}

func NewResolver() *InterfaceResolver {
	return &InterfaceResolver{}
}

func (r *InterfaceResolver) Interfaces(ctx context.Context, filter *gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error) {
	i, err := mockedresolver.MockedInterface()
	if err != nil {
		return []*gqlpublicapi.Interface{}, err
	}
	return []*gqlpublicapi.Interface{i}, nil
}

func (r *InterfaceResolver) Interface(ctx context.Context, path string) (*gqlpublicapi.Interface, error) {
	i, err := mockedresolver.MockedInterface()
	if err != nil {
		return &gqlpublicapi.Interface{}, err
	}
	if i.Path == path {
		return i, nil
	}
	return &gqlpublicapi.Interface{}, nil
}

func (r *InterfaceResolver) Revision(ctx context.Context, obj *gqlpublicapi.Interface, revision string) (*gqlpublicapi.InterfaceRevision, error) {
	i, err := mockedresolver.MockedInterface()
	if err != nil {
		return &gqlpublicapi.InterfaceRevision{}, err
	}
	for _, ir := range i.Revisions {
		if ir.Revision == revision {
			return ir, nil
		}
	}
	return &gqlpublicapi.InterfaceRevision{}, fmt.Errorf("No Interface with revision %s", revision)
}
