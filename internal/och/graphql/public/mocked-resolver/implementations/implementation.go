package implementations

import (
	"context"

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
	implementations, err := mockedresolver.MockedImplementations()
	if err != nil {
		return []*gqlpublicapi.Implementation{}, err
	}
	return implementations, nil
}

func (i ImplementationResolver) Implementation(ctx context.Context, path string) (*gqlpublicapi.Implementation, error) {
	implementations, err := mockedresolver.MockedImplementations()
	if err != nil {
		return nil, err
	}
	for _, i := range implementations {
		if i.Path == path {
			return i, nil
		}
	}
	return nil, nil
}

func (i *ImplementationResolver) Revision(ctx context.Context, obj *gqlpublicapi.Implementation, revision string) (*gqlpublicapi.ImplementationRevision, error) {
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

func (i *ImplementationRevisionResolver) Interfaces(ctx context.Context, obj *gqlpublicapi.ImplementationRevision) ([]*gqlpublicapi.InterfaceRevision, error) {
	if obj == nil {
		return []*gqlpublicapi.InterfaceRevision{}, nil
	}
	ifaces, err := mockedresolver.MockedInterfaces()
	if err != nil {
		return []*gqlpublicapi.InterfaceRevision{}, err
	}
	filtered := []*gqlpublicapi.InterfaceRevision{}
	for _, iface := range ifaces {
		for _, revision := range iface.Revisions {
			for _, impl := range obj.Spec.Implements {
				if iface.Path == impl.Path {
					filtered = append(filtered, revision)
				}
			}
		}
	}
	return filtered, nil
}
