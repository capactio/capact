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

type isPhasePossible func(a *graphql.Action) bool

func canBeRun(a *graphql.Action) bool {
	return a.Run
}

type conditionsGraphNode struct {
	condition      graphql.ActionStatusPhase
	conditionValid isPhasePossible
}

var conditionsGraph = map[graphql.ActionStatusPhase][]conditionsGraphNode{
	graphql.ActionStatusPhaseInitial: {
		{condition: graphql.ActionStatusPhaseBeingRendered},
	},
	graphql.ActionStatusPhaseBeingRendered: {
		{condition: graphql.ActionStatusPhaseReadyToRun},
	},
	graphql.ActionStatusPhaseAdvancedModeRenderingIteration: {
		{condition: graphql.ActionStatusPhaseReadyToRun},
		{condition: graphql.ActionStatusPhaseAdvancedModeRenderingIteration},
	},
	graphql.ActionStatusPhaseReadyToRun: {
		{
			condition:      graphql.ActionStatusPhaseRunning,
			conditionValid: canBeRun,
		},
	},
	graphql.ActionStatusPhaseRunning: {
		{condition: graphql.ActionStatusPhaseSucceeded},
		{condition: graphql.ActionStatusPhaseFailed},
	},

	graphql.ActionStatusPhaseBeingCanceled: {
		{condition: graphql.ActionStatusPhaseCanceled},
	},
	graphql.ActionStatusPhaseCanceled:  {},
	graphql.ActionStatusPhaseSucceeded: {},
	graphql.ActionStatusPhaseFailed:    {},
}

var conditionMessages = map[graphql.ActionStatusPhase]string{
	graphql.ActionStatusPhaseInitial:                        "Action initialized",
	graphql.ActionStatusPhaseBeingRendered:                  "Action is being rendered",
	graphql.ActionStatusPhaseAdvancedModeRenderingIteration: "Action in advanced rendering mode",
	graphql.ActionStatusPhaseReadyToRun:                     "Action is ready to run",
	graphql.ActionStatusPhaseRunning:                        "Action is running",
	graphql.ActionStatusPhaseBeingCanceled:                  "Action is being canceled",
	graphql.ActionStatusPhaseCanceled:                       "Action canceled",
	graphql.ActionStatusPhaseSucceeded:                      "Action succeeded",
	graphql.ActionStatusPhaseFailed:                         "Action failed",
}

// proceedAction simulates engine. It waits `stateChangeWaitTime`
// and randomly moves to the next possibly state
func proceedAction(action *graphql.Action) {
	rand.Seed(time.Now().Unix())
	for {
		time.Sleep(stateChangeWaitTime)

		nextPhases := conditionsGraph[action.Status.Phase]
		if len(nextPhases) == 0 {
			break
		}

		possiblePhases := []graphql.ActionStatusPhase{}
		for _, conditionNode := range nextPhases {
			if conditionNode.conditionValid != nil && !conditionNode.conditionValid(action) {
				continue
			}
			possiblePhases = append(possiblePhases, conditionNode.condition)
		}
		if len(possiblePhases) == 0 {
			continue
		}

		// #nosec G404
		newPhase := possiblePhases[rand.Intn(len(possiblePhases))]
		message := conditionMessages[newPhase]

		action.Status.Phase = newPhase
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

func (a *ActionResolver) Actions(ctx context.Context, filter *graphql.ActionFilter) ([]*graphql.Action, error) {
	err := a.init()
	if err != nil {
		return []*graphql.Action{}, err
	}
	return a.actions, nil
}

func (a *ActionResolver) CreateAction(ctx context.Context, in *graphql.ActionDetailsInput) (*graphql.Action, error) {
	message := conditionMessages[graphql.ActionStatusPhaseInitial]
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
		CreatedAt: graphql.Timestamp{Time: time.Now()},
		Input:     &graphql.ActionInput{},
		DryRun:    dryRun,

		Status: &graphql.ActionStatus{
			CreatedBy: mockUser,
			Phase:     graphql.ActionStatusPhaseInitial,
			Message:   &message,
			Timestamp: graphql.Timestamp{Time: time.Now()},
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

	if action.Status.Phase != graphql.ActionStatusPhaseReadyToRun &&
		action.Status.Phase != graphql.ActionStatusPhaseInitial &&
		action.Status.Phase != graphql.ActionStatusPhaseAdvancedModeRenderingIteration &&
		action.Status.Phase != graphql.ActionStatusPhaseBeingRendered {
		return action, fmt.Errorf("action is not ready to be run, current condition: %s", action.Status.Phase)
	}
	action.Run = true

	action.Status.RunBy = mockUser
	action.Status.Timestamp = graphql.Timestamp{Time: time.Now()}

	return action, nil
}

func (a *ActionResolver) CancelAction(ctx context.Context, id string) (*graphql.Action, error) {
	_, action, err := a.getAction(id)
	if err != nil {
		return nil, err
	}

	if action.Status.Phase != graphql.ActionStatusPhaseRunning {
		return action, fmt.Errorf("Action which is not running cannot be canceled")
	}
	action.Cancel = true

	message := conditionMessages[graphql.ActionStatusPhaseBeingCanceled]

	action.Status.CanceledBy = mockUser
	action.Status.Phase = graphql.ActionStatusPhaseBeingCanceled
	action.Status.Message = &message
	action.Status.Timestamp = graphql.Timestamp{Time: time.Now()}

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

	if action.Input == nil {
		action.Input = &graphql.ActionInput{}
	}

	var typeInstanceDetails []*graphql.InputTypeInstanceDetails
	typeInstanceDetails = append(typeInstanceDetails, action.Input.TypeInstances...)

	for _, typeInstance := range in.TypeInstances {
		typeInstanceDetails = append(typeInstanceDetails,
			&graphql.InputTypeInstanceDetails{
				ID:   typeInstance.ID,
				Name: typeInstance.Name,
				TypeRef: &graphql.ManifestReference{
					Path:     "cap.type.example",
					Revision: "0.1.0",
				},
				Optional: true,
			})
	}

	action.Input.TypeInstances = typeInstanceDetails
	message := conditionMessages[graphql.ActionStatusPhaseBeingRendered]

	action.Status.Phase = graphql.ActionStatusPhaseBeingRendered
	action.Status.Message = &message
	action.Status.Timestamp = graphql.Timestamp{Time: time.Now()}

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
