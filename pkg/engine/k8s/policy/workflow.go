package policy

import (
	"encoding/json"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"capact.io/capact/internal/ptr"
	hubpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// WorkflowPolicy represents a Workflow step policy.
type WorkflowPolicy struct {
	Interface WorkflowInterfacePolicy `json:"interface"`
}

// WorkflowInterfacePolicy represent an Interface policy.
type WorkflowInterfacePolicy struct {
	Rules WorkflowRulesList `json:"rules"`
}

// WorkflowRulesList holds the list of the rules in the policy.
type WorkflowRulesList []WorkflowRulesForInterface

// WorkflowInterfaceRef represents a reference to an Interface
// in the workflow step policy.
// The Interface can be provided either using the full path and revision
// or using an alias from the imported Interfaces in the Implementation.
type WorkflowInterfaceRef struct {
	ManifestRef *types.ManifestRefWithOptRevision
	Alias       *string
}

// WorkflowRulesForInterface holds a single policy rule for an Interface.
type WorkflowRulesForInterface struct {
	// Interface refers to a given Interface manifest.
	Interface WorkflowInterfaceRef `json:"interface"`

	OneOf []WorkflowRule `json:"oneOf"`
}

// WorkflowRule holds the constraints an Implementation must match.
// It also stores data, which should be injected,
// if this Implementation is selected.
type WorkflowRule struct {
	ImplementationConstraints ImplementationConstraints `json:"implementationConstraints,omitempty"`
	Inject                    *WorkflowInjectData       `json:"inject,omitempty"`
}

// WorkflowInjectData holds the data, which should be injected into the Action.
// Compared to other policies, injecting RequiredTypeInstances
// is not supported in the Workflow step policy.
type WorkflowInjectData struct {
	AdditionalParameters []AdditionalParametersToInject `json:"additionalParameters,omitempty"`
}

// ResolveImports is used to resolve the Manifest Reference for the rules,
// if the Interface reference is provided using an alias.
func (p *WorkflowPolicy) ResolveImports(imports []*hubpublicapi.ImplementationImport) error {
	for i, r := range p.Interface.Rules {
		if r.Interface.Alias == nil || *r.Interface.Alias == "" {
			continue
		}
		actionRef, err := hubpublicapi.ResolveActionPathFromImports(imports, *r.Interface.Alias)
		if err != nil {
			return errors.Wrap(err, "while resolving Action path")
		}
		p.Interface.Rules[i].Interface.ManifestRef.Path = actionRef.Path
		p.Interface.Rules[i].Interface.ManifestRef.Revision = &actionRef.Revision
	}
	return nil
}

// ToYAMLBytes marshals the policy into a byte slice.
func (p WorkflowPolicy) ToYAMLBytes() ([]byte, error) {
	bytes, err := yaml.Marshal(&p)

	if err != nil {
		return nil, errors.Wrap(err, "while marshaling policy to YAML")
	}

	return bytes, nil
}

// ToYAMLString converts the policy into a string.
func (p WorkflowPolicy) ToYAMLString() (string, error) {
	bytes, err := p.ToYAMLBytes()
	return string(bytes), err
}

// ToPolicy converts the WorkflowPolicy to a generic Policy struct.
func (p WorkflowPolicy) ToPolicy() (Policy, error) {
	newPolicy := Policy{}
	bytes, err := p.ToYAMLBytes()
	if err != nil {
		return Policy{}, errors.Wrap(err, "while converting to Policy")
	}

	err = yaml.Unmarshal(bytes, &newPolicy)
	if err != nil {
		return Policy{}, errors.Wrap(err, "while converting to Policy when unmarshaling")
	}
	return newPolicy, nil
}

// UnmarshalJSON fills the WorkflowInterfaceRef properties from the provided byte slice.
// The byte slice must be JSON encoded.
func (i *WorkflowInterfaceRef) UnmarshalJSON(b []byte) error {
	i.ManifestRef = &types.ManifestRefWithOptRevision{}
	err := json.Unmarshal(b, i.ManifestRef)
	if err == nil {
		return nil
	}

	i.Alias = ptr.String("")

	return json.Unmarshal(b, i.Alias)
}

// MarshalJSON marshals the WorkflowInterfaceRef to a JSON encoded byte slice.
func (i *WorkflowInterfaceRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.ManifestRef)
}
