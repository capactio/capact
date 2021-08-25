package action

import (
	"context"

	"capact.io/capact/internal/k8s-engine/graphql/model"
	"capact.io/capact/internal/multierror"
	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"github.com/pkg/errors"
)

type actionConverter interface {
	FromGraphQLInput(in graphql.ActionDetailsInput) (model.ActionToCreateOrUpdate, error)
	ToGraphQL(in v1alpha1.Action) (graphql.Action, error)
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

// Resolver provides functionality to handle Action GraphQL operation such as queries and mutations.
type Resolver struct {
	svc  actionService
	conv actionConverter
}

// NewResolver returns a new Resolver instance.
func NewResolver(svc actionService, conv actionConverter) *Resolver {
	return &Resolver{
		svc:  svc,
		conv: conv,
	}
}

// Action returns Action with a given name.
func (r *Resolver) Action(ctx context.Context, name string) (*graphql.Action, error) {
	item, err := r.svc.GetByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrActionNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "while finding Action by name")
	}

	gqlItem, err := r.conv.ToGraphQL(item)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Action to GraphQL")
	}
	return &gqlItem, nil
}

// Actions returns all Actions which meet a given filter criteria.
func (r *Resolver) Actions(ctx context.Context, filter *graphql.ActionFilter) ([]*graphql.Action, error) {
	svcFilter, err := r.conv.FilterFromGraphQL(filter)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Action filter")
	}

	items, err := r.svc.List(ctx, svcFilter)
	if err != nil {
		return nil, errors.Wrap(err, "while listing Actions")
	}

	var actErrors error

	gqlItems := make([]*graphql.Action, 0, len(items))
	for _, item := range items {
		gqlItem, err := r.conv.ToGraphQL(item)

		if err != nil {
			actErrors = multierror.Append(actErrors, err)
			continue
		}

		gqlItems = append(gqlItems, &gqlItem)
	}

	return gqlItems, actErrors
}

// CreateAction creates Action on cluster side.
func (r *Resolver) CreateAction(ctx context.Context, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	if in == nil {
		return nil, errors.New("input cannot be empty")
	}

	actionToCreate, err := r.conv.FromGraphQLInput(*in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting GraphQL input to Action")
	}

	out, err := r.svc.Create(ctx, actionToCreate)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Action")
	}

	gqlItem, err := r.conv.ToGraphQL(out)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Action to GraphQL")
	}

	return &gqlItem, nil
}

// RunAction executes a given Action.
func (r *Resolver) RunAction(ctx context.Context, name string) (*graphql.Action, error) {
	err := r.svc.RunByName(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "while running Action")
	}

	return r.findAndConvertToGQL(ctx, name)
}

// CancelAction cancels a given action.
func (r *Resolver) CancelAction(ctx context.Context, name string) (*graphql.Action, error) {
	err := r.svc.CancelByName(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "while canceling Action")
	}

	return r.findAndConvertToGQL(ctx, name)
}

// DeleteAction deletes a given Action.
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

	gqlItem, err := r.conv.ToGraphQL(item)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Action to GraphQL")
	}

	return &gqlItem, nil
}

// UpdateAction updates a given Action.
func (r *Resolver) UpdateAction(ctx context.Context, in graphql.ActionDetailsInput) (*graphql.Action, error) {
	actionToUpdate, err := r.conv.FromGraphQLInput(in)
	if err != nil {
		return nil, errors.Wrap(err, "while converting GraphQL input to Action")
	}

	out, err := r.svc.Update(ctx, actionToUpdate)
	if err != nil {
		return nil, errors.Wrap(err, "while updating Action")
	}

	gqlItem, err := r.conv.ToGraphQL(out)
	if err != nil {
		return nil, errors.Wrap(err, "while converting Action to GraphQL")
	}

	return &gqlItem, nil
}

// ContinueAdvancedRendering continues advanced rendering for a given Action. Input parameters are validate before continuation.
func (r *Resolver) ContinueAdvancedRendering(ctx context.Context, actionName string, in graphql.AdvancedModeContinueRenderingInput) (*graphql.Action, error) {
	continueRenderingInput := r.conv.AdvancedModeContinueRenderingInputFromGraphQL(in)

	err := r.svc.ContinueAdvancedRendering(ctx, actionName, continueRenderingInput)
	if err != nil {
		return nil, errors.Wrap(err, "while continuing advanced rendering for Action")
	}

	return r.findAndConvertToGQL(ctx, actionName)
}
