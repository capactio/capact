package implementations

import (
	"context"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type ImplementationResolver struct {
}

func NewResolver() *ImplementationResolver {
	return &ImplementationResolver{}
}
func (i *ImplementationResolver) Implementations(ctx context.Context, filter *gqlpublicapi.ImplementationFilter) ([]*gqlpublicapi.Implementation, error) {
	return []*gqlpublicapi.Implementation{dummyImplementation()}, nil
}

func (i ImplementationResolver) Implementation(ctx context.Context, path string) (*gqlpublicapi.Implementation, error) {
	return dummyImplementation(), nil
}

func dummyImplementation() *gqlpublicapi.Implementation {
	return &gqlpublicapi.Implementation{
		Name:   "install",
		Prefix: "cap.implementation.cms.wordpress",
		Path:   "cap.implementation.cms.wordpress.install",
	}
}
