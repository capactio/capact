package graphql

import (
	"context"
	"time"

	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
)

var _ graphql.ResolverRoot = &RootResolver{}

type RootResolver struct {
	mutationResolver
	queryResolver
}

func NewRootResolver() *RootResolver {
	return &RootResolver{}
}

func (r RootResolver) Mutation() graphql.MutationResolver {
	return r.mutationResolver
}

func (r RootResolver) Query() graphql.QueryResolver {
	return r.queryResolver
}

type queryResolver struct {
	graphql.QueryResolver
}

func (q queryResolver) Action(ctx context.Context, id string) (*graphql.Action, error) {
	return dummyAction(id), nil
}

func (q queryResolver) Actions(ctx context.Context, filter []*graphql.ActionFilter) ([]*graphql.Action, error) {
	return []*graphql.Action{dummyAction("0eb7fa4e-49e6-4734-b778-ae6231459c2c"), dummyAction("17ef2646-64bd-415f-8746-8c3576f58ce7")}, nil
}

type mutationResolver struct {
	graphql.MutationResolver
}

func (m mutationResolver) CreateAction(ctx context.Context, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	return dummyAction("fa225753-1007-4bb1-a233-c43de3bb72ed"), nil
}

func (m mutationResolver) RunAction(ctx context.Context, id string) (*graphql.Action, error) {
	return dummyAction("fa225753-1007-4bb1-a233-c43de3bb72ef"), nil
}

func (m mutationResolver) CancelAction(ctx context.Context, id string) (*graphql.Action, error) {
	return dummyAction(id), nil
}

func (m mutationResolver) UpdateAction(ctx context.Context, id string, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	return dummyAction(id), nil
}

func (m mutationResolver) ContinueAdvancedRendering(ctx context.Context, actionID string, in graphql.AdvancedModeContinueRenderingInput) (*graphql.Action, error) {
	return dummyAction(actionID), nil
}

func (m mutationResolver) DeleteAction(ctx context.Context, id string) (*graphql.Action, error) {
	return dummyAction(id), nil
}

func dummyAction(id string) *graphql.Action {
	return &graphql.Action{
		ID:        id,
		CreatedAt: graphql.Timestamp(time.Now()),
		Input: &graphql.ActionInput{
			Parameters: `{"param": "val1"}`,
		},
		Output: &graphql.ActionOutput{
			Artifacts: []*graphql.OutputArtifact{
				{
					Name:           "artifact01",
					TypeInstanceID: "74f911cc-bbc9-47f5-8c21-1045550c31ef",
					TypePath:       "cap.type.db.mysql.config",
				},
			},
		},
		Action:         "deploy",
		RenderedAction: nil,
		RenderingAdvancedMode: &graphql.ActionRenderingAdvancedMode{
			Enabled: false,
		},
		CreatedBy: &graphql.UserInfo{
			Username: "mszostok",
		},
		RunBy: &graphql.UserInfo{
			Username: "mszostok",
		},
		Status: &graphql.ActionStatus{
			Condition: graphql.ActionStatusConditionRunning,
			Timestamp: graphql.Timestamp(time.Now()),
			Runner: &graphql.RunnerStatus{
				Type: "cap.type.runner.argo",
				Status: struct {
					ArgoWorkflowRef string
				}{
					ArgoWorkflowRef: "default/WorkflowRun01",
				},
			},
		},
	}
}
