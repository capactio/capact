// We had to copy the Argo Workflow Go struct as we need to extend the WorkflowStep syntax with our own keywords
package argo

import (
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"regexp"
)

type Workflow struct {
	*wfv1.WorkflowSpec
	Templates  []*Template     `json:"templates"`
}

//func (w *Workflow) AppendWorkflowTemplateAtRoot(template Template) {
//	w.Templates = append(w.Templates, template)
//}

type Template struct {
	*wfv1.Template
	Steps     []ParallelSteps  `json:"steps,omitempty"`
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

var workflowArtifactRefRegex = regexp.MustCompile(`{{workflow\.outputs\.artifacts\.(.+)}}`)

type mapEvalParameters map[string]interface{}

func (p mapEvalParameters) Get(name string) (interface{}, error) {
	value, found := p[name]
	if !found {
		return nil, nil
	}

	return value, nil
}
