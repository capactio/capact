package interfacegroups

import (
	"context"

	"projectvoltron.dev/voltron/internal/ptr"
	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type InterfaceGroupResolver struct{}

func NewResolver() *InterfaceGroupResolver {
	return &InterfaceGroupResolver{}
}

func (r *InterfaceGroupResolver) InterfaceGroups(ctx context.Context, filter *gqlpublicapi.InterfaceGroupFilter) ([]*gqlpublicapi.InterfaceGroup, error) {
	return []*gqlpublicapi.InterfaceGroup{dummyInterfaceGroup()}, nil
}

func (r *InterfaceGroupResolver) InterfaceGroup(ctx context.Context, path string) (*gqlpublicapi.InterfaceGroup, error) {
	return dummyInterfaceGroup(), nil
}

func dummyInterfaceGroup() *gqlpublicapi.InterfaceGroup {
	return &gqlpublicapi.InterfaceGroup{
		Metadata: &gqlpublicapi.GenericMetadata{
			Name:        "wordpress",
			Prefix:      ptr.String("cap.interface.cms"),
			Path:        ptr.String("cap.interface.cms.wordpress"),
			DisplayName: ptr.String("Wordpress"),
			Description: "Wordpress Application",
			Maintainers: []*gqlpublicapi.Maintainer{
				{
					Name:  ptr.String("Foo Bar"),
					Email: "foo@example.com",
					URL:   ptr.String("https://examples.com/foo/bar"),
				},
			},
		},
		Signature: &gqlpublicapi.Signature{
			Och: "eyJ0eXAiOiJKV1QiLA0KICJhbGciOiJIUzI1NiJ9",
		},
		Interfaces: nil,
	}
}
