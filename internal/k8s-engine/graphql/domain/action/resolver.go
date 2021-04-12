package action

import (
	"context"

	"github.com/pkg/errors"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/model"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

type actionConverter interface {
	FromGraphQLInput(in graphql.ActionDetailsInput) model.ActionToCreateOrUpdate
	ToGraphQL(in v1alpha1.Action) graphql.Action
	FilterFromGraphQL(in *graphql.ActionFilter) (model.ActionFilter, error)
	AdvancedModeContinueRenderingInputFromGraphQL(in graphql.AdvancedModeContinueRenderingInput) model.AdvancedModeContinueRenderingInput
}

type actionService interface {
	Create(ctx context.Context, item model.ActionToCreateOrUpdate) (v1alpha1.Action, error)
	Update(ctx context.Context, item model.ActionToCreateOrUpdate) (v1alpha1.Action, error)
	GetByName(ctx context.Context, name string) (v1alpha1.Action, error)
	List(ctx context.Context, filter model.ActionFilter) ([]v1alpha1.Action, error)
	DeleteByName(ctx context.Context, name string) error
	RunByName(ctx context.Context, name string) error
	CancelByName(ctx context.Context, name string) error
	ContinueAdvancedRendering(ctx context.Context, actionName string, in model.AdvancedModeContinueRenderingInput) error
}

type Resolver struct {
	svc  actionService
	conv actionConverter
}

func NewResolver(svc actionService, conv actionConverter) *Resolver {
	return &Resolver{
		svc:  svc,
		conv: conv,
	}
}

func (r *Resolver) Action(ctx context.Context, name string) (*graphql.Action, error) {
	item, err := r.svc.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrActionNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "while finding Action by name")
	}

	gqlItem := r.conv.ToGraphQL(item)
	return &gqlItem, nil
}

func (r *Resolver) Actions(ctx context.Context, filter *graphql.ActionFilter) ([]*graphql.Action, error) {
	svcFilter, err := r.conv.FilterFromGraphQL(filter)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Action filter")
	}

	items, err := r.svc.List(ctx, svcFilter)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Actions")
	}

	gqlItems := make([]*graphql.Action, 0, len(items))
	for _, item := range items {
		gqlItem := r.conv.ToGraphQL(item)
		gqlItems = append(gqlItems, &gqlItem)
	}

	return gqlItems, nil
}

func (r *Resolver) CreateAction(ctx context.Context, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	if in == nil {
		return nil, errors.New("input cannot be empty")
	}

	actionToCreate := r.conv.FromGraphQLInput(*in)

	out, err := r.svc.Create(ctx, actionToCreate)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Action")
	}

	gqlItem := r.conv.ToGraphQL(out)
	return &gqlItem, nil
}

func (r *Resolver) RunAction(ctx context.Context, name string) (*graphql.Action, error) {
	err := r.svc.RunByName(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "while running Action")
	}

	return r.findAndConvertToGQL(ctx, name)
}

func (r *Resolver) CancelAction(ctx context.Context, name string) (*graphql.Action, error) {
	err := r.svc.CancelByName(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "while canceling Action")
	}

	return r.findAndConvertToGQL(ctx, name)
}

func (r *Resolver) DeleteAction(ctx context.Context, name string) (*graphql.Action, error) {
	gqlItem, err := r.findAndConvertToGQL(ctx, name)
	if err != nil {
		return nil, err
	}

	err = r.svc.DeleteByName(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "while deleting Action")
	}

	return gqlItem, nil
}

func (r *Resolver) findAndConvertToGQL(ctx context.Context, name string) (*graphql.Action, error) {
	item, err := r.svc.GetByName(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "while finding Action by name")
	}

	gqlItem := r.conv.ToGraphQL(item)
	return &gqlItem, nil
}

func (r *Resolver) UpdateAction(ctx context.Context, in graphql.ActionDetailsInput) (*graphql.Action, error) {
	actionToUpdate := r.conv.FromGraphQLInput(in)

	out, err := r.svc.Update(ctx, actionToUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "while updating Action")
	}

	gqlItem := r.conv.ToGraphQL(out)
	return &gqlItem, nil
}

func (r *Resolver) ContinueAdvancedRendering(ctx context.Context, actionName string, in graphql.AdvancedModeContinueRenderingInput) (*graphql.Action, error) {
	continueRenderingInput := r.conv.AdvancedModeContinueRenderingInputFromGraphQL(in)

	err := r.svc.ContinueAdvancedRendering(ctx, actionName, continueRenderingInput)
	if err != nil {
		return nil, errors.Wrap(err, "while continuing advanced rendering for Action")
	}

	return r.findAndConvertToGQL(ctx, actionName)
}
