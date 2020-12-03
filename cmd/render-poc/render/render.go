package render

import (
	"encoding/json"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type Workflow struct {
	Templates  []Template `json:"templates"`
	Entrypoint string     `json:"entrypoint"`
}

type Template struct {
	Name      string           `json:"name"`
	Steps     []ParallelSteps  `json:"steps,omitempty"`
	Inputs    *wfv1.Inputs     `json:"inputs,omitempty"`
	Outputs   *wfv1.Outputs    `json:"outputs,omitempty"`
	Container *apiv1.Container `json:"container,omitempty"`
}

type ParallelSteps []*WorkflowStep

type WorkflowStep struct {
	Name      string         `json:"name"`
	Action    *Action        `json:"action,omitempty"`
	Template  string         `json:"template,omitempty"`
	Arguments wfv1.Arguments `json:"arguments,omitempty"`
}

type Action struct {
	Name string `json:"name"`
}

type Renderer struct {
	Implementations  map[string]*types.Implementation
	RenderedWorkflow Workflow
}

func (r *Renderer) Render(implementation *types.Implementation) (interface{}, error) {
	workflow, err := createWorkflow(implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while creating parse tree")
	}

	r.RenderedWorkflow = *workflow

	for {
		actionSteps := r.findActionSteps()
		if len(actionSteps) == 0 {
			break
		}

		for _, step := range actionSteps {
			if err := r.resolveActionStep(step); err != nil {
				return r.RenderedWorkflow, errors.Wrap(err, "while resolving action step")
			}
		}
	}

	return r.RenderedWorkflow, nil
}

func createWorkflow(implementation *types.Implementation) (*Workflow, error) {
	rawWorkflowSpec := implementation.Spec.Action.Args["workflow"]

	tree := &Workflow{}

	b, _ := json.Marshal(rawWorkflowSpec)
	if err := json.Unmarshal(b, &tree); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling to spec")
	}
	return tree, nil
}

func (r *Renderer) findActionSteps() []*WorkflowStep {
	actionSteps := []*WorkflowStep{}

	for _, tmpl := range r.RenderedWorkflow.Templates {
		for _, parallelSteps := range tmpl.Steps {
			for i := range parallelSteps {
				step := parallelSteps[i]
				if step.Action != nil {
					actionSteps = append(actionSteps, step)
				}
			}
		}
	}

	return actionSteps
}

func (r *Renderer) resolveActionStep(step *WorkflowStep) error {
	importedImpl := r.Implementations[step.Action.Name]

	importedWorkflow, err := createWorkflow(importedImpl)
	if err != nil {
		return errors.Wrap(err, "while creating workflow for imported implementation")
	}

	// import templates
	r.RenderedWorkflow.Templates = append(r.RenderedWorkflow.Templates, importedWorkflow.Templates...)

	// replace action with template reference
	step.Action = nil
	step.Template = importedWorkflow.Entrypoint

	return nil
}
