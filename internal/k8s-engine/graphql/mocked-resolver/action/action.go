package action

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"math/rand"
	"path"
	"time"

	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
)

var stateChangeWaitTime = 10 * time.Second

type isConditionPossible func(a *graphql.Action) bool

func canBeRun(a *graphql.Action) bool {
	return a.Run
}

type conditionsGraphNode struct {
	condition      graphql.ActionStatusCondition
	conditionValid isConditionPossible
}

var conditionsGraph = map[graphql.ActionStatusCondition][]conditionsGraphNode{
	graphql.ActionStatusConditionInitial: {
		{condition: graphql.ActionStatusConditionBeingRendered},
	},
	graphql.ActionStatusConditionBeingRendered: {
		{condition: graphql.ActionStatusConditionReadyToRun},
	},
	graphql.ActionStatusConditionAdvancedModeRenderingIteration: {
		{condition: graphql.ActionStatusConditionReadyToRun},
		{condition: graphql.ActionStatusConditionAdvancedModeRenderingIteration},
	},
	graphql.ActionStatusConditionReadyToRun: {
		{
			condition:      graphql.ActionStatusConditionRunning,
			conditionValid: canBeRun,
		},
	},
	graphql.ActionStatusConditionRunning: {
		{condition: graphql.ActionStatusConditionSucceeded},
		{condition: graphql.ActionStatusConditionFailed},
	},

	graphql.ActionStatusConditionBeingCancelled: {
		{condition: graphql.ActionStatusConditionCancelled},
	},
	graphql.ActionStatusConditionCancelled: {},
	graphql.ActionStatusConditionSucceeded: {},
	graphql.ActionStatusConditionFailed:    {},
}

var conditionMessages = map[graphql.ActionStatusCondition]string{
	graphql.ActionStatusConditionInitial:                        "Action initialized",
	graphql.ActionStatusConditionBeingRendered:                  "Action is being rendered",
	graphql.ActionStatusConditionAdvancedModeRenderingIteration: "Action in advanced rendering mode",
	graphql.ActionStatusConditionReadyToRun:                     "Action is ready to run",
	graphql.ActionStatusConditionRunning:                        "Action is running",
	graphql.ActionStatusConditionBeingCancelled:                 "Action is being canceled",
	graphql.ActionStatusConditionCancelled:                      "Action canceled",
	graphql.ActionStatusConditionSucceeded:                      "Action succeeded",
	graphql.ActionStatusConditionFailed:                         "Action failed",
}

// proceedAction simulates engine. It waits `stateChangeWaitTime`
// and randomly moves to the next possibly state
func proceedAction(action *graphql.Action) {
	rand.Seed(time.Now().Unix())
	for {
		time.Sleep(stateChangeWaitTime)

		nextConditions := conditionsGraph[action.Status.Condition]
		if len(nextConditions) == 0 {
			break
		}

		possibleConditions := []graphql.ActionStatusCondition{}
		for _, conditionNode := range nextConditions {
			if conditionNode.conditionValid != nil && !conditionNode.conditionValid(action) {
				continue
			}
			possibleConditions = append(possibleConditions, conditionNode.condition)
		}
		if len(possibleConditions) == 0 {
			continue
		}

		// #nosec G404
		newCondition := possibleConditions[rand.Intn(len(possibleConditions))]
		message := conditionMessages[newCondition]

		action.Status.Condition = newCondition
		action.Status.Message = &message
	}
}

var mockUser = &graphql.UserInfo{
	Username: "user",
	Groups:   []string{"mocks"},
}

type ActionResolver struct {
	actions []*graphql.Action
}

func NewResolver() *ActionResolver {
	return &ActionResolver{}
}

func (a *ActionResolver) init() error {
	if a.actions != nil {
		return nil
	}
	actions, err := mockedActions()
	if err != nil {
		return err
	}
	a.actions = append(a.actions, actions...)
	return nil
}

func (a *ActionResolver) getAction(name string) (int, *graphql.Action, error) {
	err := a.init()
	if err != nil {
		return -1, nil, err
	}
	for i, action := range a.actions {
		if action.Name == name {
			return i, action, nil
		}
	}
	return -1, nil, fmt.Errorf("Failed to get action with id %s", name)
}

func (a *ActionResolver) Action(ctx context.Context, name string) (*graphql.Action, error) {
	_, action, err := a.getAction(name)
	if err != nil {
		return nil, err
	}
	return action, nil
}

func (a *ActionResolver) Actions(ctx context.Context, filter []*graphql.ActionFilter) ([]*graphql.Action, error) {
	err := a.init()
	if err != nil {
		return []*graphql.Action{}, err
	}
	return a.actions, nil
}

func (a *ActionResolver) CreateAction(ctx context.Context, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	message := conditionMessages[graphql.ActionStatusConditionInitial]
	var dryRun bool
	if in.DryRun != nil {
		dryRun = *in.DryRun
	}

	var revision string
	if in.ActionRef.Revision != nil {
		revision = *in.ActionRef.Revision
	} else {
		revision = "latest"
	}

	newAction := &graphql.Action{
		Name: in.Name,
		ActionRef: &graphql.ManifestReference{
			Path:     in.ActionRef.Path,
			Revision: revision,
		},
		CreatedAt: graphql.Timestamp(time.Now()),
		Input:     &graphql.ActionInput{},
		DryRun:    dryRun,

		Status: &graphql.ActionStatus{
			CreatedBy: mockUser,
			Condition: graphql.ActionStatusConditionInitial,
			Message:   &message,
			Timestamp: graphql.Timestamp(time.Now()),
			Runner:    &graphql.RunnerStatus{},
		},
	}
	updateAction(newAction, in)

	if in.AdvancedRendering != nil {
		newAction.RenderingAdvancedMode = &graphql.ActionRenderingAdvancedMode{Enabled: *in.AdvancedRendering}
	}
	if in.RenderedActionOverride != nil {
		newAction.RenderedActionOverride = *in.RenderedActionOverride
	}
	a.actions = append(a.actions, newAction)

	go proceedAction(newAction)

	return newAction, nil
}

func (a *ActionResolver) RunAction(ctx context.Context, name string) (*graphql.Action, error) {
	_, action, err := a.getAction(name)
	if err != nil {
		return nil, err
	}

	if action.Status.Condition != graphql.ActionStatusConditionReadyToRun &&
		action.Status.Condition != graphql.ActionStatusConditionInitial &&
		action.Status.Condition != graphql.ActionStatusConditionAdvancedModeRenderingIteration &&
		action.Status.Condition != graphql.ActionStatusConditionBeingRendered {
		return action, fmt.Errorf("action is not ready to be run, current condition: %s", action.Status.Condition)
	}
	action.Run = true

	action.Status.RunBy = mockUser
	action.Status.Timestamp = graphql.Timestamp(time.Now())

	return action, nil
}

func (a *ActionResolver) CancelAction(ctx context.Context, id string) (*graphql.Action, error) {
	_, action, err := a.getAction(id)
	if err != nil {
		return nil, err
	}

	if action.Status.Condition != graphql.ActionStatusConditionRunning {
		return action, fmt.Errorf("Action which is not running cannot be canceled")
	}
	action.Cancel = true

	message := conditionMessages[graphql.ActionStatusConditionBeingCancelled]

	action.Status.CancelledBy = mockUser
	action.Status.Condition = graphql.ActionStatusConditionBeingCancelled
	action.Status.Message = &message
	action.Status.Timestamp = graphql.Timestamp(time.Now())

	return action, nil
}

func (a *ActionResolver) UpdateAction(ctx context.Context, in graphql.ActionDetailsInput) (*graphql.Action, error) {
	_, action, err := a.getAction(in.Name)
	if err != nil {
		return nil, err
	}
	updateAction(action, &in)
	return action, nil
}

func (a *ActionResolver) ContinueAdvancedRendering(ctx context.Context, id string, in graphql.AdvancedModeContinueRenderingInput) (*graphql.Action, error) {
	_, action, err := a.getAction(id)
	if err != nil {
		return nil, err
	}
	if action.RenderingAdvancedMode == nil || !action.RenderingAdvancedMode.Enabled {
		return action, errors.New("Rendering in advanced mode is disabled")
	}
	typeInstances := []*graphql.InputTypeInstanceDetails{}
	for _, typeInstance := range in.TypeInstances {
		typeInstances = append(typeInstances,
			&graphql.InputTypeInstanceDetails{
				ID: typeInstance.ID,
			})
	}
	action.RenderingAdvancedMode.TypeInstancesForRenderingIteration = typeInstances
	message := conditionMessages[graphql.ActionStatusConditionBeingRendered]

	action.Status.Condition = graphql.ActionStatusConditionBeingRendered
	action.Status.Message = &message
	action.Status.Timestamp = graphql.Timestamp(time.Now())

	go proceedAction(action)

	return action, nil
}

func (a *ActionResolver) DeleteAction(ctx context.Context, name string) (*graphql.Action, error) {
	index, action, err := a.getAction(name)
	if err != nil {
		return nil, err
	}

	if index == -1 {
		return nil, nil
	}
	a.actions[index] = a.actions[len(a.actions)-1]
	a.actions[len(a.actions)-1] = nil
	a.actions = a.actions[:len(a.actions)-1]
	return action, nil
}

func mockedActions() ([]*graphql.Action, error) {
	mocksPath := "./mock/engine"
	buff, err := ioutil.ReadFile(path.Join(mocksPath, "actions.json"))
	if err != nil {
		return nil, err
	}

	types := []*graphql.Action{}
	err = json.Unmarshal(buff, &types)
	if err != nil {
		return nil, err
	}
	return types, nil
}

func updateAction(action *graphql.Action, in *graphql.ActionDetailsInput) {
	if in == nil {
		return
	}

	typeInstances := []*graphql.InputTypeInstanceDetails{}
	var parameters interface{}
	if in.Input != nil {
		if action.Input == nil {
			action.Input = &graphql.ActionInput{}
		}

		for _, t := range in.Input.TypeInstances {
			typeInstance := &graphql.InputTypeInstanceDetails{
				ID: t.ID,
			}
			typeInstances = append(typeInstances, typeInstance)
		}
		parameters = in.Input.Parameters
	}
	action.Input.Parameters = parameters
	action.Input.TypeInstances = typeInstances
	if in.DryRun != nil {
		action.DryRun = *in.DryRun
	}

	if in.AdvancedRendering != nil {
		action.RenderingAdvancedMode = &graphql.ActionRenderingAdvancedMode{Enabled: *in.AdvancedRendering}
	}
	if in.RenderedActionOverride != nil {
		action.RenderedActionOverride = *in.RenderedActionOverride
	}
}
