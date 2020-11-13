package action

import (
	"k8s.io/api/authentication/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/model"
	"projectvoltron.dev/voltron/internal/ptr"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

const ParametersSecretDataKey = "parameters"

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) FromGraphQLInput(in graphql.ActionDetailsInput, name, namespace string) model.ActionToCreateOrUpdate {
	var advancedRendering *v1alpha1.AdvancedRendering
	if in.AdvancedRendering != nil {
		advancedRendering = &v1alpha1.AdvancedRendering{
			Enabled: *in.AdvancedRendering,
		}
	}

	var renderedActionOverride *runtime.RawExtension
	if in.RenderedActionOverride != nil {
		renderedActionOverride = &runtime.RawExtension{
			Raw: []byte(*in.RenderedActionOverride),
		}
	}

	inputParamsSecret := c.inputParamsFromGraphQL(in.Input, name, namespace)
	var inputParamsSecretName *string
	if inputParamsSecret != nil {
		inputParamsSecretName = &name
	}

	return model.ActionToCreateOrUpdate{
		Action: v1alpha1.Action{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Spec: v1alpha1.ActionSpec{
				Path:                   v1alpha1.NodePath(in.Action),
				Input:                  c.actionInputFromGraphQL(in.Input, inputParamsSecretName),
				AdvancedRendering:      advancedRendering,
				RenderedActionOverride: renderedActionOverride,
			},
		},
		InputParamsSecret: inputParamsSecret,
	}
}

func (c *Converter) inputParamsFromGraphQL(in *graphql.ActionInputData, name, namespace string) *v1.Secret {
	if in == nil || in.Parameters == nil {
		return nil
	}

	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		StringData: map[string]string{
			ParametersSecretDataKey: string(*in.Parameters),
		},
	}
}

func (c *Converter) ToGraphQL(in v1alpha1.Action) graphql.Action {
	var run bool
	if in.Spec.Run != nil {
		run = *in.Spec.Run
	}

	var cancel bool
	if in.Spec.Cancel != nil {
		cancel = *in.Spec.Cancel
	}

	var renderedAction interface{}
	var actionInput *graphql.ActionInput
	if in.Status.Rendering != nil {
		if in.Status.Rendering.Action != nil {
			renderedAction = in.Status.Rendering.Action
		}

		actionInput = c.actionInputToGraphQL(in.Status.Rendering.Input)
	}

	actionOutput := c.actionOutputToGraphQL(in.Status.Output)

	var advancedRenderingEnabled bool
	if in.Spec.AdvancedRendering != nil {
		advancedRenderingEnabled = in.Spec.AdvancedRendering.Enabled
	}

	return graphql.Action{
		ID:             in.Name,
		CreatedAt:      graphql.Timestamp(in.CreationTimestamp.Time),
		Input:          actionInput,
		Output:         actionOutput,
		Path:           string(in.Spec.Path),
		Run:            run,
		Cancel:         cancel,
		RenderedAction: renderedAction,
		RenderingAdvancedMode: &graphql.ActionRenderingAdvancedMode{
			Enabled:                        advancedRenderingEnabled,
			ArtifactsForRenderingIteration: nil, // TODO: Implement once advanced rendering is supported
		},
		RenderedActionOverride: in.Spec.RenderedActionOverride,
		Status:                 c.statusToGraphQL(&in.Status),
	}
}

func (c *Converter) actionInputToGraphQL(in *v1alpha1.ResolvedActionInput) *graphql.ActionInput {
	if in == nil {
		return nil
	}

	result := &graphql.ActionInput{}

	if in.Parameters != nil {
		result.Parameters = in.Parameters
	}

	if in.Artifacts != nil {
		var gqlArtifacts []*graphql.InputArtifact
		for _, item := range *in.Artifacts {
			gqlArtifacts = append(gqlArtifacts, &graphql.InputArtifact{
				Name:           item.Name,
				TypePath:       string(item.TypePath),
				TypeInstanceID: ptr.String(item.TypeInstanceID),
				Optional:       item.Optional,
			})
		}
		result.Artifacts = gqlArtifacts
	}

	return result
}

func (c *Converter) actionOutputToGraphQL(in *v1alpha1.ActionOutput) *graphql.ActionOutput {
	if in == nil || in.Artifacts == nil {
		return nil
	}

	var gqlArtifacts []*graphql.OutputArtifact
	for _, item := range *in.Artifacts {
		gqlArtifacts = append(gqlArtifacts, &graphql.OutputArtifact{
			Name:           item.Name,
			TypePath:       string(item.TypePath),
			TypeInstanceID: item.TypeInstanceID,
		})
	}

	return &graphql.ActionOutput{
		Artifacts: gqlArtifacts,
	}
}

func (c *Converter) actionInputFromGraphQL(in *graphql.ActionInputData, inputParamsSecretName *string) *v1alpha1.ActionInput {
	if in == nil {
		return nil
	}

	actionInput := &v1alpha1.ActionInput{}
	if in.Artifacts != nil && len(in.Artifacts) > 0 {
		var inputArtifacts []v1alpha1.InputArtifact

		for _, item := range in.Artifacts {
			inputArtifacts = append(inputArtifacts, v1alpha1.InputArtifact{
				Name:           item.Name,
				TypeInstanceID: item.TypeInstanceID,
			})
		}

		actionInput.Artifacts = &inputArtifacts
	}

	if in.Parameters != nil && inputParamsSecretName != nil {
		actionInput.Parameters = &v1alpha1.InputParameters{
			SecretRef: v1.LocalObjectReference{Name: *inputParamsSecretName},
		}
	}

	return actionInput
}

func (c *Converter) statusToGraphQL(in *v1alpha1.ActionStatus) *graphql.ActionStatus {
	var runnerStatus *graphql.RunnerStatus
	if in.Runner != nil {
		runnerStatus = &graphql.RunnerStatus{
			Interface: string(in.Runner.Interface),
			Status:    in.Runner.Status,
		}
	}

	return &graphql.ActionStatus{
		Condition:   c.phaseToGraphQL(in.Phase),
		Timestamp:   graphql.Timestamp(in.LastTransitionTime.Time),
		Message:     in.Message,
		Runner:      runnerStatus,
		CreatedBy:   c.userInfoToGraphQL(in.CreatedBy),
		RunBy:       c.userInfoToGraphQL(in.RunBy),
		CancelledBy: c.userInfoToGraphQL(in.CancelledBy),
	}
}

func (c *Converter) userInfoToGraphQL(in *v1beta1.UserInfo) *graphql.UserInfo {
	if in == nil {
		return nil
	}

	return &graphql.UserInfo{
		Username: in.Username,
		Groups:   in.Groups,
		Extra:    in.Extra,
	}
}

func (c *Converter) phaseToGraphQL(in v1alpha1.ActionPhase) graphql.ActionStatusCondition {
	switch in {
	case v1alpha1.InitialActionPhase:
		return graphql.ActionStatusConditionInitial
	case v1alpha1.BeingRenderedActionPhase:
		return graphql.ActionStatusConditionBeingRendered
	case v1alpha1.AdvancedModeRenderingIterationActionPhase:
		return graphql.ActionStatusConditionAdvancedModeRenderingIteration
	case v1alpha1.ReadyToRunActionPhase:
		return graphql.ActionStatusConditionReadyToRun
	case v1alpha1.RunningActionPhase:
		return graphql.ActionStatusConditionRunning
	case v1alpha1.BeingCancelledActionPhase:
		return graphql.ActionStatusConditionBeingCancelled
	case v1alpha1.CancelledActionPhase:
		return graphql.ActionStatusConditionCancelled
	case v1alpha1.SucceededActionPhase:
		return graphql.ActionStatusConditionSucceeded
	case v1alpha1.FailedActionPhase:
		return graphql.ActionStatusConditionFailed
	}

	return graphql.ActionStatusConditionInitial
}
