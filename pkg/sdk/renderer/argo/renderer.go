package argo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"

	"github.com/Knetic/govaluate"
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
)

const userInputName = "input-parameters"

type OCHClient interface {
	GetImplementationForInterface(ctx context.Context, ref ochpublicgraphql.TypeReference) (*ochpublicgraphql.ImplementationRevision, error)
	GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
}

type Renderer struct {
	ochCli              OCHClient
	typeInstanceHandler TypeInstanceHandler
}

func NewRenderer(ochCli OCHClient) *Renderer {
	r := &Renderer{
		ochCli:              ochCli,
		typeInstanceHandler: TypeInstanceHandler{ochCli: ochCli},
	}

	return r
}

func (r *Renderer) Render(ctx context.Context, ref types.InterfaceRef, opts ...RendererOption) (*types.Action, error) {
	// 0. Populate render options
	renderOpt := &renderOptions{}
	for _, opt := range opts {
		opt(renderOpt)
	}

	// 1. Find the root implementation
	implementation, err := r.ochCli.GetImplementationForInterface(ctx, r.refToOCHRef(ref))
	if err != nil {
		return nil, err
	}

	// 2. Extract workflow from the root Implementation
	rootWorkflow, _, err := r.unmarshalWorkflowFromImplementation("", implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while creating root workflow")
	}

	// 3. Add user input if provided
	if err := r.addPlainTextUserInput(rootWorkflow, renderOpt.plainTextUserInput); err != nil {
		return nil, err
	}

	// 4. Add steps to populate rootWorkflow with input TypeInstances
	if err := r.typeInstanceHandler.AddInputTypeInstance(rootWorkflow, renderOpt.inputTypeInstances); err != nil {
		return nil, err
	}

	// 5. Render rootWorkflow templates
	rootWorkflow.Templates, err = r.renderTemplateSteps(ctx, rootWorkflow, implementation.Spec.Imports, renderOpt.inputTypeInstances)
	if err != nil {
		return nil, err
	}

	out, err := r.toMapStringInterface(rootWorkflow)
	if err != nil {
		return nil, err
	}

	runnerInterface := r.resolveRunnerInterface(implementation.Spec.Imports, implementation.Spec.Action.RunnerInterface)

	return &types.Action{
		Args:            out,
		RunnerInterface: runnerInterface,
	}, nil
}

func (r *Renderer) renderTemplateSteps(ctx context.Context, workflow *Workflow, importsCollection []*ochpublicgraphql.ImplementationImport, typeInstances []types.InputTypeInstanceRef) ([]*Template, error) {
	if shouldExit(ctx) {
		return nil, ctx.Err()
	}

	var processedTemplates []*Template

	for _, tpl := range workflow.Templates {
		// 0. Aggregate processed templates
		processedTemplates = append(processedTemplates, tpl)

		artifactMappings := map[string]string{}
		var newStepGroup []ParallelSteps

		for _, parallelSteps := range tpl.Steps {
			var newParallelSteps []*WorkflowStep

			for _, step := range parallelSteps {
				// 1. Skip steps with `voltron-when` statements which are already satisfied
				satisfied, err := r.isStepSatisfiedByInputTypeInstances(step, typeInstances)
				if err != nil {
					return nil, err
				}

				if satisfied {
					continue
				}

				// 2. Import and resolve Implementation for `volton-action`
				if step.VoltronAction != nil {
					// 2.1 Expand `voltron-action` alias based on imports section
					actionRef := r.resolveActionPathFromImports(importsCollection, *step.VoltronAction)
					if actionRef == nil {
						return nil, errors.Errorf("could not find full path in Implementation imports for action %q", *step.VoltronAction)
					}

					// 2.2 Get `voltron-action` specific implementation
					// TODO(policies): take into account polcies
					implementation, err := r.ochCli.GetImplementationForInterface(context.TODO(), *actionRef)
					if err != nil {
						return nil, errors.Wrapf(err, "while processing step: %s", step.Name)
					}

					// 2.3 Extract workflow from the imported `voltron-action`. Prefix it to avoid artifacts name collision.
					workflowPrefix := fmt.Sprintf("%s-%s", tpl.Name, step.Name)
					importedWorkflow, newArtifactMappings, err := r.unmarshalWorkflowFromImplementation(workflowPrefix, implementation)
					if err != nil {
						return nil, errors.Wrap(err, "while creating workflow for action step")
					}

					for k, v := range newArtifactMappings {
						artifactMappings[k] = v
					}
					step.Template = importedWorkflow.Entrypoint
					step.VoltronAction = nil

					// 2.4 Render imported Workflow templates and add them to root templates
					// TODO(advanced-rendering): currently not supported.
					processedImportedTemplates, err := r.renderTemplateSteps(ctx, importedWorkflow, implementation.Spec.Imports, nil)
					processedTemplates = append(processedTemplates, processedImportedTemplates...)
					if err != nil {
						return nil, err
					}
				}

				// 3. Replace global artifacts names in references, based on previous gathered mappings.
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
				newStepGroup = append(newStepGroup, newParallelSteps)
			}
		}

		tpl.Steps = newStepGroup
	}
	return processedTemplates, nil
}

func (r *Renderer) isStepSatisfiedByInputTypeInstances(step *WorkflowStep, typeInstances []types.InputTypeInstanceRef) (bool, error) {
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

// TODO(mszostok): Copied from POC algorithm, replace lib for expression
func (*Renderer) evaluateWhenExpression(typeInstances []types.InputTypeInstanceRef, exprString string) (interface{}, error) {
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

// TODO(mszostok): Change to k8s secret. This is not easy and probably we will need to use some workaround, or
// change the contract.
func (r *Renderer) addPlainTextUserInput(rootWorkflow *Workflow, input map[string]interface{}) error {
	if len(input) == 0 {
		input = map[string]interface{}{}
	}

	parameterRawData, err := yaml.Marshal(input)
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

func (*Renderer) refToOCHRef(in types.InterfaceRef) ochpublicgraphql.TypeReference {
	return ochpublicgraphql.TypeReference{
		Path:     in.Path,
		Revision: stringOrEmpty(in.Revision),
	}
}

// TODO: take into account the runner revision. Respect that also in k8s engine when scheduling runner job
func (r *Renderer) resolveRunnerInterface(imports []*ochpublicgraphql.ImplementationImport, rInterface string) string {
	fullRef := r.resolveActionPathFromImports(imports, rInterface)

	return fullRef.Path
}

func (*Renderer) resolveActionPathFromImports(imports []*ochpublicgraphql.ImplementationImport, voltronAction string) *ochpublicgraphql.TypeReference {
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

func shouldExit(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
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
