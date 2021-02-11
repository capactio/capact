package argo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"projectvoltron.dev/voltron/internal/ptr"

	ochpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/och/client/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"

	"github.com/Knetic/govaluate"
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
)

// dedicatedRenderer is dedicated for rendering a given workflow and it is not thread safe as it stores the
// data in the same way as builder does.
type dedicatedRenderer struct {
	// required vars
	maxDepth            int
	ochCli              OCHClient
	typeInstanceHandler *TypeInstanceHandler

	// set with options
	plainTextUserInput       map[string]interface{}
	inputTypeInstances       []types.InputTypeInstanceRef
	ochImplementationFilters []public.GetImplementationOption

	// internal vars
	currentIteration   int
	processedTemplates []*Template
}

func newDedicatedRenderer(maxDepth int, ochCli OCHClient, typeInstanceHandler *TypeInstanceHandler, opts ...RendererOption) *dedicatedRenderer {
	r := &dedicatedRenderer{
		maxDepth:            maxDepth,
		ochCli:              ochCli,
		typeInstanceHandler: typeInstanceHandler,
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *dedicatedRenderer) AddInputTypeInstance(workflow *Workflow) error {
	return r.typeInstanceHandler.AddInputTypeInstance(workflow, r.inputTypeInstances)
}

func (r *dedicatedRenderer) GetRootTemplates() []*Template {
	return r.processedTemplates
}

func (r *dedicatedRenderer) RenderTemplateSteps(ctx context.Context, workflow *Workflow, importsCollection []*ochpublicapi.ImplementationImport, typeInstances []types.InputTypeInstanceRef) error {
	r.currentIteration++

	if shouldExit(ctx) {
		return ctx.Err()
	}

	if r.maxDepthExceeded() {
		return NewMaxDepthError(r.maxDepth)
	}

	for _, tpl := range workflow.Templates {
		// 0. Aggregate processed templates
		r.addToRootTemplates(tpl)

		artifactMappings := map[string]string{}
		var newStepGroup []ParallelSteps

		for _, parallelSteps := range tpl.Steps {
			var newParallelSteps []*WorkflowStep

			for _, step := range parallelSteps {
				// 1. Skip steps with `voltron-when` statements which are already satisfied
				satisfied, err := r.isStepSatisfiedByInputTypeInstances(step, typeInstances)
				if err != nil {
					return err
				}

				if satisfied {
					continue
				}

				// 2. Import and resolve Implementation for `volton-action`
				if step.VoltronAction != nil {
					// 2.1 Expand `voltron-action` alias based on imports section
					actionRef, err := r.resolveActionPathFromImports(importsCollection, *step.VoltronAction)
					if err != nil {
						return err
					}

					// 2.2 Get `voltron-action` specific implementation
					implementations, err := r.ochCli.GetImplementationRevisionsForInterface(ctx, *actionRef, r.ochImplementationFilters...)
					if err != nil {
						return errors.Wrapf(err, "while processing step: %s", step.Name)
					}

					// business decision select the first one
					implementation := implementations[0]

					// 2.3 Extract workflow from the imported `voltron-action`. Prefix it to avoid artifacts name collision.
					workflowPrefix := fmt.Sprintf("%s-%s", tpl.Name, step.Name)
					importedWorkflow, newArtifactMappings, err := r.UnmarshalWorkflowFromImplementation(workflowPrefix, &implementation)
					if err != nil {
						return errors.Wrap(err, "while creating workflow for action step")
					}

					for k, v := range newArtifactMappings {
						artifactMappings[k] = v
					}
					step.Template = importedWorkflow.Entrypoint
					step.VoltronAction = nil

					// 2.4 Render imported Workflow templates and add them to root templates
					// TODO(advanced-rendering): currently not supported.
					err = r.RenderTemplateSteps(ctx, importedWorkflow, implementation.Spec.Imports, nil)
					if err != nil {
						return err
					}
				}

				// 3. Replace global artifacts names in references, based on previous gathered mappings.
				for artIdx := range step.Arguments.Artifacts {
					art := &step.Arguments.Artifacts[artIdx]

					match := workflowArtifactRefRegex.FindStringSubmatch(art.From)
					if len(match) != 2 {
						continue
					}
					oldArtifactName := match[1]
					if newArtifactName, ok := artifactMappings[oldArtifactName]; ok {
						art.From = fmt.Sprintf("{{workflow.outputs.artifacts.%s}}", newArtifactName)
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
	return nil
}

// TODO: take into account the runner revision. Respect that also in k8s engine when scheduling runner job
func (r *dedicatedRenderer) ResolveRunnerInterface(impl ochpublicapi.ImplementationRevision) (string, error) {
	imports, rInterface := impl.Spec.Imports, impl.Spec.Action.RunnerInterface
	fullRef, err := r.resolveActionPathFromImports(imports, rInterface)
	if err != nil {
		return "", err
	}

	return fullRef.Path, nil
}

func (r *dedicatedRenderer) UnmarshalWorkflowFromImplementation(prefix string, implementation *ochpublicapi.ImplementationRevision) (*Workflow, map[string]string, error) {
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

				if prefix != "" && step.Template != "" {
					step.Template = fmt.Sprintf("%s-%s", prefix, step.Template)
				}

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

// TODO(mszostok): Change to k8s secret. This is not easy and probably we will need to use some workaround, or
// change the contract.
func (r *dedicatedRenderer) AddPlainTextUserInput(rootWorkflow *Workflow) error {
	input := r.plainTextUserInput
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

func (r *dedicatedRenderer) AddRunnerContext(rootWorkflow *Workflow, secretRef RunnerContextSecretRef) error {
	if secretRef.Name == "" || secretRef.Key == "" {
		return NewRunnerContextRefEmptyError()
	}

	idx, err := getEntrypointWorkflowIndex(rootWorkflow)
	if err != nil {
		return err
	}

	container := r.sleepContainer()
	container.VolumeMounts = []apiv1.VolumeMount{
		{
			Name:      runnerContext,
			ReadOnly:  true,
			MountPath: "/input",
		},
	}

	template := &wfv1.Template{
		Name:      fmt.Sprintf("inject-%s", runnerContext),
		Container: container,
		Outputs: wfv1.Outputs{
			Artifacts: wfv1.Artifacts{
				{
					Name:       runnerContext,
					GlobalName: runnerContext,
					Path:       "/input/context.yaml",
				},
			},
		},
		Volumes: []apiv1.Volume{
			{
				Name: runnerContext,
				VolumeSource: apiv1.VolumeSource{
					Secret: &apiv1.SecretVolumeSource{
						SecretName: secretRef.Name,
						Items: []apiv1.KeyToPath{
							{
								Key:  secretRef.Key,
								Path: "context.yaml",
								Mode: nil,
							},
						},
						Optional: ptr.Bool(false),
					},
				},
			},
		},
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

	return nil
}

// Internal helpers

func (*dedicatedRenderer) resolveActionPathFromImports(imports []*ochpublicapi.ImplementationImport, actionRef string) (*ochpublicapi.InterfaceReference, error) {
	action := strings.SplitN(actionRef, ".", 2)
	if len(action) != 2 {
		return nil, NewActionReferencePatternError(actionRef)
	}

	alias, name := action[0], action[1]
	selectFirstMatchedImport := func() *ochpublicapi.InterfaceReference {
		for _, i := range imports {
			if i.Alias == nil || *i.Alias != alias {
				continue
			}
			for _, method := range i.Methods {
				if name != method.Name {
					continue
				}
				return &ochpublicapi.InterfaceReference{
					Path:     fmt.Sprintf("%s.%s", i.InterfaceGroupPath, name),
					Revision: stringOrEmpty(method.Revision),
				}
			}
		}
		return nil
	}

	ref := selectFirstMatchedImport()
	if ref == nil {
		return nil, NewActionImportsError(actionRef)
	}

	return ref, nil
}

func (*dedicatedRenderer) createWorkflow(implementation *ochpublicapi.ImplementationRevision) (*Workflow, error) {
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

func (r *dedicatedRenderer) getOutputTypeInstanceTemplate(step *WorkflowStep, output TypeInstanceDefinition, prefix string) (WorkflowStep, Template, map[string]string) {
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
		Name:      templateName,
		Container: r.sleepContainer(),
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

func (r *dedicatedRenderer) isStepSatisfiedByInputTypeInstances(step *WorkflowStep, typeInstances []types.InputTypeInstanceRef) (bool, error) {
	if step.VoltronWhen == nil {
		return false, nil
	}

	notSatisfied, err := r.evaluateWhenExpression(typeInstances, *step.VoltronWhen)
	if err != nil {
		return false, errors.Wrap(err, "while evaluating OCFWhen")
	}

	// zero value to mark as handled
	step.VoltronWhen = nil

	return notSatisfied == false, nil
}

// TODO(mszostok): Copied from POC algorithm, replace lib for expression
func (*dedicatedRenderer) evaluateWhenExpression(typeInstances []types.InputTypeInstanceRef, exprString string) (interface{}, error) {
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
func (r *dedicatedRenderer) maxDepthExceeded() bool {
	return r.currentIteration > r.maxDepth
}

func (r *dedicatedRenderer) addToRootTemplates(tpl *Template) {
	r.processedTemplates = append(r.processedTemplates, tpl)
}

func (r *dedicatedRenderer) sleepContainer() *apiv1.Container {
	return &apiv1.Container{
		Image:   "alpine:3.7",
		Command: []string{"sh", "-c"},
		Args:    []string{"sleep 1"},
	}
}
