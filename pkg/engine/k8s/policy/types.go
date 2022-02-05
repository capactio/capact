package policy

import (
	"encoding/json"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
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
	Interface InterfacePolicy `json:"interface"`
}

// InterfacePolicy holds the Policy for Interfaces.
type InterfacePolicy struct {
	Rules RulesList `json:"rules"`
}

// ActionPolicy holds the Policy injected during Action creation properties.
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

// AdditionalTypeInstancesToInject returns additional TypeInstances to inject for a given rule.
func (in *Rule) AdditionalTypeInstancesToInject() []AdditionalTypeInstanceToInject {
	if in == nil || in.Inject == nil {
		return nil
	}
	return in.Inject.AdditionalTypeInstances
}

// InjectData holds the data, which should be injected into the Action.
type InjectData struct {
	RequiredTypeInstances   []RequiredTypeInstanceToInject   `json:"requiredTypeInstances,omitempty"`
	AdditionalParameters    []AdditionalParametersToInject   `json:"additionalParameters,omitempty"`
	AdditionalTypeInstances []AdditionalTypeInstanceToInject `json:"additionalTypeInstances,omitempty"`
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

// AdditionalTypeInstanceToInject is used to represent additional TypeInstance injection for a given Implementation.
// +kubebuilder:object:generate=true
type AdditionalTypeInstanceToInject struct {
	// AdditionalTypeInstanceReference is a reference to TypeInstance provided by user.
	AdditionalTypeInstanceReference `json:",inline"`

	// TypeRef refers to a given Type.
	TypeRef *types.ManifestRef `json:"typeRef"`
}

// AdditionalTypeInstanceReference is a reference to TypeInstance provided by user.
// +kubebuilder:object:generate=true
type AdditionalTypeInstanceReference struct {
	// Name is the TypeInstance name specific for a given Implementation.
	Name string `json:"name"`

	// ID is the TypeInstance identifier.
	ID string `json:"id"`
}

// UnmarshalJSON unmarshalls AdditionalTypeInstanceToInject from bytes. It ignores all fields apart from AdditionalTypeInstanceReference files.
func (in *AdditionalTypeInstanceToInject) UnmarshalJSON(bytes []byte) error {
	var out AdditionalTypeInstanceReference
	if err := json.Unmarshal(bytes, &out); err != nil {
		return err
	}

	in.AdditionalTypeInstanceReference = out

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
