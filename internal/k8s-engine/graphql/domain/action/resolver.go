package action

import (
	"context"
	"time"

	"github.com/google/uuid"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/model"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"

	"github.com/pkg/errors"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"

	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
)

type actionConverter interface {
	FromGraphQLInput(in graphql.ActionDetailsInput, name, namespace string) model.ActionToCreateOrUpdate
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

func (r *Resolver) Action(ctx context.Context, id string) (*graphql.Action, error) {
	item, err := r.svc.FindByName(ctx, id)
	if err != nil {
		if errors.Cause(err) == ErrActionNotFound {
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

	var gqlItems []*graphql.Action
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

	ns, err := namespace.ReadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "while reading namespace from context")
	}

	id := uuid.New().String()
	actionToCreate := r.conv.FromGraphQLInput(*in, id, ns)

	err = r.svc.Create(ctx, actionToCreate)
	if err != nil {
		return nil, errors.Wrap(err, "while creating Action")
	}

	return r.findAndConvertToGQL(ctx, id)
}

func (r *Resolver) RunAction(ctx context.Context, id string) (*graphql.Action, error) {
	err := r.svc.RunByName(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "while running Action")
	}

	return r.findAndConvertToGQL(ctx, id)
}

func (r *Resolver) CancelAction(ctx context.Context, id string) (*graphql.Action, error) {
	err := r.svc.CancelByName(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "while cancelling Action")
	}

	return r.findAndConvertToGQL(ctx, id)
}

func (r *Resolver) DeleteAction(ctx context.Context, id string) (*graphql.Action, error) {
	gqlItem, err := r.findAndConvertToGQL(ctx, id)
	if err != nil {
		return nil, err
	}

	err = r.svc.DeleteByName(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "while deleting Action")
	}

	return gqlItem, nil
}

func (r *Resolver) findAndConvertToGQL(ctx context.Context, id string) (*graphql.Action, error) {
	item, err := r.svc.FindByName(ctx, id)
	if err != nil {
		return nil, errors.Wrap(err, "while finding Action by name")
	}

	gqlItem := r.conv.ToGraphQL(item)
	return &gqlItem, nil
}

// TODO: To implement as a part of SV-61

func (r *Resolver) UpdateAction(ctx context.Context, id string, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	return dummyAction(id), nil
}

func (r *Resolver) ContinueAdvancedRendering(ctx context.Context, actionID string, in graphql.AdvancedModeContinueRenderingInput) (*graphql.Action, error) {
	return dummyAction(actionID), nil
}

func dummyAction(id string) *graphql.Action {
	return &graphql.Action{
		ID:             id,
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
				Interface: "cap.type.runner.argo",
				Status: struct {
					ArgoWorkflowRef string
				}{
					ArgoWorkflowRef: "default/WorkflowRun01",
				},
			},
		},
	}
}
