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

type InterfaceRevisionResolver struct{}

func NewRevisionResolver() *InterfaceRevisionResolver {
	return &InterfaceRevisionResolver{}
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

func (r *InterfaceRevisionResolver) Implementations(ctx context.Context, obj *gqlpublicapi.InterfaceRevision, filter *gqlpublicapi.ImplementationFilter) ([]*gqlpublicapi.Implementation, error) {
	if obj == nil || obj.Metadata == nil || obj.Metadata.Path == nil {
		return []*gqlpublicapi.Implementation{}, nil
	}

	implementations, err := mockedresolver.MockedImplementations()
	if err != nil {
		return []*gqlpublicapi.Implementation{}, err
	}

	filtered := []*gqlpublicapi.Implementation{}
	for _, i := range implementations {
		for _, revision := range i.Revisions {
			for _, iface := range revision.Spec.Implements {
				if iface.Path == *obj.Metadata.Path && iface.Revision == obj.Revision {
					filtered = append(filtered, i)
				}
			}
		}
	}
	return filtered, nil
}
