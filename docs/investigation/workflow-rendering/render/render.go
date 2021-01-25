package render

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Knetic/govaluate"
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

	VoltronWhen                *string                  `json:"voltron-when,omitempty"`
	VoltronAction              *string                  `json:"voltron-action,omitempty"`
	VoltronTypeInstanceOutputs []TypeInstanceDefinition `json:"voltron-outputTypeInstances,omitempty"`
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

type mapEvalParameters map[string]interface{}

func (p mapEvalParameters) Get(name string) (interface{}, error) {
	value, found := p[name]
	if !found {
		return nil, nil
	}

	return value, nil
}

var workflowArtifactRefRegex = regexp.MustCompile(`{{workflow\.outputs\.artifacts\.(.+)}}`)

func (r *Renderer) Render(ref v1alpha1.ManifestReference, parameters map[string]interface{}, typeInstances []*v1alpha1.InputTypeInstance, filter RequiresFilter) (*Workflow, error) {
	implementation := r.ManifestStore.GetImplementationForInterface(string(ref.Path), GetImplementationForInterfaceInput{})
	if implementation == nil {
		return nil, fmt.Errorf("implementation for %v not found", ref)
	}

	workflow, _, err := r.renderFunc("", implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while creating root workflow")
	}

	r.RenderedWorkflow = workflow

	parameterRawData, err := yaml.Marshal(parameters)
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal input parameters")
	}

	// Add Action InputParamaters to Workflow.
	r.RenderedWorkflow.Arguments.Artifacts = append(workflow.Arguments.Artifacts, wfv1.Artifact{
		Name: "input-parameters",
		ArtifactLocation: wfv1.ArtifactLocation{
			Raw: &wfv1.RawArtifact{
				Data: string(parameterRawData),
			},
		},
	})

	// Inject TypeInstances to the workflow.
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

	// Rendering iterations
	for {
		// Remove steps, which would create TypeInstances, which are already provided
		if err := r.removeConditionalActionSteps(typeInstances); err != nil {
			return nil, errors.Wrap(err, "while removing conditional action steps")
		}

		// Stop rendering, if there are no more Action steps.
		actionStepRefs := r.findActionSteps()
		if len(actionStepRefs) == 0 {
			break
		}

		for _, tmpl := range r.RenderedWorkflow.Templates {
			artifactMappings := map[string]string{}

			for _, parallelSteps := range tmpl.Steps {
				for i := range parallelSteps {
					step := parallelSteps[i]

					if step.VoltronAction != nil {
						// Get Implementation for action
						actionRef := resolveActionPathFromImports(implementation, *step.VoltronAction)
						if actionRef == "" {
							return nil, errors.Errorf("could not find full path in Implementation imports for action %q", *step.VoltronAction)
						}

						implementation := r.ManifestStore.GetImplementationForInterface(actionRef, GetImplementationForInterfaceInput{RequireFilter: filter})
						if implementation == nil {
							return nil, fmt.Errorf("implementation for %v not found", actionRef)
						}

						// Render the referenced action.
						workflowPrefix := fmt.Sprintf("%s-%s", tmpl.Name, step.Name)
						importedWorkflow, newArtifactMappings, err := r.renderFunc(workflowPrefix, implementation)
						if err != nil {
							return workflow, errors.Wrap(err, "while creating workflow for action step")
						}

						for k, v := range newArtifactMappings {
							artifactMappings[k] = v
						}

						// Include the rendered workflow.
						r.RenderedWorkflow.Templates = append(r.RenderedWorkflow.Templates, importedWorkflow.Templates...)
						step.Template = importedWorkflow.Entrypoint
						step.VoltronAction = nil
					}

					// Replace global artifacts names in references, based on previous gathered mappings.
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

func resolveActionPathFromImports(impl *types.Implementation, voltronAction string) string {
	action := strings.SplitN(voltronAction, ".", 2)
	alias, name := action[0], action[1]
	for _, i := range impl.Spec.Imports {
		if *i.Alias == alias {
			return fmt.Sprintf("%s.%s", i.InterfaceGroupPath, name)
		}
	}
	return ""
}

func (r *Renderer) renderFunc(prefix string,
	implementation *types.Implementation,
) (*Workflow, map[string]string, error) {
	workflow, err := createWorkflow(implementation)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while creating workflow from implementation")
	}

	artifactsNameMapping := map[string]string{}

	for i := range workflow.Templates {
		tmpl := &workflow.Templates[i]

		// Change global artifacts names
		if prefix != "" {
			tmpl.Name = fmt.Sprintf("%s-%s", prefix, tmpl.Name)

			for artIdx := range tmpl.Outputs.Artifacts {
				artifact := &tmpl.Outputs.Artifacts[artIdx]

				if artifact.GlobalName == "" {
					continue
				}

				newName := fmt.Sprintf("%s-%s", prefix, artifact.GlobalName)
				artifactsNameMapping[artifact.GlobalName] = newName
				artifact.GlobalName = newName
			}
		}

		// Add output TypeInstance workflow steps
		for psIdx := range tmpl.Steps {
			parallelSteps := tmpl.Steps[psIdx]

			for sIdx := range parallelSteps {
				step := parallelSteps[sIdx]

				for _, ti := range step.VoltronTypeInstanceOutputs {
					tiStep, template, artifactMappings := r.getOutputTypeInstanceTemplate(step, ti, prefix)
					workflow.Templates = append(workflow.Templates, template)
					tmpl.Steps = append(tmpl.Steps, ParallelSteps{&tiStep})

					for k, v := range artifactMappings {
						artifactsNameMapping[k] = v
					}
				}

				step.VoltronTypeInstanceOutputs = nil
			}
		}
	}

	if prefix != "" {
		workflow.Entrypoint = fmt.Sprintf("%s-%s", prefix, workflow.Entrypoint)
	}

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

func (r *Renderer) removeConditionalActionSteps(instances []*v1alpha1.InputTypeInstance) error {
	for i, tmpl := range r.RenderedWorkflow.Templates {
		newSteps := []ParallelSteps{}

		for _, parallelSteps := range tmpl.Steps {
			newParallelSteps := []*WorkflowStep{}

			for i := range parallelSteps {
				step := parallelSteps[i]

				if step.VoltronWhen != nil {
					if result, err := r.evaluateWhenExpression(instances, *step.VoltronWhen); err != nil {
						return errors.Wrap(err, "while evaluating OCFWhen")
					} else if result == false {
						continue
					}
				}

				step.VoltronWhen = nil

				newParallelSteps = append(newParallelSteps, step)
			}

			if len(newParallelSteps) > 0 {
				newSteps = append(newSteps, newParallelSteps)
			}
		}

		r.RenderedWorkflow.Templates[i].Steps = newSteps
	}

	return nil
}

func (r *Renderer) evaluateWhenExpression(typeInstances []*v1alpha1.InputTypeInstance, exprString string) (interface{}, error) {
	params := mapEvalParameters{}

	for _, ti := range typeInstances {
		params[ti.Name] = ti
	}

	expr, err := govaluate.NewEvaluableExpression(exprString)
	if err != nil {
		return nil, errors.Wrap(err, "while parsing expression")
	}

	result, err := expr.Eval(params)
	if err != nil {
		return nil, errors.Wrap(err, "while evaluating expression")
	}

	return result, nil
}

func (r *Renderer) findActionSteps() []actionStepRef {
	actionStepsRef := []actionStepRef{}

	for _, tmpl := range r.RenderedWorkflow.Templates {
		for _, parallelSteps := range tmpl.Steps {
			for i := range parallelSteps {
				step := parallelSteps[i]
				if step.VoltronAction != nil {
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

func (r *Renderer) getOutputTypeInstanceTemplate(step *WorkflowStep, output TypeInstanceDefinition, prefix string) (WorkflowStep, Template, map[string]string) {
	artifactPath := "/typeinstance"

	stepName := fmt.Sprintf("output-%s", output.Name)

	var templateName string
	var artifactGlobalName string
	if prefix == "" {
		templateName = stepName
		artifactGlobalName = output.Name
	} else {
		templateName = fmt.Sprintf("%s-%s", prefix, stepName)
		artifactGlobalName = fmt.Sprintf("%s-%s", prefix, output.Name)
	}

	fromDirective := fmt.Sprintf("{{steps.%s.outputs.artifacts.%s}}", step.Name, output.From)

	return WorkflowStep{
			Name:     stepName,
			Template: templateName,
			Arguments: wfv1.Arguments{Artifacts: wfv1.Artifacts{
				wfv1.Artifact{
					Name: output.Name,
					From: fromDirective,
				},
			}},
		}, Template{
			Name: templateName,
			Container: &apiv1.Container{
				Image:   "alpine:3.7",
				Command: []string{"sh", "-c"},
				Args:    []string{"sleep 1"},
			},
			Inputs: wfv1.Inputs{
				Artifacts: wfv1.Artifacts{
					{
						Name: output.Name,
						Path: artifactPath,
					},
				},
			},
			Outputs: wfv1.Outputs{
				Artifacts: wfv1.Artifacts{
					{
						Name:       output.Name,
						GlobalName: artifactGlobalName,
						Path:       artifactPath,
					},
				},
			},
		}, map[string]string{
			output.Name: artifactGlobalName,
		}
}
