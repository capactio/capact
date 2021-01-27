package action

import (
	"encoding/json"

	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/model"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

const ParametersSecretDataKey = "parameters.json"
const LatestRevision = "latest"
const secretKind = "Secret"

type Converter struct{}

func NewConverter() *Converter {
	return &Converter{}
}

func (c *Converter) FromGraphQLInput(in graphql.ActionDetailsInput) model.ActionToCreateOrUpdate {
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

	inputParamsSecret := c.inputParamsFromGraphQL(in.Input, in.Name)
	var inputParamsSecretName *string
	if inputParamsSecret != nil {
		inputParamsSecretName = &in.Name
	}

	var actionRef v1alpha1.ManifestReference
	if in.ActionRef != nil {
		actionRef = v1alpha1.ManifestReference{
			Path:     v1alpha1.NodePath(in.ActionRef.Path),
			Revision: in.ActionRef.Revision,
		}
	}

	return model.ActionToCreateOrUpdate{
		Action: v1alpha1.Action{
			TypeMeta: metav1.TypeMeta{
				Kind:       v1alpha1.ActionKind,
				APIVersion: v1alpha1.GroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name: in.Name,
			},
			Spec: v1alpha1.ActionSpec{
				DryRun:                 in.DryRun,
				ActionRef:              actionRef,
				Input:                  c.actionInputFromGraphQL(in.Input, inputParamsSecretName),
				AdvancedRendering:      advancedRendering,
				RenderedActionOverride: renderedActionOverride,
			},
		},
		InputParamsSecret: inputParamsSecret,
	}
}

func (c *Converter) ToGraphQL(in v1alpha1.Action) graphql.Action {
	var run bool
	if in.Spec.Run != nil {
		run = *in.Spec.Run
	}

	var dryRun bool
	if in.Spec.DryRun != nil {
		dryRun = *in.Spec.DryRun
	}

	var cancel bool
	if in.Spec.Cancel != nil {
		cancel = *in.Spec.Cancel
	}

	var renderedAction interface{}
	var actionInput *graphql.ActionInput
	if in.Status.Rendering != nil {
		if in.Status.Rendering.Action != nil {
			renderedAction = c.runtimeExtensionToJSONRawMessage(in.Status.Rendering.Action)
		}

		actionInput = c.actionInputToGraphQL(in.Status.Rendering.Input)
	}

	actionOutput := c.actionOutputToGraphQL(in.Status.Output)

	actionRef := c.manifestRefToGraphQL(&in.Spec.ActionRef)

	return graphql.Action{
		Name:                   in.Name,
		CreatedAt:              graphql.Timestamp{in.CreationTimestamp.Time},
		Input:                  actionInput,
		Output:                 actionOutput,
		DryRun:                 dryRun,
		Run:                    run,
		ActionRef:              actionRef,
		Cancel:                 cancel,
		RenderedAction:         renderedAction,
		RenderingAdvancedMode:  c.advancedRenderingToGraphQL(&in),
		RenderedActionOverride: c.runtimeExtensionToJSONRawMessage(in.Spec.RenderedActionOverride),
		Status:                 c.statusToGraphQL(&in.Status),
	}
}

func (c *Converter) FilterFromGraphQL(in graphql.ActionFilter) model.ActionFilter {
	var phase *v1alpha1.ActionPhase
	if in.Phase != nil {
		phaseValue := c.phaseFromGraphQL(*in.Phase)
		phase = &phaseValue
	}
	return model.ActionFilter{
		Phase: phase,
	}
}

func (c *Converter) AdvancedModeContinueRenderingInputFromGraphQL(in graphql.AdvancedModeContinueRenderingInput) model.AdvancedModeContinueRenderingInput {
	return model.AdvancedModeContinueRenderingInput{
		TypeInstances: c.inputTypeInstanceDataFromGraphQL(in.TypeInstances),
	}
}

func (c *Converter) inputParamsFromGraphQL(in *graphql.ActionInputData, name string) *v1.Secret {
	if in == nil || in.Parameters == nil {
		return nil
	}

	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       secretKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		StringData: map[string]string{
			ParametersSecretDataKey: string(*in.Parameters),
		},
	}
}

func (c *Converter) actionInputToGraphQL(in *v1alpha1.ResolvedActionInput) *graphql.ActionInput {
	if in == nil {
		return nil
	}

	result := &graphql.ActionInput{}

	if in.Parameters != nil {
		result.Parameters = c.runtimeExtensionToJSONRawMessage(in.Parameters)
	}

	if in.TypeInstances != nil {
		var gqlTypeInstances []*graphql.InputTypeInstanceDetails
		for _, item := range *in.TypeInstances {
			gqlTypeInstances = append(gqlTypeInstances, &graphql.InputTypeInstanceDetails{
				Name:     item.Name,
				TypeRef:  c.manifestRefToGraphQL(item.TypeRef),
				ID:       item.ID,
				Optional: item.Optional,
			})
		}
		result.TypeInstances = gqlTypeInstances
	}

	return result
}

func (c *Converter) actionOutputToGraphQL(in *v1alpha1.ActionOutput) *graphql.ActionOutput {
	if in == nil || in.TypeInstances == nil {
		return nil
	}

	var gqlTypeInstances []*graphql.OutputTypeInstanceDetails
	for _, item := range *in.TypeInstances {
		gqlTypeInstances = append(gqlTypeInstances, &graphql.OutputTypeInstanceDetails{
			Name:    item.Name,
			ID:      item.ID,
			TypeRef: c.manifestRefToGraphQL(item.TypeRef),
		})
	}

	return &graphql.ActionOutput{
		TypeInstances: gqlTypeInstances,
	}
}

func (c *Converter) actionInputFromGraphQL(in *graphql.ActionInputData, inputParamsSecretName *string) *v1alpha1.ActionInput {
	if in == nil {
		return nil
	}

	actionInput := &v1alpha1.ActionInput{}
	actionInput.TypeInstances = c.inputTypeInstanceDataFromGraphQL(in.TypeInstances)

	if in.TypeInstances != nil && len(in.TypeInstances) > 0 {
		var inputTypeInstances []v1alpha1.InputTypeInstance

		for _, item := range in.TypeInstances {
			inputTypeInstances = append(inputTypeInstances, v1alpha1.InputTypeInstance{
				Name: item.Name,
				ID:   item.ID,
			})
		}

		actionInput.TypeInstances = &inputTypeInstances
	}

	if in.Parameters != nil && inputParamsSecretName != nil {
		actionInput.Parameters = &v1alpha1.InputParameters{
			SecretRef: v1.LocalObjectReference{Name: *inputParamsSecretName},
		}
	}

	return actionInput
}

func (c *Converter) inputTypeInstanceDataFromGraphQL(in []*graphql.InputTypeInstanceData) *[]v1alpha1.InputTypeInstance {
	if len(in) == 0 {
		return nil
	}

	var inputTypeInstances []v1alpha1.InputTypeInstance
	for _, item := range in {
		if item == nil {
			continue
		}

		inputTypeInstances = append(inputTypeInstances, v1alpha1.InputTypeInstance{
			Name: item.Name,
			ID:   item.ID,
		})
	}

	return &inputTypeInstances
}

func (c *Converter) advancedRenderingToGraphQL(in *v1alpha1.Action) *graphql.ActionRenderingAdvancedMode {
	if in == nil || in.Spec.AdvancedRendering == nil {
		return nil
	}

	return &graphql.ActionRenderingAdvancedMode{
		Enabled:                            in.Spec.AdvancedRendering.Enabled,
		TypeInstancesForRenderingIteration: c.typeInstancesForRenderingIterationToGraphQL(in),
	}
}

func (c *Converter) typeInstancesForRenderingIterationToGraphQL(in *v1alpha1.Action) []*graphql.InputTypeInstanceToProvide {
	if in == nil ||
		in.Spec.AdvancedRendering == nil ||
		in.Status.Rendering == nil ||
		in.Status.Rendering.AdvancedRendering == nil ||
		in.Status.Rendering.AdvancedRendering.RenderingIteration == nil ||
		in.Status.Rendering.AdvancedRendering.RenderingIteration.InputTypeInstancesToProvide == nil {
		return nil
	}

	var typeInstancesForRenderingIteration []*graphql.InputTypeInstanceToProvide

	for _, typeInstance := range *in.Status.Rendering.AdvancedRendering.RenderingIteration.InputTypeInstancesToProvide {
		typeInstancesForRenderingIteration = append(typeInstancesForRenderingIteration, &graphql.InputTypeInstanceToProvide{
			Name:    typeInstance.Name,
			TypeRef: c.manifestRefToGraphQL(typeInstance.TypeRef),
		})
	}

	return typeInstancesForRenderingIteration
}

func (c *Converter) statusToGraphQL(in *v1alpha1.ActionStatus) *graphql.ActionStatus {
	var runnerStatus *graphql.RunnerStatus
	if in.Runner != nil {
		runnerStatus = &graphql.RunnerStatus{
			Status: c.runtimeExtensionToJSONRawMessage(in.Runner.Status),
		}
	}

	return &graphql.ActionStatus{
		Phase:      c.phaseToGraphQL(in.Phase),
		Timestamp:  graphql.Timestamp{in.LastTransitionTime.Time},
		Message:    in.Message,
		Runner:     runnerStatus,
		CreatedBy:  c.userInfoToGraphQL(in.CreatedBy),
		RunBy:      c.userInfoToGraphQL(in.RunBy),
		CanceledBy: c.userInfoToGraphQL(in.CanceledBy),
	}
}

func (c *Converter) userInfoToGraphQL(in *authv1.UserInfo) *graphql.UserInfo {
	if in == nil {
		return nil
	}

	extras := map[string][]string{}
	if in.Extra != nil {
		for key, value := range in.Extra {
			extras[key] = value
		}
	}

	return &graphql.UserInfo{
		Username: in.Username,
		Groups:   in.Groups,
		Extra:    extras,
	}
}

func (c *Converter) manifestRefToGraphQL(in *v1alpha1.ManifestReference) *graphql.ManifestReference {
	if in == nil {
		return nil
	}

	var revision string
	if in.Revision != nil {
		revision = *in.Revision
	} else {
		revision = LatestRevision
	}

	return &graphql.ManifestReference{
		Path:     string(in.Path),
		Revision: revision,
	}
}

func (c *Converter) phaseToGraphQL(in v1alpha1.ActionPhase) graphql.ActionStatusPhase {
	switch in {
	case v1alpha1.InitialActionPhase:
		return graphql.ActionStatusPhaseInitial
	case v1alpha1.BeingRenderedActionPhase:
		return graphql.ActionStatusPhaseBeingRendered
	case v1alpha1.AdvancedModeRenderingIterationActionPhase:
		return graphql.ActionStatusPhaseAdvancedModeRenderingIteration
	case v1alpha1.ReadyToRunActionPhase:
		return graphql.ActionStatusPhaseReadyToRun
	case v1alpha1.RunningActionPhase:
		return graphql.ActionStatusPhaseRunning
	case v1alpha1.BeingCanceledActionPhase:
		return graphql.ActionStatusPhaseBeingCanceled
	case v1alpha1.CanceledActionPhase:
		return graphql.ActionStatusPhaseCanceled
	case v1alpha1.SucceededActionPhase:
		return graphql.ActionStatusPhaseSucceeded
	case v1alpha1.FailedActionPhase:
		return graphql.ActionStatusPhaseFailed
	}

	return graphql.ActionStatusPhaseInitial
}

func (c *Converter) phaseFromGraphQL(in graphql.ActionStatusPhase) v1alpha1.ActionPhase {
	switch in {
	case graphql.ActionStatusPhaseInitial:
		return v1alpha1.InitialActionPhase
	case graphql.ActionStatusPhaseBeingRendered:
		return v1alpha1.BeingRenderedActionPhase
	case graphql.ActionStatusPhaseAdvancedModeRenderingIteration:
		return v1alpha1.AdvancedModeRenderingIterationActionPhase
	case graphql.ActionStatusPhaseReadyToRun:
		return v1alpha1.ReadyToRunActionPhase
	case graphql.ActionStatusPhaseRunning:
		return v1alpha1.RunningActionPhase
	case graphql.ActionStatusPhaseBeingCanceled:
		return v1alpha1.BeingCanceledActionPhase
	case graphql.ActionStatusPhaseCanceled:
		return v1alpha1.CanceledActionPhase
	case graphql.ActionStatusPhaseSucceeded:
		return v1alpha1.SucceededActionPhase
	case graphql.ActionStatusPhaseFailed:
		return v1alpha1.FailedActionPhase
	}

	return v1alpha1.InitialActionPhase
}

func (c *Converter) runtimeExtensionToJSONRawMessage(extension *runtime.RawExtension) *json.RawMessage {
	if extension == nil {
		return nil
	}

	var jsonRaw json.RawMessage
	bytes := extension.Raw
	jsonRaw = bytes
	return &jsonRaw
}
