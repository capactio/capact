package argo

import (
	"context"
	"encoding/json"
	"fmt"
	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	"strings"

	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"

	"github.com/Knetic/govaluate"
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
)

type OCHClient interface {
	GetImplementationForInterface(ctx context.Context, ref ochpublicgraphql.TypeReference) (*ochpublicgraphql.ImplementationRevision, error)
	GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
}

type Renderer struct {
	ochCli        OCHClient
	rootTemplates []*Template
}

func NewRenderer(ochCli OCHClient) *Renderer {
	return &Renderer{ochCli: ochCli}
}

// what about policies (should be passed here? if so, func opt?)
// renderer + och + polices ... = facade ?
// argoRenderer + ocfImplGetter + policiesFilters(knows engine + och filters)
//func (r *Renderer) Render(ctx context.Context, ocfImpl types.Implementation) (rendered []byte) {
//
//}
// I decided to do not overcomplicate and make it more Argo specific

type RendererOption func(workflow *Workflow)

const userInputName = "input-parameters"

func (*Renderer) getEntrypointWorkflowIndex(w *Workflow) (int, bool) {
	if w == nil {
		return 0, false
	}
	for idx, tmpl := range w.Templates {
		if tmpl.Name == w.Entrypoint {
			return idx, true
		}
	}

	return 0, false
}

func (r *Renderer) refToOCHRef(in types.TypeRef) ochpublicgraphql.TypeReference {
	return ochpublicgraphql.TypeReference{
		Path:     in.Path,
		Revision: stringOrEmpty(in.Revision),
	}
}

func (r *Renderer) Render(ref types.TypeRef, parameters map[string]interface{}, typeInstances []v1alpha1.InputTypeInstance) (*types.Action, error) {
	// 1. Find the root implementation
	implementation, err := r.ochCli.GetImplementationForInterface(context.TODO(), r.refToOCHRef(ref))
	if err != nil {
		return nil, err
	}

	// 2. Extract workflow from the root Implementation
	rootWorkflow, _, err := r.unmarshalWorkflowFromImplementation("", implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while creating root workflow")
	}

	// 3. Add user input if provided
	if err := r.addUserInput(rootWorkflow, parameters); err != nil {
		return nil, err
	}

	// 4. Add steps to populate rootWorkflow with input TypeInstances
	if err := r.addInputTypeInstance(rootWorkflow, typeInstances); err != nil {
		return nil, err
	}

	// 5. Render rootWorkflow templates
	err = r.renderFullWorkflowTemplates(rootWorkflow, implementation.Spec.Imports, typeInstances)
	if err != nil {
		return nil, err
	}

	rootWorkflow.Templates = r.rootTemplates
	out, err := r.toMapStringInterface(rootWorkflow)
	if err != nil {
		return nil, err
	}
	return &types.Action{
		Args:            out,
		RunnerInterface: implementation.Spec.Action.RunnerInterface,
	}, nil
}

// TODO(SV-189): Handle that properly
func (r *Renderer) addInputTypeInstance(rootWorkflow *Workflow, typeInstances []v1alpha1.InputTypeInstance) error {
	idx, found := r.getEntrypointWorkflowIndex(rootWorkflow)
	if !found {
		return errors.Errorf("cannot find workflow index specified by entrypoint %q", rootWorkflow.Entrypoint)
	}

	for _, tiInput := range typeInstances {
		template, err := r.getInjectTypeInstanceTemplate(tiInput)
		if err != nil {
			return errors.Wrapf(err, "while getting inject TypeInstance template for %s", tiInput.ID)
		}

		rootWorkflow.Templates[idx].Steps = append([]ParallelSteps{
			{
				&WorkflowStep{
					WorkflowStep: &wfv1.WorkflowStep{
						Name:     fmt.Sprintf("%s-step", template.Name),
						Template: template.Name,
					},
				},
			},
		}, rootWorkflow.Templates[idx].Steps...)

		rootWorkflow.Templates = append(rootWorkflow.Templates, &Template{Template: template})
	}

	return nil
}

// TODO(mszostok): Change to k8s secret. This is not easy and probably we will need to use some workaround, or
// change the contract.
func (r *Renderer) addUserInput(rootWorkflow *Workflow, parameters map[string]interface{}) error {
	if len(parameters) == 0 {
		return nil
	}

	parameterRawData, err := yaml.Marshal(parameters)
	if err != nil {
		return errors.Wrap(err, "failed to marshal input parameters")
	}

	rootWorkflow.Arguments.Artifacts = append(rootWorkflow.Arguments.Artifacts, wfv1.Artifact{
		Name: userInputName,
		ArtifactLocation: wfv1.ArtifactLocation{
			Raw: &wfv1.RawArtifact{
				Data: string(parameterRawData),
			},
		},
	})

	return nil
}

func (r *Renderer) renderFullWorkflowTemplates(workflow *Workflow, importsCollection []*ochpublicgraphql.ImplementationImport, typeInstances []v1alpha1.InputTypeInstance) error {
	for idx := range workflow.Templates {
		tpl := workflow.Templates[idx]

		r.rootTemplates = append(r.rootTemplates, tpl)
		err := r.renderTemplateStepsInPlace(tpl, importsCollection, typeInstances)
		if err != nil {
			return err
		}

	}
	return nil
}

func (r *Renderer) isStepSatisfiedByInputTypeInstances(step *WorkflowStep, typeInstances []v1alpha1.InputTypeInstance) (bool, error) {
	if step.VoltronWhen != nil {
		result, err := r.evaluateWhenExpression(typeInstances, *step.VoltronWhen)
		if err != nil {
			return false, errors.Wrap(err, "while evaluating OCFWhen")
		}

		if result == false { // continue as already satisfied and now need to resolve it
			return true, nil
		}

		// zero value to mark as handled
		step.VoltronWhen = nil
	}
	return false, nil
}

func (r *Renderer) renderTemplateStepsInPlace(tmpl *Template, importsCollection []*ochpublicgraphql.ImplementationImport, typeInstances []v1alpha1.InputTypeInstance) error {
	artifactMappings := map[string]string{}
	var newSteps []ParallelSteps

	for _, parallelSteps := range tmpl.Steps {
		var newParallelSteps []*WorkflowStep
		for i := range parallelSteps {
			step := parallelSteps[i]

			satisfied, err := r.isStepSatisfiedByInputTypeInstances(step, typeInstances)
			if err != nil {
				return err
			}

			if satisfied {
				continue
			}

			if step.VoltronAction != nil {
				// Get Implementation for action
				actionRef := resolveActionPathFromImports(importsCollection, *step.VoltronAction)
				if actionRef == nil {
					return errors.Errorf("could not find full path in Implementation imports for action %q", *step.VoltronAction)
				}

				implementation, err := r.ochCli.GetImplementationForInterface(context.TODO(), *actionRef)
				if err != nil {
					return errors.Wrapf(err, "while processing step: %s", step.Name)
				}

				// Render the referenced action.
				workflowPrefix := fmt.Sprintf("%s-%s", tmpl.Name, step.Name)
				importedWorkflow, newArtifactMappings, err := r.unmarshalWorkflowFromImplementation(workflowPrefix, implementation)
				if err != nil {
					return errors.Wrap(err, "while creating workflow for action step")
				}

				for k, v := range newArtifactMappings {
					artifactMappings[k] = v
				}
				step.Template = importedWorkflow.Entrypoint
				step.VoltronAction = nil

				// TODO(mszostok): support advanced rendering? maybe instead of recursion we can create a bucket per layer
				// save imported workflow and execute (renderFullWorkflowTemplates with user data)
				err = r.renderFullWorkflowTemplates(importedWorkflow, implementation.Spec.Imports, nil)
				if err != nil {
					return err
				}
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
			newParallelSteps = append(newParallelSteps, step)
		}
		if len(newParallelSteps) > 0 {
			newSteps = append(newSteps, newParallelSteps)
		}
	}

	tmpl.Steps = newSteps
	return nil
}

func resolveActionPathFromImports(imports []*ochpublicgraphql.ImplementationImport, voltronAction string) *ochpublicgraphql.TypeReference {
	action := strings.SplitN(voltronAction, ".", 2)
	alias, name := action[0], action[1]
	for _, i := range imports {
		if *i.Alias == alias {
			return &ochpublicgraphql.TypeReference{
				Path:     fmt.Sprintf("%s.%s", i.InterfaceGroupPath, name),
				Revision: stringOrEmpty(i.AppVersion),
			}
		}
	}
	return nil
}

func stringOrEmpty(in *string) string {
	if in != nil {
		return *in
	}
	return ""
}

func (r *Renderer) unmarshalWorkflowFromImplementation(prefix string, implementation *ochpublicgraphql.ImplementationRevision) (*Workflow, map[string]string, error) {
	workflow, err := r.createWorkflow(implementation)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while unmarshaling Argo Workflow from OCF Implementation")
	}

	artifactsNameMapping := map[string]string{}

	for i := range workflow.Templates {
		tmpl := workflow.Templates[i]

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
					workflow.Templates = append(workflow.Templates, &template)
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

func (r *Renderer) toMapStringInterface(w *Workflow) (map[string]interface{}, error) {
	var renderedWorkflow = struct {
		Spec Workflow `json:"workflow"`
	}{
		Spec: *w,
	}
	out := map[string]interface{}{}
	marshaled, err := json.Marshal(renderedWorkflow)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(marshaled, &out); err != nil {
		return nil, err
	}

	return out, nil
}

func (*Renderer) createWorkflow(implementation *ochpublicgraphql.ImplementationRevision) (*Workflow, error) {
	var renderedWorkflow = struct {
		Spec Workflow `json:"workflow"`
	}{}

	b, err := json.Marshal(implementation.Spec.Action.Args)
	if err != nil {
		return nil, errors.Wrap(err, "while marshaling Implementation workflow")
	}
	if err := json.Unmarshal(b, &renderedWorkflow); err != nil {
		return nil, errors.Wrap(err, "while unmarshaling to spec")
	}
	return &renderedWorkflow.Spec, nil
}

// TODO(mszostok): Copied from POC algorithm, replace lib for expression
func (r *Renderer) evaluateWhenExpression(typeInstances []v1alpha1.InputTypeInstance, exprString string) (interface{}, error) {
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

func (r *Renderer) getInjectTypeInstanceTemplate(input v1alpha1.InputTypeInstance) (*wfv1.Template, error) {
	typeInstance, err := r.ochCli.GetTypeInstance(context.TODO(), input.ID)
	if err != nil {
		return nil, err
	}
	if typeInstance == nil {
		return nil, fmt.Errorf("failed to find TypeInstance %s", input.ID)
	}

	data, err := yaml.Marshal(typeInstance.Spec.Value)
	if err != nil {
		return nil, errors.Wrap(err, "while to marshal TypeInstance to YAML")
	}

	return &wfv1.Template{
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

	argoWfStep := &wfv1.WorkflowStep{
		Name:     stepName,
		Template: templateName,
		Arguments: wfv1.Arguments{Artifacts: wfv1.Artifacts{
			wfv1.Artifact{
				Name: output.Name,
				From: fromDirective,
			},
		}},
	}
	argoWfTemplate := &wfv1.Template{
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
	}
	return WorkflowStep{WorkflowStep: argoWfStep}, Template{Template: argoWfTemplate}, map[string]string{
		output.Name: artifactGlobalName,
	}
}
