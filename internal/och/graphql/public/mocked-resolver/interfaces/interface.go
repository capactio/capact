package interfaces

import (
	"context"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type InterfaceResolver struct{}

func NewResolver() *InterfaceResolver {
	return &InterfaceResolver{}
}

func (r *InterfaceResolver) Interfaces(ctx context.Context, filter *gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error) {
	ifaces, err := mockedresolver.MockedInterfaces()
	if err != nil {
		return []*gqlpublicapi.Interface{}, err
	}
	return ifaces, nil
}

func (r *InterfaceResolver) Interface(ctx context.Context, path string) (*gqlpublicapi.Interface, error) {
	ifaces, err := mockedresolver.MockedInterfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		if i.Path == path {
			return i, nil
		}
	}
	return nil, nil
}

func (r *InterfaceResolver) Revision(ctx context.Context, obj *gqlpublicapi.Interface, revision string) (*gqlpublicapi.InterfaceRevision, error) {
	if obj == nil {
		return nil, nil
	}
	for _, ir := range obj.Revisions {
		if ir.Revision == revision {
			return ir, nil
		}
	}
	return nil, nil
}
