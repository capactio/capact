package clusterpolicy

import (
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type ClusterPolicy struct {
	APIVersion string   `json:"apiVersion"`
	Rules      RulesMap `json:"rules"`
}

const AnyInterfacePath InterfacePath = "cap.*"

// TODO: Change structure to preserve keys order in map, once we support regexes
type RulesMap map[InterfacePath]Rules

type Rules struct {
	OneOf []Rule `json:"oneOf"`
}

type InterfacePath string

type Rule struct {
	ImplementationConstraints ImplementationConstraints `json:"implementationConstraints"`
	InjectTypeInstances       []TypeInstanceToInject    `json:"injectTypeInstances"`
}

type ImplementationConstraints struct {
	// Requires refers a specific requirement by path and optional revision.
	Requires *[]types.TypeRefWithOptRevision `json:"requires,omitempty"`

	// Attributes refers a specific Attribute by path and optional revision.
	Attributes *[]types.AttributeRef `json:"attributes,omitempty"`

	// Path refers a specific Implementation with exact path.
	Path *string `json:"path,omitempty"`
}

type TypeInstanceToInject struct {
	ID      string                       `json:"id"`
	TypeRef types.TypeRefWithOptRevision `json:"typeRef"`
}
