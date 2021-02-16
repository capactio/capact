// We had to copy the Argo Workflow Go struct as we need to extend the WorkflowStep syntax with our own keywords
package argo

import (
	"regexp"

	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
)

type Workflow struct {
	*wfv1.WorkflowSpec
	Templates []*Template `json:"templates"`
}

type Template struct {
	*wfv1.Template
	Steps []ParallelSteps `json:"steps,omitempty"`
}

type ParallelSteps []*WorkflowStep

type WorkflowStep struct {
	*wfv1.WorkflowStep
	VoltronWhen                *string                  `json:"voltron-when,omitempty"`
	VoltronAction              *string                  `json:"voltron-action,omitempty"`
	VoltronTypeInstanceOutputs []TypeInstanceDefinition `json:"voltron-outputTypeInstances,omitempty"`
}

type TypeInstanceDefinition struct {
	Name string `json:"name"`
	From string `json:"from"`
}

type RunnerContextSecretRef struct {
	Name string
	Key  string
}

type UserInputSecretRef struct {
	Name string
	Key  string
}

var workflowArtifactRefRegex = regexp.MustCompile(`{{workflow\.outputs\.artifacts\.(.+)}}`)

type mapEvalParameters struct {
	items        map[string]interface{}
	lastAccessed string
}

func (p *mapEvalParameters) Set(name string) {
	if p.items == nil {
		p.items = map[string]interface{}{}
	}

	p.items[name] = name
}

func (p *mapEvalParameters) Get(name string) (interface{}, error) {
	value, found := p.items[name]
	if !found {
		return nil, nil
	}

	p.lastAccessed = name
	return value, nil
}

// TODO: Move it to engine/apis?
type ClusterPolicy struct {
	APIVersion string             `json:"apiVersion"`
	Rules      ClusterPolicyRules `json:"rules"`
}

type ClusterPolicyRules struct {
	OneOf map[InterfacePath]ClusterPolicyRule `json:"oneOf"`
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
