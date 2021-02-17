package clusterpolicy

import (
	"github.com/pkg/errors"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"sigs.k8s.io/yaml"
)

const supportedApiVersion = "0.1.0"

type ClusterPolicy struct {
	APIVersion string             `json:"apiVersion"`
	Rules      map[InterfacePath]ClusterPolicyRules `json:"rules"`
}

type ClusterPolicyRules struct {
	OneOf []ClusterPolicyRule `json:"oneOf"`
}

type InterfacePath string

type ClusterPolicyRule struct {
	ImplementationConstraints ImplementationConstraints `json:"implementationConstraints"`
	InjectTypeInstances       []TypeInstanceToInject    `json:"injectTypeInstances"`
}

type ImplementationConstraints struct {
	// Requires refers a specific requirement by path and optional revision.
	Requires *[]ImplementationManifestRefConstraint `json:"requires,omitempty"`

	// Attributes refers a specific Attribute by path and optional revision.
	Attributes *[]ImplementationManifestRefConstraint `json:"attributes,omitempty"`

	// Exact refers a specific Implementation by path and optional revision.
	Exact *ImplementationManifestRefConstraint `json:"path,omitempty"`
}

type TypeInstanceToInject struct {
	ID      string        `json:"id"`
	TypeRef types.TypeRef `json:"typeRef"`
}

type ImplementationManifestRefConstraint struct {
	Path     string  `json:"path"`
	Revision *string `json:"revision,omitempty"`
}

func FromYAMLBytes(in []byte) (ClusterPolicy, error) {
	var policy ClusterPolicy
	if err := yaml.Unmarshal(in, &policy); err != nil {
		return ClusterPolicy{}, errors.Wrap(err, "while unmarshalling policy from YAML „bytes")
	}

	return policy, nil
}
