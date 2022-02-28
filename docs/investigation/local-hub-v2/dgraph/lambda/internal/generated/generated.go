// @generated - This was created as a part of investigation. We mark it as generate to exclude it from goreportcard to do not have misleading issues.
package generated

import (
	"capactio/lambda/internal/model"
	"capactio/lambda/internal/resolvers"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/schartey/dgraph-lambda-go/api"
)

type Executer struct {
	api.ExecuterInterface
	fieldResolver      resolvers.FieldResolver
	queryResolver      resolvers.QueryResolver
	mutationResolver   resolvers.MutationResolver
	middlewareResolver resolvers.MiddlewareResolver
	webhookResolver    resolvers.WebhookResolver
}

func NewExecuter(resolver *resolvers.Resolver) api.ExecuterInterface {
	return Executer{fieldResolver: resolvers.FieldResolver{Resolver: resolver}, queryResolver: resolvers.QueryResolver{Resolver: resolver}, mutationResolver: resolvers.MutationResolver{Resolver: resolver}, middlewareResolver: resolvers.MiddlewareResolver{Resolver: resolver}, webhookResolver: resolvers.WebhookResolver{Resolver: resolver}}
}

func (e Executer) Resolve(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	if request.Resolver == "$webhook" {
		return nil, e.resolveWebhook(ctx, request)
	} else {
		parentsBytes, underlyingError := request.Parents.MarshalJSON()
		if underlyingError != nil {
			return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
		}

		mc := &api.MiddlewareContext{Ctx: ctx, Request: request}
		if err = e.middleware(mc); err != nil {
			return nil, err
		}
		ctx = mc.Ctx
		request = mc.Request

		if strings.HasPrefix(request.Resolver, "Query.") {
			return e.resolveQuery(ctx, request)
		} else if strings.HasPrefix(request.Resolver, "Mutation.") {
			return e.resolveMutation(ctx, request)
		} else {
			return e.resolveField(ctx, request, parentsBytes)
		}
	}
}

func (e Executer) middleware(mc *api.MiddlewareContext) (err *api.LambdaError) {
	switch mc.Request.Resolver {
	}
	return nil
}

func (e Executer) resolveField(ctx context.Context, request *api.Request, parentsBytes []byte) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
	case "TypeInstanceResourceVersionSpec.value":
		{
			var parents []*model.TypeInstanceResourceVersionSpec
			json.Unmarshal(parentsBytes, &parents)

			result, err := e.fieldResolver.TypeInstanceResourceVersionSpec_value(ctx, parents, request.AuthHeader)
			if err != nil {
				return nil, err
			}

			var underlyingError error
			response, underlyingError = json.Marshal(result)
			if underlyingError != nil {
				return nil, &api.LambdaError{Underlying: underlyingError, Status: http.StatusInternalServerError}
			} else {
				return response, nil
			}
			break
		}
	}

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveQuery(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
	}

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveMutation(ctx context.Context, request *api.Request) (response []byte, err *api.LambdaError) {
	switch request.Resolver {
	}

	return nil, &api.LambdaError{Underlying: errors.New("could not find query resolver"), Status: http.StatusNotFound}
}

func (e Executer) resolveWebhook(ctx context.Context, request *api.Request) (err *api.LambdaError) {
	switch request.Event.TypeName {
	}

	return &api.LambdaError{Underlying: errors.New("could not find webhook resolver"), Status: http.StatusNotFound}
}
