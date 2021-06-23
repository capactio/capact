package policy

import (
	"encoding/json"

	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"

	"capact.io/capact/internal/ptr"
	hubpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

type WorkflowPolicy struct {
	APIVersion string            `json:"apiVersion"`
	Rules      WorkflowRulesList `json:"rules"`
}

type WorkflowRulesList []WorkflowRulesForInterface

type WorkflowInterfaceRef struct {
	ManifestRef *types.ManifestRef
	Alias       *string
}

type WorkflowRulesForInterface struct {
	// Interface refers to a given Interface manifest.
	Interface WorkflowInterfaceRef `json:"interface"`

	OneOf []WorkflowRule `json:"oneOf"`
}

type WorkflowRule struct {
	ImplementationConstraints ImplementationConstraints `json:"implementationConstraints,omitempty"`
	Inject                    *WorkflowInjectData       `json:"inject,omitempty"`
}

type WorkflowInjectData struct {
	AdditionalInput map[string]interface{} `json:"additionalInput,omitempty"`
}

func (p *WorkflowPolicy) ResolveImports(imports []*hubpublicapi.ImplementationImport) error {
	for i, r := range p.Rules {
		if r.Interface.Alias == nil || *r.Interface.Alias == "" {
			continue
		}
		actionRef, err := hubpublicapi.ResolveActionPathFromImports(imports, *r.Interface.Alias)
		if err != nil {
			return errors.Wrap(err, "while resolving Action path")
		}
		p.Rules[i].Interface.ManifestRef.Path = actionRef.Path
		p.Rules[i].Interface.ManifestRef.Revision = &actionRef.Revision
	}
	return nil
}

func (p WorkflowPolicy) ToYAMLBytes() ([]byte, error) {
	bytes, err := yaml.Marshal(&p)

	if err != nil {
		return nil, errors.Wrap(err, "while marshaling policy to YAML bytes")
	}

	return bytes, nil
}

func (p WorkflowPolicy) ToYAMLString() (string, error) {
	bytes, err := p.ToYAMLBytes()
	return string(bytes), err
}

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

func (i *WorkflowInterfaceRef) UnmarshalJSON(b []byte) error {
	i.ManifestRef = &types.ManifestRef{}
	err := json.Unmarshal(b, i.ManifestRef)
	if err == nil {
		return nil
	}

	i.Alias = ptr.String("")
	if err := json.Unmarshal(b, i.Alias); err != nil {
		return err
	}
	return nil
}

func (i *WorkflowInterfaceRef) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.ManifestRef)
}
