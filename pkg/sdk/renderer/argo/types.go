// We had to copy the Argo Workflow Go struct as we need to extend the WorkflowStep syntax with our own keywords
package argo

import (
	"regexp"

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
