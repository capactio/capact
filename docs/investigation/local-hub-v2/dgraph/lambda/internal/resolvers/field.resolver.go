// @generated - This was created as a part of investigation. We mark it as generate to exlude it from goreportcard to do not have missleading issues.:golint
package resolvers

import (
	"capactio/lambda/internal/model"
	"context"
	"fmt"

	"github.com/sanity-io/litter"
	"github.com/schartey/dgraph-lambda-go/api"
)

type FieldResolverInterface interface {
	TypeInstanceResourceVersionSpec_value(ctx context.Context, parents []*model.TypeInstanceResourceVersionSpec, authHeader api.AuthHeader) ([]string, *api.LambdaError)
}

type FieldResolver struct {
	*Resolver
}

func (f *FieldResolver) TypeInstanceResourceVersionSpec_value(ctx context.Context, parents []*model.TypeInstanceResourceVersionSpec, authHeader api.AuthHeader) ([]string, *api.LambdaError) {
	litter.Dump(parents)
	var out []string
	for idx := range parents {
		out = append(out, fmt.Sprintf("value for %d", idx))
	}
	return out, nil
}
