package argo

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strings"

	"capact.io/capact/internal/ctxutil"
	"capact.io/capact/internal/k8s-engine/graphql/domain/action"

	"capact.io/capact/internal/ptr"
	hubpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"

	"github.com/Knetic/govaluate"
	wfv1 "github.com/argoproj/argo-workflows/v3/pkg/apis/workflow/v1alpha1"
	"github.com/pkg/errors"
	apiv1 "k8s.io/api/core/v1"
)

// dedicatedRenderer is dedicated for rendering a given workflow and it is not thread safe as it stores the
// data in the same way as builder does.
type dedicatedRenderer struct {
	// required vars
	maxDepth            int
	policyEnforcedCli   PolicyEnforcedHubClient
	typeInstanceHandler *TypeInstanceHandler

	// set with options
	inputParametersSecretRef  *UserInputSecretRef
	inputParametersCollection types.ParametersCollection
	inputTypeInstances        []types.InputTypeInstanceRef
	ownerID                   *string

	// internal vars
	currentIteration   int
	processedTemplates []*Template
	rootTemplate       *Template
	entrypointStep     *WorkflowStep
	tplInputArguments  map[string][]InputArtifact

	typeInstancesToOutput             *OutputTypeInstances
	typeInstancesToUpdate             UpdateTypeInstances
	registeredOutputTypeInstanceNames []*string
}

// InputArtifact is an Argo artifact with a reference to a Capact TypeInstance.
// It is used to track the TypeInstance, which is handled in the workflow.
type InputArtifact struct {
	artifact              wfv1.Artifact
	typeInstanceReference *string
}

func newDedicatedRenderer(maxDepth int, policyEnforcedCli PolicyEnforcedHubClient, typeInstanceHandler *TypeInstanceHandler, opts ...RendererOption) *dedicatedRenderer {
	r := &dedicatedRenderer{
		maxDepth:            maxDepth,
		policyEnforcedCli:   policyEnforcedCli,
		typeInstanceHandler: typeInstanceHandler,
		tplInputArguments:   map[string][]InputArtifact{},

		typeInstancesToOutput: &OutputTypeInstances{
			typeInstances: []OutputTypeInstance{},
			relations:     []OutputTypeInstanceRelation{},
		},
		typeInstancesToUpdate:             UpdateTypeInstances{},
		registeredOutputTypeInstanceNames: []*string{},
	}

	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *dedicatedRenderer) WrapEntrypointWithRootStep(workflow *Workflow) (*Workflow, *WorkflowStep) {
	r.entrypointStep = &WorkflowStep{
		WorkflowStep: &wfv1.WorkflowStep{
			Name:     "start-entrypoint",
			Template: workflow.Entrypoint, // previous entry point
		}}

	r.rootTemplate = &Template{
		Template: &wfv1.Template{
			Name: "capact-root",
		},
		Steps: []ParallelSteps{
			{
				{
					WorkflowStep: r.entrypointStep.WorkflowStep,
				},
			},
		},
	}

	workflow.Entrypoint = r.rootTemplate.Name
	workflow.Templates = append(workflow.Templates, r.rootTemplate)

	return workflow, r.entrypointStep
}

func (r *dedicatedRenderer) AppendAdditionalInputTypeInstances(typeInstances []types.InputTypeInstanceRef) {
	if len(typeInstances) == 0 {
		return
	}

	r.inputTypeInstances = append(r.inputTypeInstances, typeInstances...)
}

func (r *dedicatedRenderer) AddInputTypeInstances(workflow *Workflow) error {
	availableTypeInstances := map[argoArtifactRef]*string{}

	for _, ti := range r.inputTypeInstances {
		r.entrypointStep.Arguments.Artifacts = append(r.entrypointStep.Arguments.Artifacts, wfv1.Artifact{
			Name: ti.Name,
			From: fmt.Sprintf("{{workflow.outputs.artifacts.%s}}", ti.Name),
		})

		tiPtr := r.addTypeInstanceName(ti.ID)
		availableTypeInstances[argoArtifactRef{
			step: ArgoArtifactNoStep,
			name: ti.Name,
		}] = tiPtr
	}

	r.registerTemplateInputArguments(r.entrypointStep, availableTypeInstances)

	return r.typeInstanceHandler.AddInputTypeInstances(workflow, r.inputTypeInstances)
}

func (r *dedicatedRenderer) AddOutputTypeInstancesStep(workflow *Workflow) error {
	if r.ownerID == nil {
		return NewMissingOwnerIDError()
	}

	if len(r.typeInstancesToOutput.relations) != 0 || len(r.typeInstancesToOutput.typeInstances) != 0 {
		if err := r.typeInstanceHandler.AddUploadTypeInstancesStep(workflow, r.typeInstancesToOutput, *r.ownerID); err != nil {
			return err
		}
	}

	if len(r.typeInstancesToUpdate) > 0 {
		if err := r.typeInstanceHandler.AddUpdateTypeInstancesStep(workflow, r.typeInstancesToUpdate, *r.ownerID); err != nil {
			return err
		}
	}

	return nil
}

func (r *dedicatedRenderer) GetTypeInstancesToLock() []string {
	var typeInstances []string
	for _, ti := range r.typeInstancesToUpdate {
		typeInstances = append(typeInstances, ti.ID)
	}

	return typeInstances
}

func (r *dedicatedRenderer) GetRootTemplates() []*Template {
	return r.processedTemplates
}

// TODO Refactor it. It's too long
// 1. Split it to smaller functions and leave only high level steps here
// 2. Do not use global state, calling it multiple times seems not to work
//nolint:gocyclo // This legacy function is complex but the team too busy to simplify it
func (r *dedicatedRenderer) RenderTemplateSteps(ctx context.Context, workflow *Workflow, importsCollection []*hubpublicapi.ImplementationImport,
	typeInstances []types.InputTypeInstanceRef, prefix string) (map[string]*string, error) {
	r.currentIteration++

	if ctxutil.ShouldExit(ctx) {
		return nil, ctx.Err()
	}

	if r.maxDepthExceeded() {
		return nil, NewMaxDepthError(r.maxDepth)
	}

	outputTypeInstances := map[string]*string{}

	for _, tpl := range workflow.Templates {
		// 0. Aggregate processed templates
		r.addToRootTemplates(tpl)

		// 0.1. Get TypeInstances available for this template
		availableTypeInstances := getAvailableTypeInstancesFromInputArtifacts(r.tplInputArguments[tpl.Name])

		artifactMappings := map[string]string{}
		var newStepGroup []ParallelSteps

		for _, parallelSteps := range tpl.Steps {
			var newParallelSteps []*WorkflowStep

			for _, step := range parallelSteps {
				// 1. Register step arguments, so we can process them in referenced template and check
				// whether steps in referenced template are satisfied
				r.registerTemplateInputArguments(step, availableTypeInstances)

				// 2. Check step with `capact-when` statements if it can be satisfied by input arguments
				satisfiedArg, err := r.getInputArgWhichSatisfyStep(tpl.Name, step)
				if err != nil {
					return nil, err
				}

				// 2.1 Replace step and emit input arguments as step output
				if satisfiedArg != "" {
					emitStep, wfTpl := r.emitWorkflowInputArgsAsStepOutput(tpl.Name, step, satisfiedArg)
					step = emitStep
					r.addToRootTemplates(wfTpl)

					artifact := findInputArtifact(r.tplInputArguments[tpl.Name], satisfiedArg)
					if artifact == nil {
						return nil, errors.Errorf("failed to find InputArtifact %s for step %s", satisfiedArg, step)
					}

					availableTypeInstances[argoArtifactRef{step.Name, satisfiedArg}] = artifact.typeInstanceReference
				}

				// 2.2 Check step with `capact-when` statements if it can be satisfied by input TypeInstances
				if satisfiedArg == "" {
					satisfiedArg, err = r.getInputTypeInstanceWhichSatisfyStep(step, typeInstances)
					if err != nil {
						return nil, err
					}

					// 2. Replace step and emit input TypeInstance as step output
					if satisfiedArg != "" {
						emitStep, wfTpl := r.emitWorkflowInputTypeInstanceAsStepOutput(tpl.Name, step, satisfiedArg)
						step = emitStep
						r.addToRootTemplates(wfTpl)

						typeInstance := findTypeInstanceInputRef(r.inputTypeInstances, satisfiedArg)
						if typeInstance == nil {
							return nil, errors.Errorf("failed to find InputTypeInstanceRef for %s", satisfiedArg)
						}
						r.tryReplaceTypeInstanceName(satisfiedArg, typeInstance.ID)

						namePtr := r.findTypeInstanceName(typeInstance.ID)
						availableTypeInstances[argoArtifactRef{step.Name, satisfiedArg}] = namePtr
					}
				}

				if err := r.registerUpdatedTypeInstances(step, availableTypeInstances, prefix); err != nil {
					return nil, errors.Wrap(err, "while registering updated TypeInstances")
				}

				// 3. Import and resolve Implementation for `capact-action`
				if step.CapactAction != nil {
					// 3.1 Expand `capact-action` alias based on imports section
					actionRef, err := hubpublicapi.ResolveActionPathFromImports(importsCollection, *step.CapactAction)
					if err != nil {
						return nil, err
					}

					// 3.2 Get InterfaceRevision
					iface, err := r.policyEnforcedCli.FindInterfaceRevision(ctx, *actionRef)
					if err != nil {
						return nil, err
					}

					if step.CapactPolicy != nil {
						err = step.CapactPolicy.ResolveImports(importsCollection)
						if err != nil {
							return nil, errors.Wrap(err, "while resolving import in WorkflowPolicy")
						}
						err = r.policyEnforcedCli.PushWorkflowStepPolicy(*step.CapactPolicy)
						if err != nil {
							return nil, errors.Wrap(err, "while adding WorkflowPolicy")
						}
					}
					// 3.3 Get all ImplementationRevisions for a given `capact-action`
					implementations, rule, err := r.policyEnforcedCli.ListImplementationRevisionForInterface(ctx, *actionRef)
					if err != nil {
						return nil, errors.Wrapf(err,
							`while listing ImplementationRevisions for step %q with action reference "%s:%s"`,
							step.Name, actionRef.Path, actionRef.Revision)
					}

					// 3.4 Pick one of the Implementations
					implementation, err := r.PickImplementationRevision(implementations)
					if err != nil {
						return nil, errors.Wrapf(err,
							`while picking ImplementationRevision for step %q with action reference with action reference "%s:%s"`,
							step.Name, actionRef.Path, actionRef.Revision)
					}

					workflowPrefix := addPrefix(tpl.Name, step.Name)

					// 3.6 Extract workflow from the imported `capact-action`. Prefix it to avoid artifacts name collision.
					importedWorkflow, newArtifactMappings, err := r.UnmarshalWorkflowFromImplementation(workflowPrefix, &implementation)
					if err != nil {
						return nil, errors.Wrap(err, "while creating workflow for action step")
					}

					// 3.7. List data based on policy and inject them if provided
					// 3.7.1 Required TypeInstances
					requiredTypeInstances, err := r.policyEnforcedCli.ListRequiredTypeInstancesToInjectBasedOnPolicy(rule, implementation)
					if err != nil {
						return nil, errors.Wrapf(err, "while listing RequiredTypeInstances based on policy for step: %s", step.Name)
					}

					err = r.InjectDownloadStepForTypeInstancesIfProvided(importedWorkflow, requiredTypeInstances)
					if err != nil {
						return nil, errors.Wrapf(err, "while injecting step for downloading TypeInstances based on policy for step: %s", step.Name)
					}
					// 3.7.2 Additional Input
					additionalParameters, err := r.policyEnforcedCli.ListAdditionalInputToInjectBasedOnPolicy(ctx, rule, implementation)
					if err != nil {
						return nil, errors.Wrap(err, "while converting additional input parameters")
					}
					r.InjectAdditionalInput(step, additionalParameters)

					for k, v := range newArtifactMappings {
						artifactMappings[k] = v
					}

					step.Template = importedWorkflow.Entrypoint
					step.CapactAction = nil

					// 3.8 Right now we know the template name, so let's try to register step input arguments
					r.registerTemplateInputArguments(step, availableTypeInstances)

					// 3.9 Add TypeInstances to the upload graph
					inputArtifacts := r.tplInputArguments[step.Template]
					if err := r.addOutputTypeInstancesToGraph(step, workflowPrefix, iface, &implementation, inputArtifacts); err != nil {
						return nil, errors.Wrap(err, "while adding TypeInstances to graph")
					}

					// 3.10 Render imported Workflow templates and add them to root templates
					// TODO(advanced-rendering): currently not supported.
					actionOutputTypeInstances, err := r.RenderTemplateSteps(ctx, importedWorkflow, implementation.Spec.Imports, nil, workflowPrefix)
					if err != nil {
						return nil, err
					}

					if step.CapactPolicy != nil {
						r.policyEnforcedCli.PopWorkflowStepPolicy()
					}
					step.CapactPolicy = nil

					// 3.11 Register output TypeInstances from this action step
					r.registerStepOutputTypeInstances(step, workflowPrefix, iface, actionOutputTypeInstances)
				}

				for name, tiPtr := range step.typeInstanceOutputs {
					availableTypeInstances[argoArtifactRef{step.Name, name}] = tiPtr
					outputTypeInstances[name] = tiPtr
				}

				step.CapactTypeInstanceOutputs = nil
				step.CapactTypeInstanceUpdates = nil

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

	return outputTypeInstances, nil
}

// TODO: take into account the runner revision. Respect that also in k8s engine when scheduling runner job
func (r *dedicatedRenderer) ResolveRunnerInterface(impl hubpublicapi.ImplementationRevision) (string, error) {
	imports, rInterface := impl.Spec.Imports, impl.Spec.Action.RunnerInterface
	fullRef, err := hubpublicapi.ResolveActionPathFromImports(imports, rInterface)
	if err != nil {
		return "", err
	}

	return fullRef.Path, nil
}

func (r *dedicatedRenderer) UnmarshalWorkflowFromImplementation(prefix string, implementation *hubpublicapi.ImplementationRevision) (*Workflow, map[string]string, error) {
	workflow, err := r.createWorkflow(implementation)
	if err != nil {
		return nil, nil, errors.Wrap(err, "while unmarshaling Argo Workflow from OCF Implementation")
	}

	artifactsNameMapping := map[string]string{}

	for i := range workflow.Templates {
		tmpl := workflow.Templates[i]

		// Change global artifacts names
		if prefix != "" {
			tmpl.Name = addPrefix(prefix, tmpl.Name)

			for artIdx := range tmpl.Outputs.Artifacts {
				artifact := &tmpl.Outputs.Artifacts[artIdx]

				if artifact.GlobalName == "" {
					continue
				}

				newName := addPrefix(prefix, artifact.GlobalName)
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
					step.Template = addPrefix(prefix, step.Template)
				}

				typeInstances := make([]TypeInstanceDefinition, 0, len(step.CapactTypeInstanceOutputs)+len(step.CapactTypeInstanceUpdates))
				typeInstances = append(typeInstances, step.CapactTypeInstanceOutputs...)
				typeInstances = append(typeInstances, step.CapactTypeInstanceUpdates...)

				for _, ti := range typeInstances {
					tiStep, template, artifactMappings := r.getOutputTypeInstanceTemplate(step, ti, prefix)
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
		workflow.Entrypoint = addPrefix(prefix, workflow.Entrypoint)
	}

	return workflow, artifactsNameMapping, nil
}

func (r *dedicatedRenderer) AddUserInputSecretRefIfProvided(rootWorkflow *Workflow) error {
	if r.inputParametersSecretRef == nil {
		return nil
	}

	// To ensure fixed order of steps in rendered workflow,
	// we have to iterate over the paramNames in a deterministic order.
	// In other case tests would fail sometimes.
	paramNames := make([]string, 0, len(r.inputParametersCollection))
	for name := range r.inputParametersCollection {
		paramNames = append(paramNames, name)
	}
	sort.Strings(paramNames)

	for _, paramName := range paramNames {
		r.addUserInputFromSecret(rootWorkflow, paramName)
	}

	return nil
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

func (r *dedicatedRenderer) InjectDownloadStepForTypeInstancesIfProvided(workflow *Workflow, typeInstances []types.InputTypeInstanceRef) error {
	if len(typeInstances) == 0 {
		return nil
	}
	return r.typeInstanceHandler.AddInputTypeInstances(workflow, typeInstances)
}

func (r *dedicatedRenderer) InjectAdditionalInput(step *WorkflowStep, additionalParams types.ParametersCollection) {
	for name, data := range additionalParams {
		step.Arguments.Artifacts = append(step.Arguments.Artifacts, wfv1.Artifact{
			Name: name,
			ArtifactLocation: wfv1.ArtifactLocation{
				Raw: &wfv1.RawArtifact{Data: data},
			}})
	}
}

func (r *dedicatedRenderer) PickImplementationRevision(in []hubpublicapi.ImplementationRevision) (hubpublicapi.ImplementationRevision, error) {
	if len(in) == 0 {
		return hubpublicapi.ImplementationRevision{}, errors.New("No Implementations found with current policy for given Interface")
	}

	// business decision - pick first Implementation
	return in[0], nil
}

// Internal helpers

func (r *dedicatedRenderer) registerStepOutputTypeInstances(step *WorkflowStep, prefix string, iface *hubpublicapi.InterfaceRevision, stepOutputTypeInstances map[string]*string) {
	step.typeInstanceOutputs = make(map[string]*string)

	if iface == nil || iface.Spec.Output == nil || iface.Spec.Output.TypeInstances == nil {
		return
	}

outer:
	for i := range iface.Spec.Output.TypeInstances {
		ti := iface.Spec.Output.TypeInstances[i]

		for _, tiOutput := range step.CapactTypeInstanceOutputs {
			if tiOutput.Name != ti.Name {
				continue
			}

			if ptr, ok := stepOutputTypeInstances[tiOutput.From]; ok {
				step.typeInstanceOutputs[ti.Name] = ptr
				continue outer
			}
		}

		newName := addPrefix(prefix, ti.Name)
		newNamePtr := r.addTypeInstanceName(newName)
		step.typeInstanceOutputs[ti.Name] = newNamePtr
	}
}

func (*dedicatedRenderer) createWorkflow(implementation *hubpublicapi.ImplementationRevision) (*Workflow, error) {
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

	// ToLower as the name is used for k8s pod name:
	//    Pod "e2e-test-1-8081-output-testUpload-3658329211" is invalid: metadata.name: Invalid value: "e2e-test-1-8081-output-testUpload-3658329211": a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character (e.g. 'example.com', regex used for validation is '[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')
	// change in Argo was introduced in: https://github.com/argoproj/argo-workflows/commit/8897fff15776f31fbd7f65bbee4f93b2101110f7
	stepName := strings.ToLower(fmt.Sprintf("output-%s", output.Name))

	var templateName string
	var artifactGlobalName string
	if prefix == "" {
		templateName = stepName
		artifactGlobalName = output.Name
	} else {
		templateName = fmt.Sprintf("output-%s-%s", prefix, output.Name)
		artifactGlobalName = addPrefix(prefix, output.Name)
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
//         template: app-install								# We record input arguments under template name.
//         arguments:
//           artifacts:
//             - name: postgresql
//               from: "{{steps.install-db.outputs.artifacts.postgresql}}"
//   - name: app-install
//     inputs:
//      artifacts:
//        - name: input-parameters
//        - name: postgresql
//          optional: true
//    steps:
//      - - capact-when: postgresql == nil						# Check whether this step is satisfied by input arguments.
//          name: install-db									# For that we need to have option to check which arguments were passed
//																# to this step.
func (r *dedicatedRenderer) getInputArgWhichSatisfyStep(tplOwnerName string, step *WorkflowStep) (string, error) {
	if step.CapactWhen == nil {
		return "", nil
	}

	args, found := r.tplInputArguments[tplOwnerName]
	if !found {
		return "", nil
	}

	params := &mapEvalParameters{}
	for _, a := range args {
		params.Set(a.artifact.Name)
	}

	notSatisfied, err := r.evaluateWhenExpression(params, *step.CapactWhen)
	if err != nil {
		return "", errors.Wrap(err, "while evaluating OCFWhen")
	}

	if notSatisfied == true {
		return "", nil
	}

	step.CapactWhen = nil
	return params.lastAccessed, nil
}

func (r *dedicatedRenderer) getInputTypeInstanceWhichSatisfyStep(step *WorkflowStep, typeInstances []types.InputTypeInstanceRef) (string, error) {
	if step.CapactWhen == nil {
		return "", nil
	}

	params := &mapEvalParameters{}
	for _, t := range typeInstances {
		params.Set(t.Name)
	}

	notSatisfied, err := r.evaluateWhenExpression(params, *step.CapactWhen)
	if err != nil {
		return "", errors.Wrap(err, "while evaluating OCFWhen")
	}

	// zero value to mark as handled
	step.CapactWhen = nil

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
	escapeDashes := strings.Replace(exprString, "-", "\\-", -1)
	expr, err := govaluate.NewEvaluableExpression(escapeDashes)
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

func (r *dedicatedRenderer) addUserInputFromSecret(rootWorkflow *Workflow, parameterName string) {
	var (
		volumeName = "user-secret-volume"
		mountPath  = "/input"
		outputPath = path.Join(mountPath, parameterName)
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
		Name:      fmt.Sprintf("populate-%s", parameterName),
		Container: container,
		Outputs: wfv1.Outputs{
			Artifacts: wfv1.Artifacts{
				{
					Name: parameterName,
					Path: outputPath,
				},
			},
		},
		Volumes: []apiv1.Volume{
			{
				Name: volumeName,
				VolumeSource: apiv1.VolumeSource{
					Secret: &apiv1.SecretVolumeSource{
						SecretName: r.inputParametersSecretRef.Name,
						Items: []apiv1.KeyToPath{
							{
								Key:  action.GetParameterDataKey(parameterName),
								Path: parameterName,
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
		Name: parameterName,
		From: fmt.Sprintf("{{steps.%s.outputs.artifacts.%s}}", userInputWfStep.Name, parameterName),
	})
}

// TODO: current limitation: we handle properly only one artifacts `capact-when: postgres == nil` but not `capact-when: postgres == nil && app-config == nil`
func (r *dedicatedRenderer) emitWorkflowInputAsStepOutput(tplName string, step *WorkflowStep, inputArgName string, reference string) (*WorkflowStep, *Template) {
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
					From: fmt.Sprintf(reference, inputArgName),
				},
			},
		},
	}
	return &WorkflowStep{WorkflowStep: userInputWfStep}, &Template{Template: userInputWfTpl}
}

func (r *dedicatedRenderer) emitWorkflowInputArgsAsStepOutput(tplName string, step *WorkflowStep, inputArgName string) (*WorkflowStep, *Template) {
	return r.emitWorkflowInputAsStepOutput(tplName, step, inputArgName, "{{inputs.artifacts.%s}}")
}

func (r *dedicatedRenderer) emitWorkflowInputTypeInstanceAsStepOutput(tplName string, step *WorkflowStep, inputArgName string) (*WorkflowStep, *Template) {
	return r.emitWorkflowInputAsStepOutput(tplName, step, inputArgName, "{{workflow.outputs.artifacts.%s}}")
}

func (r *dedicatedRenderer) registerTemplateInputArguments(step *WorkflowStep, availableTypeInstances map[argoArtifactRef]*string) {
	if step.GetTemplateName() == "" || len(step.Arguments.Artifacts) == 0 {
		return
	}

	inputArtifacts := []InputArtifact{}

	for _, art := range step.Arguments.Artifacts {
		inputArt := InputArtifact{
			artifact: art,
		}

		ref, err := getArgoArtifactRef(art.From)
		if err != nil {
			continue
		}

		if tiName, ok := availableTypeInstances[*ref]; ok {
			inputArt.typeInstanceReference = tiName
		}

		inputArtifacts = append(inputArtifacts, inputArt)
	}

	r.tplInputArguments[step.Template] = inputArtifacts
}

func (r *dedicatedRenderer) addOutputTypeInstancesToGraph(step *WorkflowStep, prefix string, iface *hubpublicapi.InterfaceRevision, impl *hubpublicapi.ImplementationRevision, inputArtifacts []InputArtifact) error {
	artifactNamesMap := map[string]*string{}
	for _, artifact := range inputArtifacts {
		artifactNamesMap[artifact.artifact.Name] = artifact.typeInstanceReference
	}

	for _, item := range impl.Spec.OutputTypeInstanceRelations {
		name := item.TypeInstanceName
		if step != nil {
			// we have to track the renaming based on capact-outputTypeInstances and prefix it
			if output := findOutputTypeInstance(step, item.TypeInstanceName); output != nil {
				name = addPrefix(prefix, output.From)
				r.tryReplaceTypeInstanceName(output.Name, name)
			} else {
				// if the TypeInstance was not defined in capact-outputTypeInstances, then just prefix it
				name = addPrefix(prefix, item.TypeInstanceName)
				r.tryReplaceTypeInstanceName(item.TypeInstanceName, name)
			}
		}

		typeRef, err := findTypeInstanceTypeRef(item.TypeInstanceName, impl, iface)
		if err != nil {
			return err
		}

		artifactName := r.addTypeInstanceName(name)
		artifactNamesMap[item.TypeInstanceName] = artifactName

		r.typeInstancesToOutput.typeInstances = append(r.typeInstancesToOutput.typeInstances, OutputTypeInstance{
			ArtifactName: artifactName,
			TypeInstance: types.OutputTypeInstance{
				TypeRef: &types.TypeRef{
					Path:     typeRef.Path,
					Revision: typeRef.Revision,
				},
			},
		})

		for _, uses := range item.Uses {
			usesArtifactName, ok := artifactNamesMap[uses]
			if !ok {
				usesArtifactName = r.addTypeInstanceName(uses)
			}

			r.typeInstancesToOutput.relations = append(r.typeInstancesToOutput.relations, OutputTypeInstanceRelation{
				From: artifactName,
				To:   usesArtifactName,
			})
		}
	}

	return nil
}

func (r *dedicatedRenderer) registerUpdatedTypeInstances(step *WorkflowStep, availableTypeInstances map[argoArtifactRef]*string, prefix string) error {
	for _, update := range step.CapactTypeInstanceUpdates {
		typeInstance, ok := availableTypeInstances[argoArtifactRef{
			step: ArgoArtifactNoStep,
			name: update.Name,
		}]

		if !ok {
			return errors.Errorf("failed to find TypeInstance for %s", update.Name)
		}

		name := update.Name
		if prefix != "" {
			name = addPrefix(prefix, name)
		}

		r.typeInstancesToUpdate = append(r.typeInstancesToUpdate, UpdateTypeInstance{
			ArtifactName: name,
			ID:           *typeInstance,
		})
	}

	return nil
}

func (r *dedicatedRenderer) addTypeInstanceName(name string) *string {
	foundName := r.findTypeInstanceName(name)
	if foundName != nil {
		return foundName
	}

	r.registeredOutputTypeInstanceNames = append(r.registeredOutputTypeInstanceNames, &name)

	return &name
}

func (r *dedicatedRenderer) findTypeInstanceName(name string) *string {
	for i := range r.registeredOutputTypeInstanceNames {
		if *r.registeredOutputTypeInstanceNames[i] == name {
			return r.registeredOutputTypeInstanceNames[i]
		}
	}

	return nil
}

func (r *dedicatedRenderer) tryReplaceTypeInstanceName(oldName, newName string) {
	found := r.findTypeInstanceName(oldName)
	if found == nil {
		return
	}

	*found = newName
}
