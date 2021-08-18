package policy

import (
	"context"
	"encoding/json"
	"fmt"

	"capact.io/capact/internal/maps"

	hublocalgraphql "capact.io/capact/pkg/hub/api/graphql/local"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

const (
	// AnyInterfacePath holds a value, which represents any Interface path.
	AnyInterfacePath string = "cap.*"
)

// Type is the type of the Policy.
type Type string

// MergeOrder holds the merge order of the Policies.
type MergeOrder []Type

const (
	// Global indicates the Global policy.
	Global Type = "GLOBAL"
	// Action indicates the Action policy.
	Action Type = "ACTION"
	// Workflow indicates the Workflow step policy.
	Workflow Type = "WORKFLOW"
)

// Policy holds the policy properties.
type Policy struct {
	Rules RulesList `json:"rules"`
}

// ActionPolicy holds the Action policy properties.
type ActionPolicy Policy

// RulesList holds the list of the rules in the policy.
type RulesList []RulesForInterface

// RulesForInterface holds a single policy rule for an Interface.
// +kubebuilder:object:generate=true
type RulesForInterface struct {
	// Interface refers to a given Interface manifest.
	Interface types.ManifestRefWithOptRevision `json:"interface"`

	OneOf []Rule `json:"oneOf"`
}

// Rule holds the constraints an Implementation must match.
// It also stores data, which should be injected,
// if this Implementation is selected.
// +kubebuilder:object:generate=true
type Rule struct {
	ImplementationConstraints ImplementationConstraints `json:"implementationConstraints,omitempty"`
	Inject                    *InjectData               `json:"inject,omitempty"`
}

// RequiredTypeInstancesToInject returns required TypeInstances to inject for a given rule.
func (in *Rule) RequiredTypeInstancesToInject() []RequiredTypeInstanceToInject {
	if in == nil || in.Inject == nil {
		return nil
	}
	return in.Inject.RequiredTypeInstances
}

// ValidateTypeInstanceMetadata validates whether the TypeInstance injection metadata are resolved.
func (in *Rule) ValidateTypeInstanceMetadata() error {
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterTypeInstancesWithEmptyTypeRef)
	return validateTypeInstancesMetadata(unresolvedTypeInstances)
}

func (in *Rule) filterRequiredTypeInstances(filterFn func(ti RequiredTypeInstanceToInject) bool) []RequiredTypeInstanceToInject {
	if in.Inject == nil {
		return nil
	}

	var typeInstances []RequiredTypeInstanceToInject
	for _, tiToInject := range in.Inject.RequiredTypeInstances {
		if !filterFn(tiToInject) {
			continue
		}

		typeInstances = append(typeInstances, tiToInject)
	}

	return typeInstances
}

// InjectData holds the data, which should be injected into the Action.
type InjectData struct {
	RequiredTypeInstances []RequiredTypeInstanceToInject `json:"requiredTypeInstances,omitempty"`
	AdditionalParameters  []AdditionalParametersToInject `json:"additionalParameters,omitempty"`
}

// AdditionalParametersToInject holds parameters to be injected to the Action.
type AdditionalParametersToInject struct {
	// Name refers to parameter name.
	Name string `json:"name"`
	// Value holds provided parameters.
	Value map[string]interface{} `json:"value"`
}

// ImplementationConstraints represents the constraints
// for an Implementation to match a rule.
// +kubebuilder:object:generate=true
type ImplementationConstraints struct {
	// Requires refers a specific requirement path and optional revision.
	Requires *[]types.ManifestRefWithOptRevision `json:"requires,omitempty"`

	// Attributes refers a specific Attribute by path and optional revision.
	Attributes *[]types.ManifestRefWithOptRevision `json:"attributes,omitempty"`

	// Path refers a specific Implementation with exact path.
	Path *string `json:"path,omitempty"`
}

// RequiredTypeInstanceToInject holds a RequiredTypeInstances to be injected to the Action.
// +kubebuilder:object:generate=true
type RequiredTypeInstanceToInject struct {
	// RequiredTypeInstanceReference is a reference to TypeInstance provided by user.
	RequiredTypeInstanceReference `json:",inline"`

	// TypeRef refers to a given Type.
	TypeRef *types.ManifestRef `json:"typeRef"`
}

// RequiredTypeInstanceReference is a reference to TypeInstance provided by user.
// +kubebuilder:object:generate=true
type RequiredTypeInstanceReference struct {
	// ID is the TypeInstance identifier.
	ID string `json:"id"`

	// Description contains user's description for a given RequiredTypeInstanceToInject.
	Description *string `json:"description,omitempty"`
}

// UnmarshalJSON unmarshalls RequiredTypeInstanceToInject from bytes. It ignores all fields apart from RequiredTypeInstanceReference files.
func (in *RequiredTypeInstanceToInject) UnmarshalJSON(bytes []byte) error {
	var out RequiredTypeInstanceReference
	if err := json.Unmarshal(bytes, &out); err != nil {
		return err
	}

	in.RequiredTypeInstanceReference = out

	return nil
}

// ToYAMLString converts the Policy to a string.
func (in Policy) ToYAMLString() (string, error) {
	bytes, err := yaml.Marshal(&in)
	if err != nil {
		return "", errors.Wrap(err, "while marshaling policy to YAML")
	}

	return string(bytes), nil
}

// HubClient defines Hub client which is able to find TypeInstance Type references.
type HubClient interface {
	FindTypeInstancesTypeRef(ctx context.Context, ids []string) (map[string]hublocalgraphql.TypeInstanceTypeReference, error)
}

// ResolveTypeInstanceMetadata resolves needed TypeInstance metadata based on IDs.
func (in *Policy) ResolveTypeInstanceMetadata(ctx context.Context, hubCli HubClient) error {
	if in == nil {
		return errors.New("policy cannot be nil")
	}

	if hubCli == nil {
		return errors.New("hub client cannot be nil")
	}

	err := in.resolveTypeRefsForRequiredTypeInstances(ctx, hubCli)
	if err != nil {
		return err
	}

	err = in.ValidateTypeInstancesMetadata()
	if err != nil {
		return errors.Wrap(err, "while TypeInstance metadata validation after resolving TypeRefs")
	}

	return nil
}

// AreTypeInstancesMetadataResolved returns whether every TypeInstance has metadata resolved.
func (in *Policy) AreTypeInstancesMetadataResolved() bool {
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterTypeInstancesWithEmptyTypeRef)

	return len(unresolvedTypeInstances) == 0
}

// ValidateTypeInstancesMetadata validates that every TypeInstance has metadata resolved.
func (in *Policy) ValidateTypeInstancesMetadata() error {
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterTypeInstancesWithEmptyTypeRef)
	return validateTypeInstancesMetadata(unresolvedTypeInstances)
}

func (in *Policy) resolveTypeRefsForRequiredTypeInstances(ctx context.Context, hubCli HubClient) error {
	unresolvedTypeInstances := in.filterRequiredTypeInstances(filterTypeInstancesWithEmptyTypeRef)

	var idsToQuery []string
	for _, ti := range unresolvedTypeInstances {
		idsToQuery = append(idsToQuery, ti.ID)
	}

	if len(idsToQuery) == 0 {
		return nil
	}

	res, err := hubCli.FindTypeInstancesTypeRef(ctx, idsToQuery)
	if err != nil {
		return errors.Wrap(err, "while finding TypeRef for TypeInstances")
	}

	for ruleIdx, rule := range in.Rules {
		for ruleItemIdx, ruleItem := range rule.OneOf {
			if ruleItem.Inject == nil {
				continue
			}
			for reqTIIdx, reqTI := range ruleItem.Inject.RequiredTypeInstances {
				typeRef, exists := res[reqTI.ID]
				if !exists {
					continue
				}

				in.Rules[ruleIdx].OneOf[ruleItemIdx].Inject.RequiredTypeInstances[reqTIIdx].TypeRef = &types.ManifestRef{
					Path:     typeRef.Path,
					Revision: typeRef.Revision,
				}
			}
		}
	}

	return nil
}

func (in *Policy) filterRequiredTypeInstances(filterFn func(ti RequiredTypeInstanceToInject) bool) []RequiredTypeInstanceToInject {
	var typeInstances []RequiredTypeInstanceToInject
	for _, rule := range in.Rules {
		for _, ruleItem := range rule.OneOf {
			typeInstances = append(typeInstances, ruleItem.filterRequiredTypeInstances(filterFn)...)
		}
	}

	return typeInstances
}

var filterTypeInstancesWithEmptyTypeRef = func(ti RequiredTypeInstanceToInject) bool {
	return ti.TypeRef == nil || ti.TypeRef.Path == "" || ti.TypeRef.Revision == ""
}

// DeepCopyInto writes a deep copy of AdditionalParametersToInject into out.
// controller-gen doesn't support interface{} so writing it manually
func (in *AdditionalParametersToInject) DeepCopyInto(out *AdditionalParametersToInject) {
	*out = *in
	out.Value = maps.Merge(out.Value, in.Value)
}

// DeepCopy returns a new deep copy of AdditionalParametersToInject.
// controller-gen doesn't support interface{} so writing it manually
func (in *AdditionalParametersToInject) DeepCopy() *AdditionalParametersToInject {
	if in == nil {
		return nil
	}
	out := new(AdditionalParametersToInject)
	in.DeepCopyInto(out)
	return out
}

// DeepCopy returns a new deep copy of InjectData.
// controller-gen doesn't support interface{} so writing it manually
func (in *InjectData) DeepCopy() *InjectData {
	if in == nil {
		return nil
	}
	out := new(InjectData)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto writes a deep copy of InjectData into out.
// controller-gen doesn't support interface{} so writing it manually
func (in *InjectData) DeepCopyInto(out *InjectData) {
	*out = *in
	if in.RequiredTypeInstances != nil {
		in, out := &in.RequiredTypeInstances, &out.RequiredTypeInstances
		*out = make([]RequiredTypeInstanceToInject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AdditionalParameters != nil {
		in, out := &in.AdditionalParameters, &out.AdditionalParameters
		*out = make([]AdditionalParametersToInject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

func validateTypeInstancesMetadata(requiredTypeInstances []RequiredTypeInstanceToInject) error {
	if len(requiredTypeInstances) == 0 {
		return nil
	}

	multiErr := &multierror.Error{}
	for _, ti := range requiredTypeInstances {
		tiDesc := ""
		if ti.Description != nil {
			tiDesc = *ti.Description
		}

		multiErr = multierror.Append(
			multiErr,
			fmt.Errorf("missing Type reference for TypeInstance %q (description: %q)", ti.ID, tiDesc),
		)
	}

	return errors.Wrap(multiErr, "while validating TypeInstance metadata for Policy")
}
