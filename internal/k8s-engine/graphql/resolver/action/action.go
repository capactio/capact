package action

import (
	"context"
	"time"

	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
)

type ActionResolver struct{}

func NewResolver() *ActionResolver {
	return &ActionResolver{}
}

func (r *ActionResolver) Action(ctx context.Context, id string) (*graphql.Action, error) {
	return dummyAction(id), nil
}

func (r *ActionResolver) Actions(ctx context.Context, filter []*graphql.ActionFilter) ([]*graphql.Action, error) {
	return []*graphql.Action{dummyAction("0eb7fa4e-49e6-4734-b778-ae6231459c2c"), dummyAction("17ef2646-64bd-415f-8746-8c3576f58ce7")}, nil
}

func (r *ActionResolver) CreateAction(ctx context.Context, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	return dummyAction("fa225753-1007-4bb1-a233-c43de3bb72ed"), nil
}

func (r *ActionResolver) RunAction(ctx context.Context, id string) (*graphql.Action, error) {
	return dummyAction("fa225753-1007-4bb1-a233-c43de3bb72ef"), nil
}

func (r *ActionResolver) CancelAction(ctx context.Context, id string) (*graphql.Action, error) {
	return dummyAction(id), nil
}

func (r *ActionResolver) UpdateAction(ctx context.Context, id string, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	return dummyAction(id), nil
}

func (r *ActionResolver) ContinueAdvancedRendering(ctx context.Context, actionID string, in graphql.AdvancedModeContinueRenderingInput) (*graphql.Action, error) {
	return dummyAction(actionID), nil
}

func (r *ActionResolver) DeleteAction(ctx context.Context, id string) (*graphql.Action, error) {
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
		Path:         "deploy",
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
