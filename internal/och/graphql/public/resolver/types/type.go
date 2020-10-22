package types

import (
	"context"

	gqlpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
)

type TypeResolver struct{}

func NewResolver() *TypeResolver {
	return &TypeResolver{}
}

func (r *TypeResolver) Types(ctx context.Context, filter *gqlpublicapi.TypeFilter) ([]*gqlpublicapi.Type, error) {
	return []*gqlpublicapi.Type{dummyType("config"), dummyType("user")}, nil
}

func (r *TypeResolver) Type(ctx context.Context, path string) (*gqlpublicapi.Type, error) {
	return dummyType("config"), nil
}

func dummyType(name string) *gqlpublicapi.Type {
	return &gqlpublicapi.Type{
		Name:   name,
		Prefix: "cap.type.database.mysql",
		Path:   "cap.type.database.mysql." + name,
	}
}
