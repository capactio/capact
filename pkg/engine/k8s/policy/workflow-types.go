package policy

import (
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

type WorkflowPolicy struct {
	APIVersion string            `json:"apiVersion"`
	Rules      WorkflowRulesList `json:"rules"`
}

type WorkflowRulesList []WorkflowRulesForInterface

type WorkflowRulesForInterface struct {
	// Interface refers to a given Interface manifest.
	Interface types.ManifestRef `json:"interface"`

	OneOf []WorkflowRule `json:"oneOf"`
}

type WorkflowRule struct {
	ImplementationConstraints ImplementationConstraints `json:"implementationConstraints,omitempty"`
	Inject                    *WorkflowInjectData       `json:"inject,omitempty"`
}

type WorkflowInjectData struct {
	AdditionalInput map[string]interface{} `json:"additionalInput,omitempty"`
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
