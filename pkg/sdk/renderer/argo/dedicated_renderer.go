package argo

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"projectvoltron.dev/voltron/internal/ptr"

	ochpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"

	"github.com/Knetic/govaluate"
	wfv1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
)

// dedicatedRenderer is dedicated for rendering a given workflow and it is not thread safe as it stores the
// data in the same way as builder does.
type dedicatedRenderer struct {
	// required vars
	maxDepth            int
	policyEnforcedCli   PolicyEnforcedOCHClient
	typeInstanceHandler *TypeInstanceHandler

	// set with options
	userInputSecretRef *UserInputSecretRef
	inputTypeInstances []types.InputTypeInstanceRef

	// internal vars
	currentIteration   int
	processedTemplates []*Template
	rootTemplate       *Template
	entrypointStep     *wfv1.WorkflowStep
	tplInputArguments  map[string]wfv1.Artifacts

	outputTypeInstances     *OutputTypeInstances
	outputTypeInstanceNames map[string]struct{}
}

func newDedicatedRenderer(maxDepth int, policyEnforcedCli PolicyEnforcedOCHClient, typeInstanceHandler *TypeInstanceHandler, opts ...RendererOption) *dedicatedRenderer {
	r := &dedicatedRenderer{
		maxDepth:            maxDepth,
		policyEnforcedCli:   policyEnforcedCli,
		typeInstanceHandler: typeInstanceHandler,
		tplInputArguments:   map[string]wfv1.Artifacts{},

		outputTypeInstances: &OutputTypeInstances{
			typeInstances: []OutputTypeInstance{},
			relations:     []OutputTypeInstanceRelation{},
		},
		outputTypeInstanceNames: map[string]struct{}{},
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *dedicatedRenderer) WrapEntrypointWithRootStep(workflow *Workflow) *Workflow {
	r.entrypointStep = &wfv1.WorkflowStep{
		Name:     "start-entrypoint",
		Template: workflow.Entrypoint, // previous entry point
	}

	r.rootTemplate = &Template{
		Template: &wfv1.Template{
			Name: "voltron-root",
		},
		Steps: []ParallelSteps{
			{
				{
					WorkflowStep: r.entrypointStep,
				},
			},
		},
	}

	workflow.Entrypoint = r.rootTemplate.Name
	workflow.Templates = append(workflow.Templates, r.rootTemplate)

	return workflow
}

func (r *dedicatedRenderer) AddInputTypeInstance(workflow *Workflow) error {
	return r.typeInstanceHandler.AddInputTypeInstance(workflow, r.inputTypeInstances)
}

func (r *dedicatedRenderer) AddOutputTypeInstancesStep(workflow *Workflow) error {
	return r.typeInstanceHandler.AddUploadTypeInstancesStep(workflow, r.outputTypeInstances)
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
				// 1. Register step arguments, so we can process them in referenced template and check
				// whether steps in referenced template are satisfied
				r.registerTemplateInputArguments(step)

				// 2. Check step with `voltron-when` statements if it can be satisfied by input arguments
				satisfiedArg, err := r.getInputArgWhichSatisfyStep(tpl.Name, step)
				if err != nil {
					return err
				}

				// 2.1 Replace step and emit input arguments as step output
				if satisfiedArg != "" {
					emitStep, wfTpl := r.emitWorkflowInputAsStepOutput(tpl.Name, step, satisfiedArg)
					step = emitStep
					r.addToRootTemplates(wfTpl)
				}

				// 3. Import and resolve Implementation for `volton-action`
				if step.VoltronAction != nil {
					// 3.1 Expand `voltron-action` alias based on imports section
					actionRef, err := r.resolveActionPathFromImports(importsCollection, *step.VoltronAction)
					if err != nil {
						return err
					}

					// 3.2 Get `voltron-action` Interface
					iface, err := r.policyEnforcedCli.GetInterfaceRevision(ctx, *actionRef)
					if err != nil {
						return err
					}

					// 3.3 Get `voltron-action` specific implementation
					implementations, rule, err := r.policyEnforcedCli.ListImplementationRevisionForInterface(ctx, *actionRef)
					if err != nil {
						return errors.Wrapf(err, "while processing step: %s", step.Name)
					}

					// business decision select the first one
					implementation := implementations[0]

					// 3.4 Get TypeInstances to inject based on policy
					typeInstances, err := r.policyEnforcedCli.ListTypeInstancesToInjectBasedOnPolicy(ctx, rule, implementation)
					if err != nil {
						return errors.Wrapf(err, "while listing TypeInstances to inject based on policy for step: %s", step.Name)
					}

					// 3.5 Inject step which downloads TypeInstances
					err = r.InjectDownloadStepForTypeInstancesIfProvided(workflow, typeInstances)
					if err != nil {
						return errors.Wrapf(err, "while injecting step downloading TypeInstances based on policy for step: %s", step.Name)
					}

					workflowPrefix := fmt.Sprintf("%s-%s", tpl.Name, step.Name)

					// 3.6 Prefix output TypeInstances in the manifests
					r.prefixOutputTypeInstancesInManifests(step, workflowPrefix, &implementation, iface)

					// 3.7 Extract workflow from the imported `voltron-action`. Prefix it to avoid artifacts name collision.
					importedWorkflow, newArtifactMappings, err := r.UnmarshalWorkflowFromImplementation(workflowPrefix, &implementation)
					if err != nil {
						return errors.Wrap(err, "while creating workflow for action step")
					}

					for k, v := range newArtifactMappings {
						artifactMappings[k] = v
					}

					if err := r.registerOutputTypeInstances(iface, &implementation); err != nil {
						return errors.Wrap(err, "while noting output artifacts")
					}

					step.Template = importedWorkflow.Entrypoint
					step.VoltronAction = nil

					// 3.8 Right now we know the template name, so let's try to register step input arguments
					r.registerTemplateInputArguments(step)

					// 3.9 Render imported Workflow templates and add them to root templates
					// TODO(advanced-rendering): currently not supported.
					err = r.RenderTemplateSteps(ctx, importedWorkflow, implementation.Spec.Imports, nil)
					if err != nil {
						return err
					}
				}

				step.VoltronTypeInstanceOutputs = nil

				// 4. Replace global artifacts names in references, based on previous gathered mappings.
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

				for i := range step.VoltronTypeInstanceOutputs {
					ti := &step.VoltronTypeInstanceOutputs[i]

					if prefix != "" {
						ti.Name = fmt.Sprintf("%s-%s", prefix, ti.Name)
					}

					tiStep, template, artifactMappings := r.getOutputTypeInstanceTemplate(step, *ti, prefix)
					workflow.Templates = append(workflow.Templates, &template)
					tmpl.Steps = append(tmpl.Steps, ParallelSteps{&tiStep})

					for k, v := range artifactMappings {
						artifactsNameMapping[k] = v
					}
				}
			}
		}
	}

	if prefix != "" {
		workflow.Entrypoint = fmt.Sprintf("%s-%s", prefix, workflow.Entrypoint)
	}

	return workflow, artifactsNameMapping, nil
}

func (r *dedicatedRenderer) AddUserInputSecretRefIfProvided(rootWorkflow *Workflow) {
	if r.userInputSecretRef == nil {
		return
	}

	var (
		volumeName   = "user-secret-volume"
		mountPath    = "/output"
		artifactPath = fmt.Sprintf("%s/%s", mountPath, r.userInputSecretRef.Key)
	)

	// 1. Create step which consumes user data from Secret and outputs it as artifact
	container := r.sleepContainer()
	container.VolumeMounts = []apiv1.VolumeMount{
		{
			Name:      volumeName,
			MountPath: mountPath,
		},
	}

	userInputWfTpl := &wfv1.Template{
		Name:      "populate-user-input",
		Container: container,
		Outputs: wfv1.Outputs{
			Artifacts: wfv1.Artifacts{
				{
					Name: userInputName,
					Path: artifactPath,
				},
			},
		},
		Volumes: []apiv1.Volume{
			{
				Name: volumeName,
				VolumeSource: apiv1.VolumeSource{
					Secret: &apiv1.SecretVolumeSource{
						SecretName: r.userInputSecretRef.Name,
						Items: []apiv1.KeyToPath{
							{
								Key:  r.userInputSecretRef.Key,
								Path: r.userInputSecretRef.Key,
							},
						},
						Optional: ptr.Bool(false),
					},
				},
			},
		},
	}
	userInputWfStep := &wfv1.WorkflowStep{
		Name:     fmt.Sprintf("%s-step", userInputWfTpl.Name),
		Template: userInputWfTpl.Name,
	}
	r.rootTemplate.Steps = append([]ParallelSteps{
		{
			&WorkflowStep{
				WorkflowStep: userInputWfStep,
			},
		},
	}, r.rootTemplate.Steps...)
	rootWorkflow.Templates = append(rootWorkflow.Templates, &Template{Template: userInputWfTpl})

	// 2. Add input arguments artifacts with user data. Thanks to that Content Developer can
	// refer to it via "{{inputs.artifacts.input-parameters}}"
	r.entrypointStep.Arguments.Artifacts = append(r.entrypointStep.Arguments.Artifacts, wfv1.Artifact{
		Name: userInputName,
		From: fmt.Sprintf("{{steps.%s.outputs.artifacts.%s}}", userInputWfStep.Name, userInputName),
	})
}

func (r *dedicatedRenderer) AddRunnerContext(rootWorkflow *Workflow, secretRef RunnerContextSecretRef) error {
	if secretRef.Name == "" || secretRef.Key == "" {
		return NewRunnerContextRefEmptyError()
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

	r.rootTemplate.Steps = append([]ParallelSteps{
		{
			&WorkflowStep{
				WorkflowStep: &wfv1.WorkflowStep{
					Name:     fmt.Sprintf("%s-step", template.Name),
					Template: template.Name,
				},
			},
		},
	}, r.rootTemplate.Steps...)

	rootWorkflow.Templates = append(rootWorkflow.Templates, &Template{Template: template})

	return nil
}

// TODO: Rework if needed as a part of SV-185
func (r *dedicatedRenderer) InjectDownloadStepForTypeInstancesIfProvided(workflow *Workflow, typeInstances []types.InputTypeInstanceRef) error {
	if len(typeInstances) == 0 {
		return nil
	}
	return r.typeInstanceHandler.AddInputTypeInstance(workflow, typeInstances)
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

func (r *dedicatedRenderer) prefixOutputTypeInstancesInManifests(step *WorkflowStep, prefix string,
	impl *ochpublicapi.ImplementationRevision, iface *ochpublicapi.InterfaceRevision) {
	if step.VoltronTypeInstanceOutputs != nil {
		for _, output := range step.VoltronTypeInstanceOutputs {
			for i := range impl.Spec.OutputTypeInstanceRelations {
				ti := impl.Spec.OutputTypeInstanceRelations[i]
				if ti.TypeInstanceName == output.From {
					ti.TypeInstanceName = output.Name

					for usesIdx := range ti.Uses {
						ti.Uses[usesIdx] = fmt.Sprintf("%s-%s", prefix, ti.Uses[usesIdx])
					}
				}
			}
		}
	}

	if impl.Spec.AdditionalOutput != nil && impl.Spec.AdditionalOutput.TypeInstances != nil {
		for i := range impl.Spec.AdditionalOutput.TypeInstances {
			ti := impl.Spec.AdditionalOutput.TypeInstances[i]
			ti.Name = fmt.Sprintf("%s-%s", prefix, ti.Name)
		}
	}

	if iface.Spec.Output != nil && iface.Spec.Output.TypeInstances != nil {
		for i := range iface.Spec.Output.TypeInstances {
			ti := iface.Spec.Output.TypeInstances[i]
			ti.Name = fmt.Sprintf("%s-%s", prefix, ti.Name)
		}
	}
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
		artifactGlobalName = output.Name
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

//  This function checks if a given step is satisfied by input arguments
//
//  Example:
//
//   - name: stack-install
//     steps:
//     - - name: entrypoint										# Step which execute template with arguments.
//         template: jira-install								# We record input arguments under template name.
//         arguments:
//           artifacts:
//             - name: postgresql
//               from: "{{steps.install-db.outputs.artifacts.postgresql}}"
//   - name: jira-install
//     inputs:
//      artifacts:
//        - name: input-parameters
//        - name: postgresql
//          optional: true
//    steps:
//      - - voltron-when: postgresql == nil						# Check whether this step is satisfied by input arguments.
//          name: install-db									# For that we need to have option to check which arguments were passed
//																# to this step.
func (r *dedicatedRenderer) getInputArgWhichSatisfyStep(tplOwnerName string, step *WorkflowStep) (string, error) {
	if step.VoltronWhen == nil {
		return "", nil
	}

	args, found := r.tplInputArguments[tplOwnerName]
	if !found {
		// zero value to mark as handled
		step.VoltronWhen = nil
		return "", nil
	}

	params := &mapEvalParameters{}
	for _, a := range args {
		params.Set(a.Name)
	}

	notSatisfied, err := r.evaluateWhenExpression(params, *step.VoltronWhen)
	if err != nil {
		return "", errors.Wrap(err, "while evaluating OCFWhen")
	}

	// zero value to mark as handled
	step.VoltronWhen = nil

	if notSatisfied == true {
		return "", nil
	}

	return params.lastAccessed, nil
}

// TODO:
//   We can change lib to `github.com/antonmedv/expr` and create our own functions which will allow us to introspect which artifact satisfied a given step:
//    - isDefined(foo,bar,baz)
//    - isNotDefined(foo,bar,baz)
//    - isDefined(foo,bar,baz) && isNotDefined(foo,bar,baz)
//    - isDefined(foo,bar,baz) || isNotDefined(foo,bar,baz)
func (*dedicatedRenderer) evaluateWhenExpression(params *mapEvalParameters, exprString string) (interface{}, error) {
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

// TODO: current limitation: we handle properly only one artifacts `voltron-when: postgres == nil` but not `voltron-when: postgres == nil && jira-config == nil`
func (r *dedicatedRenderer) emitWorkflowInputAsStepOutput(tplName string, step *WorkflowStep, inputArgName string) (*WorkflowStep, *Template) {
	var artifactPath = fmt.Sprintf("output/%s", inputArgName)

	// 1. Create step which outputs workflow input argument as step artifact
	userInputWfTpl := &wfv1.Template{
		Name:      fmt.Sprintf("mock-%s-%s", tplName, step.Name),
		Container: r.sleepContainer(),
		Outputs: wfv1.Outputs{
			Artifacts: wfv1.Artifacts{
				{
					Name: inputArgName,
					Path: artifactPath,
				},
			},
		},
		Inputs: wfv1.Inputs{
			Artifacts: wfv1.Artifacts{
				{
					Name:     inputArgName,
					Optional: false,
					Path:     artifactPath,
				},
			},
		},
	}

	userInputWfStep := &wfv1.WorkflowStep{
		Name:     step.Name,
		Template: userInputWfTpl.Name,
		Arguments: wfv1.Arguments{
			Artifacts: wfv1.Artifacts{
				{
					Name: inputArgName,
					From: fmt.Sprintf("{{inputs.artifacts.%s}}", inputArgName),
				},
			},
		},
	}
	return &WorkflowStep{WorkflowStep: userInputWfStep}, &Template{Template: userInputWfTpl}
}

func (r *dedicatedRenderer) registerTemplateInputArguments(step *WorkflowStep) {
	if step.GetTemplateName() == "" || len(step.Arguments.Artifacts) == 0 {
		return
	}
	r.tplInputArguments[step.Template] = step.Arguments.Artifacts
}

func (r *dedicatedRenderer) registerOutputTypeInstances(iface *ochpublicapi.InterfaceRevision, impl *ochpublicapi.ImplementationRevision) error {
	for _, item := range impl.Spec.OutputTypeInstanceRelations {
		artifactName, isNew := r.addTypeInstanceName(item.TypeInstanceName)

		if isNew {
			typeRef, err := findTypeInstanceTypeRef(item.TypeInstanceName, impl, iface)
			if err != nil {
				return err
			}

			r.outputTypeInstances.typeInstances = append(r.outputTypeInstances.typeInstances, OutputTypeInstance{
				ArtifactName: artifactName,
				TypeInstance: types.OutputTypeInstance{
					TypeRef: &types.TypeRef{
						Path:     typeRef.Path,
						Revision: &typeRef.Revision,
					},
				},
			})
		}

		for _, uses := range item.Uses {
			usesArtifactName, isNew := r.addTypeInstanceName(uses)

			r.outputTypeInstances.relations = append(r.outputTypeInstances.relations, OutputTypeInstanceRelation{
				From: artifactName,
				To:   usesArtifactName,
			})

			if isNew {
				typeRef, err := findTypeInstanceTypeRef(uses, impl, iface)
				if err != nil {
					return NewTypeReferenceNotFoundError(uses)
				}

				r.outputTypeInstances.typeInstances = append(r.outputTypeInstances.typeInstances, OutputTypeInstance{
					ArtifactName: usesArtifactName,
					TypeInstance: types.OutputTypeInstance{
						TypeRef: &types.TypeRef{
							Path:     typeRef.Path,
							Revision: &typeRef.Revision,
						},
					},
				})
			}
		}
	}

	return nil
}

func (r *dedicatedRenderer) addTypeInstanceName(name string) (*string, bool) {
	_, found := r.outputTypeInstanceNames[name]
	if !found {
		r.outputTypeInstanceNames[name] = struct{}{}
	}

	return &name, !found
}
