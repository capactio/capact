package policy

import (
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

const (
	// CurrentAPIVersion holds the current Policy API version.
	CurrentAPIVersion = "0.2.0"
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
	APIVersion string    `json:"apiVersion"`
	Rules      RulesList `json:"rules"`
}

// ActionPolicy holds the Action policy properties.
type ActionPolicy Policy

// RulesList holds the list of the rules in the policy.
type RulesList []RulesForInterface

// RulesForInterface holds a single policy rule for an Interface.
type RulesForInterface struct {
	// Interface refers to a given Interface manifest.
	Interface types.ManifestRef `json:"interface"`

	OneOf []Rule `json:"oneOf"`
}

// Rule holds the constraints an Implementation must match.
// It also stores data, which should be injected,
// if this Implementation is selected.
type Rule struct {
	ImplementationConstraints ImplementationConstraints `json:"implementationConstraints,omitempty"`
	Inject                    *InjectData               `json:"inject,omitempty"`
}

// InjectData holds the data, which should be injected into the Action.
type InjectData struct {
	TypeInstances   []TypeInstanceToInject `json:"typeInstances,omitempty"`
	AdditionalInput map[string]interface{} `json:"additionalInput,omitempty"`
}

// ImplementationConstraints represents the constraints
// for an Implementation to match a rule.
type ImplementationConstraints struct {
	// Requires refers a specific requirement path and optional revision.
	Requires *[]types.ManifestRef `json:"requires,omitempty"`

	// Attributes refers a specific Attribute by path and optional revision.
	Attributes *[]types.ManifestRef `json:"attributes,omitempty"`

	// Path refers a specific Implementation with exact path.
	Path *string `json:"path,omitempty"`
}

// TypeInstanceToInject holds a TypeInstances to be injected to the Action.
type TypeInstanceToInject struct {
	ID string `json:"id"`

	// TypeRef refers to a given Type.
	TypeRef types.ManifestRef `json:"typeRef"`
}

// ToYAMLString converts the Policy to a string.
func (p Policy) ToYAMLString() (string, error) {
	bytes, err := yaml.Marshal(&p)

	if err != nil {
		return "", errors.Wrap(err, "while marshaling policy to YAML bytes")
	}

	return string(bytes), nil
}

// controller-gen doesn't support interface{} so writing it manually
func (in *InjectData) DeepCopy() *InjectData {
	if in == nil {
		return nil
	}
	out := new(InjectData)
	in.DeepCopyInto(out)
	return out
}

// controller-gen doesn't support interface{} so writing it manually
func (in *InjectData) DeepCopyInto(out *InjectData) {
	*out = *in
	if in.TypeInstances != nil {
		in, out := &in.TypeInstances, &out.TypeInstances
		*out = make([]TypeInstanceToInject, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.AdditionalInput != nil {
		out.AdditionalInput = MergeMaps(out.AdditionalInput, in.AdditionalInput)
	}
}

// MergeMaps performs a deep merge of two maps.
// It is used to merge the additional parameters in the policies.
func MergeMaps(current, overwrite map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(current))
	for k, v := range current {
		out[k] = v
	}
	for k, v := range overwrite {
		if v, ok := v.(map[string]interface{}); ok {
			if bv, ok := out[k]; ok {
				if bv, ok := bv.(map[string]interface{}); ok {
					out[k] = MergeMaps(bv, v)
					continue
				}
			}
		}
		out[k] = v
	}
	return out
}
