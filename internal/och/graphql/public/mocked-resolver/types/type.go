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
	return []*gqlpublicapi.Type{}, nil
}

func (r *TypeResolver) Type(ctx context.Context, path string) (*gqlpublicapi.Type, error) {
	return &gqlpublicapi.Type{}, nil
}

func (r *TypeResolver) Revision(ctx context.Context, obj *gqlpublicapi.Type, revision string) (*gqlpublicapi.TypeRevision, error) {
	return &gqlpublicapi.TypeRevision{}, nil
}

func getType() *gqlpublicapi.Type {
	return &gqlpublicapi.Type{}
}
