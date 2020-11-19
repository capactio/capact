package interfacegroups

import (
	"context"

	mockedresolver "projectvoltron.dev/voltron/internal/och/graphql/public/mocked-resolver"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type InterfaceGroupResolver struct{}

func NewResolver() *InterfaceGroupResolver {
	return &InterfaceGroupResolver{}
}

type InterfaceGroupInterfacesResolver struct {
	*InterfaceGroupResolver
}

func NewInterfacesResolver() *InterfaceGroupInterfacesResolver {
	return &InterfaceGroupInterfacesResolver{}
}

func (r *InterfaceGroupResolver) InterfaceGroups(ctx context.Context, filter *gqlpublicapi.InterfaceGroupFilter) ([]*gqlpublicapi.InterfaceGroup, error) {
	ig, err := mockedresolver.MockedInterfaceGroup()
	if err != nil {
		return []*gqlpublicapi.InterfaceGroup{}, err
	}
	return []*gqlpublicapi.InterfaceGroup{ig}, nil
}

func (r *InterfaceGroupResolver) InterfaceGroup(ctx context.Context, path string) (*gqlpublicapi.InterfaceGroup, error) {
	ig, err := mockedresolver.MockedInterfaceGroup()
	if err != nil {
		return &gqlpublicapi.InterfaceGroup{}, err
	}
	if *ig.Metadata.Path == path {
		return ig, nil
	}
	return &gqlpublicapi.InterfaceGroup{}, nil
}

func (r *InterfaceGroupInterfacesResolver) Interfaces(ctx context.Context, obj *gqlpublicapi.InterfaceGroup, filter *gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error) {
	i, err := mockedresolver.MockedInterface()
	if err != nil {
		return []*gqlpublicapi.Interface{}, err
	}
	if obj != nil && *obj.Metadata.Path == i.Prefix {
		return []*gqlpublicapi.Interface{i}, nil
	}
	return []*gqlpublicapi.Interface{}, nil
}
