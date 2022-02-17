package action

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"capact.io/capact/internal/k8s-engine/graphql/model"
	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	authv1 "k8s.io/api/authentication/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	// ParameterDataKeyPrefix prefixes every user input parameter key in the Secret data
	ParameterDataKeyPrefix = "parameter-"
	// ActionPolicySecretDataKey defines key name for Action Policy.
	ActionPolicySecretDataKey = "action-policy.json"
	// LatestRevision defines keyword to indicate latest revision.
	LatestRevision = "latest"

	secretKind = "Secret"
)

// GetParameterDataKey returns the parameter data key in the Secret resource
func GetParameterDataKey(parameterName string) string {
	return fmt.Sprintf("%s%s", ParameterDataKeyPrefix, parameterName)
}

// IsParameterDataKey returns two values given a Secret data key:
// First value - bool indicating, if the key represents an user input parameter
// Second value - name of the user input parameter
func IsParameterDataKey(key string) (bool, string) {
	if !strings.HasPrefix(key, ParameterDataKeyPrefix) {
		return false, ""
	}

	return true, strings.TrimPrefix(key, ParameterDataKeyPrefix)
}

// Converter provides functionality to convert GraphQL DTO to models.
type Converter struct{}

// NewConverter returns a new Converter instance.
func NewConverter() *Converter {
	return &Converter{}
}

// FromGraphQLInput coverts create/update Action input to model.
func (c *Converter) FromGraphQLInput(in graphql.ActionDetailsInput) (model.ActionToCreateOrUpdate, error) {
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

	inputParamsSecret, err := c.inputParamsFromGraphQL(in.Input, in.Name)
	if err != nil {
		return model.ActionToCreateOrUpdate{}, err
	}

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
	}, nil
}

// ToGraphQL converts Kubernetes Action representation to GraphQL DTO.
func (c *Converter) ToGraphQL(in v1alpha1.Action) (graphql.Action, error) {
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
	var err error
	if in.Status.Rendering != nil {
		if in.Status.Rendering.Action != nil {
			renderedAction = c.runtimeExtensionToJSONRawMessage(in.Status.Rendering.Action)
		}

		actionInput, err = c.actionInputToGraphQL(in.Status.Rendering.Input)
		if err != nil {
			return graphql.Action{}, errors.Wrap(err, "while converting ActionInput from CR to GraphQL")
		}
	}

	actionOutput := c.actionOutputToGraphQL(in.Status.Output)

	actionRef := c.manifestRefToGraphQL(&in.Spec.ActionRef)

	return graphql.Action{
		Name:                   in.Name,
		CreatedAt:              graphql.Timestamp{Time: in.CreationTimestamp.Time},
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
	}, nil
}

// FilterFromGraphQL converts GraphQL Action filters to model.
func (c *Converter) FilterFromGraphQL(in *graphql.ActionFilter) (model.ActionFilter, error) {
	if in == nil {
		return model.ActionFilter{}, nil
	}

	var (
		phase        *v1alpha1.ActionPhase
		pattern      *regexp.Regexp
		interfaceRef *v1alpha1.ManifestReference
	)

	if in.Phase != nil {
		phaseValue := c.phaseFromGraphQL(*in.Phase)
		phase = &phaseValue
	}

	if in.NameRegex != nil {
		nPattern, err := regexp.Compile(*in.NameRegex)
		if err != nil {
			return model.ActionFilter{}, errors.Wrap(err, "while compiling regex")
		}
		pattern = nPattern
	}

	if in.InterfaceRef != nil {
		interfaceRef = &v1alpha1.ManifestReference{
			Path:     v1alpha1.NodePath(in.InterfaceRef.Path),
			Revision: in.InterfaceRef.Revision,
		}
	}

	return model.ActionFilter{
		Phase:        phase,
		NameRegex:    pattern,
		InterfaceRef: interfaceRef,
	}, nil
}

// AdvancedModeContinueRenderingInputFromGraphQL converts GraphQL advance mode input to model.
func (c *Converter) AdvancedModeContinueRenderingInputFromGraphQL(in graphql.AdvancedModeContinueRenderingInput) model.AdvancedModeContinueRenderingInput {
	return model.AdvancedModeContinueRenderingInput{
		TypeInstances: c.inputTypeInstanceDataFromGraphQL(in.TypeInstances),
	}
}

func (c *Converter) inputParamsFromGraphQL(in *graphql.ActionInputData, name string) (*v1.Secret, error) {
	if in == nil || (in.Parameters == nil && in.ActionPolicy == nil) {
		return nil, nil
	}

	var (
		data = make(map[string]string)
		err  error
	)

	if in.Parameters != nil {
		data, err = toParametersData(json.RawMessage(*in.Parameters))
		if err != nil {
			return nil, errors.Wrap(err, "while getting parameters collection")
		}
	}

	if in.ActionPolicy != nil {
		policyData, err := json.Marshal(in.ActionPolicy)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling policy to JSON")
		}

		data[ActionPolicySecretDataKey] = string(policyData)
	}

	if len(data) == 0 { // e.g. after unmarshaling we discovered that empty params were submitted
		return nil, nil
	}

	return &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       secretKind,
			APIVersion: v1.SchemeGroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		StringData: data,
	}, nil
}

func toParametersData(parameters json.RawMessage) (map[string]string, error) {
	if parameters == nil {
		return nil, nil
	}

	parametersMap := make(map[string]interface{})
	if err := json.Unmarshal(parameters, &parametersMap); err != nil {
		return make(map[string]string), err
	}

	result := make(map[string]string)

	for name := range parametersMap {
		value := parametersMap[name]
		valueData, err := json.Marshal(&value)
		if err != nil {
			return types.ParametersCollection{}, errors.Wrapf(err, "while marshaling %s parameter to JSON", name)
		}

		key := GetParameterDataKey(name)
		result[key] = string(valueData)
	}

	return result, nil
}

func (c *Converter) actionInputToGraphQL(in *v1alpha1.ResolvedActionInput) (*graphql.ActionInput, error) {
	if in == nil {
		return nil, nil
	}

	result := &graphql.ActionInput{}

	if in.Parameters != nil {
		result.Parameters = c.runtimeExtensionToJSONRawMessage(in.Parameters)
	}

	if in.TypeInstances != nil {
		var gqlTypeInstances []*graphql.InputTypeInstanceDetails
		for _, item := range *in.TypeInstances {
			gqlTypeInstances = append(gqlTypeInstances, &graphql.InputTypeInstanceDetails{
				Name: item.Name,
				ID:   item.ID,
			})
		}
		result.TypeInstances = gqlTypeInstances
	}

	if in.ActionPolicy != nil {
		policyData := c.runtimeExtensionToJSONRawMessage(in.ActionPolicy)

		if policyData != nil {
			if err := json.Unmarshal(*policyData, &result.ActionPolicy); err != nil {
				return nil, err
			}
		}
	}

	return result, nil
}

func (c *Converter) actionOutputToGraphQL(in *v1alpha1.ActionOutput) *graphql.ActionOutput {
	if in == nil || in.TypeInstances == nil {
		return nil
	}

	var gqlTypeInstances []*graphql.OutputTypeInstanceDetails
	for _, item := range *in.TypeInstances {
		gqlBackend := graphql.TypeInstanceBackendDetails(item.Backend)
		gqlTypeInstances = append(gqlTypeInstances, &graphql.OutputTypeInstanceDetails{
			ID:      item.ID,
			TypeRef: c.manifestRefToGraphQL(item.TypeRef),
			Backend: &gqlBackend,
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

	if in.ActionPolicy != nil && inputParamsSecretName != nil {
		actionInput.ActionPolicy = &v1alpha1.ActionPolicy{
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
		Timestamp:  graphql.Timestamp{Time: in.LastTransitionTime.Time},
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
