package argo

// We had to copy the Argo Workflow Go struct as we need to extend the WorkflowStep syntax with our own keywords.

import (
	"regexp"

	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
)

// Workflow is the specification of a Workflow.
type Workflow struct {
	*wfv1.WorkflowSpec
	Templates []*Template `json:"templates"`
}

// Template is a reusable and composable unit of execution in a workflow.
type Template struct {
	*wfv1.Template
	Steps []ParallelSteps `json:"steps,omitempty"`
}

// ParallelSteps define a series of sequential/parallel workflow steps.
type ParallelSteps []*WorkflowStep

// WorkflowStep is a reference to a template to execute in a series of step.
// It extends the Argo WorkflowStep and adds Capact specific properties.
type WorkflowStep struct {
	*wfv1.WorkflowStep
	CapactWhen                *string                     `json:"capact-when,omitempty"`
	CapactAction              *string                     `json:"capact-action,omitempty"`
	CapactPolicy              *policy.WorkflowPolicy      `json:"capact-policy,omitempty"`
	CapactTypeInstanceOutputs []CapactTypeInstanceOutputs `json:"capact-outputTypeInstances,omitempty"`
	CapactTypeInstanceUpdates []TypeInstanceDefinition    `json:"capact-updateTypeInstances,omitempty"`

	// internal fields
	typeInstanceOutputs map[string]*string
}

// CapactTypeInstanceOutputs holds data defined for `capact-outputTypeInstances` field on step level.
type CapactTypeInstanceOutputs struct {
	TypeInstanceDefinition `json:",inline"`
	Backend                *string `json:"backend,omitempty"`
}

// TypeInstanceDefinition represents a TypeInstance,
// which is created in a step.
type TypeInstanceDefinition struct {
	Name string `json:"name"`
	From string `json:"from"`
}

// RunnerContextSecretRef hold a reference to the runner context
// in a Kubernetes Secret resource.
type RunnerContextSecretRef struct {
	Name string
	Key  string
}

// UserInputSecretRef hold a reference to the runner context
// in a Kubernetes Secret resource.
type UserInputSecretRef struct {
	Name string
}

// RenderInput holds the input parameters to the Render method.
type RenderInput struct {
	RunnerContextSecretRef RunnerContextSecretRef
	InterfaceRef           types.InterfaceRef
	Options                []RendererOption
}

// RenderOutput holds the output of the Render method.
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
