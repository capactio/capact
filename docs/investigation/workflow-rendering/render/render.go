package render

import (
	"encoding/json"
	"fmt"
	"regexp"

	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type Workflow struct {
	Templates  []Template     `json:"templates"`
	Arguments  wfv1.Arguments `json:"arguments,omitempty" protobuf:"bytes,3,opt,name=arguments"`
	Entrypoint string         `json:"entrypoint"`
}

type Template struct {
	Name      string           `json:"name"`
	Steps     []ParallelSteps  `json:"steps,omitempty"`
	Inputs    wfv1.Inputs      `json:"inputs,omitempty"`
	Outputs   wfv1.Outputs     `json:"outputs,omitempty"`
	Container *apiv1.Container `json:"container,omitempty"`
}

type ParallelSteps []*WorkflowStep

type WorkflowStep struct {
	Name      string         `json:"name"`
	Template  string         `json:"template,omitempty"`
	Arguments wfv1.Arguments `json:"arguments,omitempty"`

	ProvidesInstance *string                     `json:"providesInstance,omitempty"`
	Action           *v1alpha1.ManifestReference `json:"action,omitempty"`
	TypeInstances    []TypeInstanceDefinition    `json:"typeInstances,omitempty"`
}

type Action struct {
	Name     string `json:"name"`
	Revision string `json:"revision"`
}

type TypeInstanceDefinition struct {
	Name string `json:"name"`
	From string `json:"from"`
}

type Renderer struct {
	ManifestStore    *ManifestStore
	RenderedWorkflow *Workflow
}

type actionStepRef struct {
	Path string
	Step *WorkflowStep
}

var workflowArtifactRefRegex = regexp.MustCompile(`{{workflow\.outputs\.artifacts\.(.+)}}`)

func (r *Renderer) Render(implementation *types.Implementation, parameters map[string]interface{}, typeInstances []*v1alpha1.InputTypeInstance) (*Workflow, error) {
	workflow, err := createWorkflow(implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while creating workflow")
	}

	r.RenderedWorkflow = workflow

	parameterRawData, err := yaml.Marshal(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal input parameters")
	}

	// add Action InputParamaters to Workflow
	r.RenderedWorkflow.Arguments.Artifacts = append(workflow.Arguments.Artifacts, wfv1.Artifact{
		Name: "input-parameters",
		ArtifactLocation: wfv1.ArtifactLocation{
			Raw: &wfv1.RawArtifact{
				Data: string(parameterRawData),
			},
		},
	})

	// inject TypeInstances to the workflow
	for _, tiInput := range typeInstances {
		template, err := r.getInjectTypeInstanceTemplate(*tiInput)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting inject TypeInstance template for %s", tiInput.ID)
		}

		r.RenderedWorkflow.Templates[0].Steps = append([]ParallelSteps{
			{
				&WorkflowStep{
					Name:     fmt.Sprintf("%s-step", template.Name),
					Template: template.Name,
				},
			},
		}, r.RenderedWorkflow.Templates[0].Steps...)

		r.RenderedWorkflow.Templates = append(r.RenderedWorkflow.Templates, *template)
	}

	// rendering iterations
	for {
		// remove steps, which would create TypeInstances, which are already provided
		r.removeActionStepForProvidedTypeInstances(typeInstances)

		// render action steps
		actionStepRefs := r.findActionSteps()
		if len(actionStepRefs) == 0 {
			break
		}

		artifactMappings := map[string]string{}

		for _, tmpl := range r.RenderedWorkflow.Templates {
			for _, parallelSteps := range tmpl.Steps {
				for i := range parallelSteps {
					step := parallelSteps[i]

					// render TypeInstances step
					if step.TypeInstances != nil {
						r.renderSetTypeInstanceStep(r.RenderedWorkflow, step)
					}

					if step.Action != nil {
						// flatten the workflows
						importedWorkflow, newArtifactMappings, err := r.renderFunc(tmpl.Name, *step.Action)
						if err != nil {
							return workflow, errors.Wrap(err, "while creating workflow for action step")
						}

						for k, v := range newArtifactMappings {
							artifactMappings[k] = v
						}

						r.RenderedWorkflow.Templates = append(r.RenderedWorkflow.Templates, importedWorkflow.Templates...)
						step.Template = importedWorkflow.Entrypoint
						step.Action = nil
					}

					// replace global artifacts names in references
					for artIdx := range step.Arguments.Artifacts {
						art := &step.Arguments.Artifacts[artIdx]

						match := workflowArtifactRefRegex.FindStringSubmatch(art.From)
						if len(match) > 0 {
							oldArtifactName := match[1]
							if newArtifactName, ok := artifactMappings[oldArtifactName]; ok {
								art.From = fmt.Sprintf("{{workflow.outputs.artifacts.%s}}", newArtifactName)
							}
						}
					}
				}
			}
		}
	}

	return r.RenderedWorkflow, nil
}

func (r *Renderer) renderFunc(prefix string,
	manifestRef v1alpha1.ManifestReference,
) (*Workflow, map[string]string, error) {
	implementation := r.ManifestStore.GetImplementationForInterface(manifestRef)
	if implementation == nil {
		return nil, nil, fmt.Errorf("implementation for %v not found", manifestRef)
	}

	workflow, err := createWorkflow(implementation)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating workflow from implementation")
	}

	for _, tmpl := range workflow.Templates {
		for _, parallelSteps := range tmpl.Steps {
			for i := range parallelSteps {
				step := parallelSteps[i]

				if step.TypeInstances != nil {
					r.renderSetTypeInstanceStep(workflow, step)
				}

				step.Name = fmt.Sprintf("%s-%s", prefix, step.Name)
			}
		}
	}

	artifactsNameMapping := map[string]string{}

	for i := range workflow.Templates {
		tmpl := &workflow.Templates[i]
		tmpl.Name = fmt.Sprintf("%s-%s", prefix, tmpl.Name)

		for artIdx := range tmpl.Outputs.Artifacts {
			artifact := &tmpl.Outputs.Artifacts[artIdx]

			if artifact.GlobalName != "" {
				newName := fmt.Sprintf("%s-%s", prefix, artifact.GlobalName)
				artifactsNameMapping[artifact.GlobalName] = newName
				artifact.GlobalName = newName
			}
		}
	}

	workflow.Entrypoint = fmt.Sprintf("%s-%s", prefix, workflow.Entrypoint)

	return workflow, artifactsNameMapping, nil
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

func (r *Renderer) removeActionStepForProvidedTypeInstances(instances []*v1alpha1.InputTypeInstance) {
	for i, tmpl := range r.RenderedWorkflow.Templates {
		newSteps := []ParallelSteps{}

		for _, parallelSteps := range tmpl.Steps {
			newParallelSteps := []*WorkflowStep{}

			for i := range parallelSteps {
				step := parallelSteps[i]

				if step.ProvidesInstance != nil && containsTypeInstance(instances, *step.ProvidesInstance) != nil {
					continue
				}

				step.ProvidesInstance = nil

				newParallelSteps = append(newParallelSteps, step)
			}

			if len(newParallelSteps) > 0 {
				newSteps = append(newSteps, newParallelSteps)
			}
		}

		r.RenderedWorkflow.Templates[i].Steps = newSteps
	}
}

func (r *Renderer) findActionSteps() []actionStepRef {
	actionStepsRef := []actionStepRef{}

	for _, tmpl := range r.RenderedWorkflow.Templates {
		for _, parallelSteps := range tmpl.Steps {
			for i := range parallelSteps {
				step := parallelSteps[i]
				if step.Action != nil {
					actionStepsRef = append(actionStepsRef, actionStepRef{
						Path: tmpl.Name,
						Step: step,
					})
				}
			}
		}
	}

	return actionStepsRef
}

func containsTypeInstance(instances []*v1alpha1.InputTypeInstance, name string) *v1alpha1.InputTypeInstance {
	for i := range instances {
		instance := instances[i]

		if instance.Name == name {
			return instance
		}
	}

	return nil
}

func (r *Renderer) getInjectTypeInstanceTemplate(input v1alpha1.InputTypeInstance) (*Template, error) {
	typeInstance := r.ManifestStore.GetTypeInstance(input.ID)
	if typeInstance == nil {
		return nil, fmt.Errorf("failed to find TypeInstance %s", input.ID)
	}

	data, err := yaml.Marshal(typeInstance.Spec.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while to marshal TypeInstance to YAML")
	}

	return &Template{
		Name: fmt.Sprintf("inject-%s", input.Name),
		Container: &apiv1.Container{
			Image:   "alpine:3.7",
			Command: []string{"sh", "-c"},
			Args:    []string{fmt.Sprintf("sleep 2 && echo '%s' | tee /output", string(data))},
		},
		Outputs: wfv1.Outputs{
			Artifacts: wfv1.Artifacts{
				{
					Name:       input.Name,
					GlobalName: input.Name,
					Path:       "/output",
				},
			},
		},
	}, nil
}

func (r *Renderer) renderSetTypeInstanceStep(workflow *Workflow, step *WorkflowStep) {
	template := Template{
		Name: step.Name,
		Container: &apiv1.Container{
			Image:   "alpine:3.7",
			Command: []string{"sh", "-c"},
			Args:    []string{"sleep 2"},
		},
		Inputs: wfv1.Inputs{
			Artifacts: wfv1.Artifacts{},
		},
		Outputs: wfv1.Outputs{
			Artifacts: wfv1.Artifacts{},
		},
	}

	for _, typeInstanceDef := range step.TypeInstances {
		step.Arguments.Artifacts = append(step.Arguments.Artifacts, wfv1.Artifact{
			Name: typeInstanceDef.Name,
			From: typeInstanceDef.From,
		})

		template.Inputs.Artifacts = append(template.Inputs.Artifacts, wfv1.Artifact{
			Name: typeInstanceDef.Name,
			Path: "/type-instance",
		})

		template.Outputs.Artifacts = append(template.Outputs.Artifacts, wfv1.Artifact{
			Name:       typeInstanceDef.Name,
			GlobalName: typeInstanceDef.Name,
			Path:       "/type-instance",
		})
	}

	workflow.Templates = append(workflow.Templates, template)

	step.TypeInstances = nil
	step.Template = template.Name
}
