package interfaces

import (
	"context"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type InterfaceResolver struct{}

func NewResolver() *InterfaceResolver {
	return &InterfaceResolver{}
}

func (r *InterfaceResolver) Interfaces(ctx context.Context, filter *gqlpublicapi.InterfaceFilter) ([]*gqlpublicapi.Interface, error) {
	return []*gqlpublicapi.Interface{dummyInterface("install"), dummyInterface("upgrade")}, nil
}

func (r *InterfaceResolver) Interface(ctx context.Context, path string) (*gqlpublicapi.Interface, error) {
	return dummyInterface("install"), nil
}

func (r *InterfaceResolver) Revision(ctx context.Context, obj *gqlpublicapi.Interface, revision string) (*gqlpublicapi.InterfaceRevision, error) {
	return &gqlpublicapi.InterfaceRevision{}, nil
}

func dummyInterface(name string) *gqlpublicapi.Interface {
	return &gqlpublicapi.Interface{
		Name:   name,
		Prefix: "cap.interface.cms.wordpress",
		Path:   "cap.interface.cms.wordpress." + name,
	}
}
