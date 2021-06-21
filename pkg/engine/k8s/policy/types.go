package policy

import (
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

const (
	CurrentAPIVersion        = "0.2.0"
	AnyInterfacePath  string = "cap.*"
)

type Type string
type MergeOrder []Type

const (
	Global   Type = "GLOBAL"
	Action   Type = "ACTION"
	Workflow Type = "WORKFLOW"
)

type Policy struct {
	APIVersion string    `json:"apiVersion"`
	Rules      RulesList `json:"rules"`
}

type ActionPolicy Policy

type RulesList []RulesForInterface

type RulesForInterface struct {
	// Interface refers to a given Interface manifest.
	Interface types.ManifestRef `json:"interface"`

	OneOf []Rule `json:"oneOf"`
}

type Rule struct {
	ImplementationConstraints ImplementationConstraints `json:"implementationConstraints,omitempty"`
	Inject                    *InjectData               `json:"inject,omitempty"`
}

type InjectData struct {
	TypeInstances   []TypeInstanceToInject `json:"typeInstances,omitempty"`
	AdditionalInput map[string]interface{} `json:"additionalInput,omitempty"`
}

type ImplementationConstraints struct {
	// Requires refers a specific requirement path and optional revision.
	Requires *[]types.ManifestRef `json:"requires,omitempty"`

	// Attributes refers a specific Attribute by path and optional revision.
	Attributes *[]types.ManifestRef `json:"attributes,omitempty"`

	// Path refers a specific Implementation with exact path.
	Path *string `json:"path,omitempty"`
}

type TypeInstanceToInject struct {
	ID string `json:"id"`

	// TypeRef refers to a given Type.
	TypeRef types.ManifestRef `json:"typeRef"`
}

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
