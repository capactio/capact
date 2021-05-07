// We had to copy the Argo Workflow Go struct as we need to extend the WorkflowStep syntax with our own keywords
package argo

import (
	"regexp"

	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	wfv1 "github.com/argoproj/argo/v2/pkg/apis/workflow/v1alpha1"
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
	CapactWhen                *string                  `json:"capact-when,omitempty"`
	CapactAction              *string                  `json:"capact-action,omitempty"`
	CapactTypeInstanceOutputs []TypeInstanceDefinition `json:"capact-outputTypeInstances,omitempty"`
	CapactTypeInstanceUpdates []TypeInstanceDefinition `json:"capact-updateTypeInstances,omitempty"`

	// internal fields
	typeInstanceOutputs map[string]*string
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

type RenderInput struct {
	RunnerContextSecretRef RunnerContextSecretRef
	InterfaceRef           types.InterfaceRef
	Options                []RendererOption
}

type RenderOutput struct {
	Action              *types.Action
	TypeInstancesToLock []string
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
