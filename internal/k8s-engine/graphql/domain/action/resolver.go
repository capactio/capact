package action

import (
	"context"
	"time"

	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/model"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"

	"github.com/pkg/errors"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

type actionConverter interface {
	FromGraphQLInput(in graphql.ActionDetailsInput, namespace string) model.ActionToCreateOrUpdate
	ToGraphQL(in v1alpha1.Action) graphql.Action
}

type actionService interface {
	Create(ctx context.Context, item model.ActionToCreateOrUpdate) error
	FindByName(ctx context.Context, name string) (v1alpha1.Action, error)
	List(ctx context.Context) ([]v1alpha1.Action, error)
	DeleteByName(ctx context.Context, name string) error
	RunByName(ctx context.Context, name string) error
	CancelByName(ctx context.Context, name string) error
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
	item, err := r.svc.FindByName(ctx, name)
	if err != nil {
		if errors.Is(err, ErrActionNotFound) {
			return nil, nil
		}

		return nil, errors.Wrap(err, "while finding Action by name")
	}

	gqlItem := r.conv.ToGraphQL(item)
	return &gqlItem, nil
}

// TODO: Implement filter as a part of SV-60
func (r *Resolver) Actions(ctx context.Context, filter []*graphql.ActionFilter) ([]*graphql.Action, error) {
	items, err := r.svc.List(ctx)
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

	ns, err := namespace.FromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while reading namespace from context")
	}

	actionToCreate := r.conv.FromGraphQLInput(*in, ns)

	err = r.svc.Create(ctx, actionToCreate)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Action")
	}

	return r.findAndConvertToGQL(ctx, in.Name)
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
		return nil, errors.Wrap(err, "while cancelling Action")
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
	item, err := r.svc.FindByName(ctx, name)
	if err != nil {
		return nil, errors.Wrap(err, "while finding Action by name")
	}

	gqlItem := r.conv.ToGraphQL(item)
	return &gqlItem, nil
}

// TODO: To implement as a part of SV-61

func (r *Resolver) UpdateAction(ctx context.Context, in graphql.ActionDetailsInput) (*graphql.Action, error) {
	return dummyAction(in.Name), nil
}

func (r *Resolver) ContinueAdvancedRendering(ctx context.Context, actionName string, in graphql.AdvancedModeContinueRenderingInput) (*graphql.Action, error) {
	return dummyAction(actionName), nil
}

func dummyAction(name string) *graphql.Action {
	return &graphql.Action{
		Name:           name,
		CreatedAt:      graphql.Timestamp(time.Now()),
		Path:           "deploy",
		RenderedAction: nil,
		RenderingAdvancedMode: &graphql.ActionRenderingAdvancedMode{
			Enabled: false,
		},
		Status: &graphql.ActionStatus{
			Condition: graphql.ActionStatusConditionRunning,
			Timestamp: graphql.Timestamp(time.Now()),
			CreatedBy: &graphql.UserInfo{
				Username: "mszostok",
			},
			RunBy: &graphql.UserInfo{
				Username: "mszostok",
			},
			Runner: &graphql.RunnerStatus{
				Status: struct {
					ArgoWorkflowRef string
				}{
					ArgoWorkflowRef: "default/WorkflowRun01",
				},
			},
		},
	}
}
