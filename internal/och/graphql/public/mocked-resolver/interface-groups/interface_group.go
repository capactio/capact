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
	groups, err := mockedresolver.MockedInterfaceGroups()
	if err != nil {
		return []*gqlpublicapi.InterfaceGroup{}, err
	}
	return groups, nil
}

func (r *InterfaceGroupResolver) InterfaceGroup(ctx context.Context, path string) (*gqlpublicapi.InterfaceGroup, error) {
	groups, err := mockedresolver.MockedInterfaceGroups()
	if err != nil {
		return &gqlpublicapi.InterfaceGroup{}, err
	}
	for _, group := range groups {
		if group.Metadata != nil && group.Metadata.Path == path {
			return group, nil
		}
	}
	return nil, nil
}

func (r *InterfaceGroupInterfacesResolver) Interfaces(ctx context.Context, obj *gqlpublicapi.InterfaceGroup, filter *gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error) {
	if obj == nil || obj.Metadata == nil {
		return []*gqlpublicapi.Interface{}, nil
	}
	ifaces, err := mockedresolver.MockedInterfaces()
	if err != nil {
		return []*gqlpublicapi.Interface{}, err
	}

	filtered := []*gqlpublicapi.Interface{}
	for _, iface := range ifaces {
		if obj.Metadata.Path == iface.Prefix {
			filtered = append(filtered, iface)
		}
	}
	return filtered, nil
}
