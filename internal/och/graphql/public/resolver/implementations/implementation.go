package implementations

import (
	"context"
	"fmt"

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

func (r *ImplementationResolver) Revision(ctx context.Context, obj *gqlpublicapi.Implementation, revision string) (*gqlpublicapi.ImplementationRevision, error) {
	return &gqlpublicapi.ImplementationRevision{}, fmt.Errorf("No Implementation with revision %s", revision)
}

func dummyImplementation() *gqlpublicapi.Implementation {
	return &gqlpublicapi.Implementation{
		Name:   "install",
		Prefix: "cap.implementation.cms.wordpress",
		Path:   "cap.implementation.cms.wordpress.install",
	}
}
