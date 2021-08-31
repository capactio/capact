package argo

import (
	"context"
	"encoding/json"
	"time"

	"capact.io/capact/pkg/engine/k8s/policy"
	hubpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/renderer"

	"github.com/pkg/errors"
)

const (
	// UserInputName is exported so we can use that as a reference in `capact act create` validation process.
	UserInputName = "input-parameters"
	runnerContext = "runner-context"
)

// PolicyEnforcedHubClient is a interfaces used to interact with the Capact Hubs
// and enforce the policies.
type PolicyEnforcedHubClient interface {
	ListImplementationRevisionForInterface(ctx context.Context, interfaceRef hubpublicapi.InterfaceReference) ([]hubpublicapi.ImplementationRevision, policy.Rule, error)
	ListRequiredTypeInstancesToInjectBasedOnPolicy(policyRule policy.Rule, implRev hubpublicapi.ImplementationRevision) ([]types.InputTypeInstanceRef, error)
	ListAdditionalTypeInstancesToInjectBasedOnPolicy(policyRule policy.Rule, implRev hubpublicapi.ImplementationRevision) ([]types.InputTypeInstanceRef, error)
	ListAdditionalInputToInjectBasedOnPolicy(ctx context.Context, policyRule policy.Rule, implRev hubpublicapi.ImplementationRevision) (types.ParametersCollection, error)
	SetGlobalPolicy(policy policy.Policy)
	SetActionPolicy(policy policy.ActionPolicy)
	PushWorkflowStepPolicy(policy policy.WorkflowPolicy) error
	PopWorkflowStepPolicy()
	SetPolicyOrder(policy.MergeOrder)
	FindInterfaceRevision(ctx context.Context, ref hubpublicapi.InterfaceReference) (*hubpublicapi.InterfaceRevision, error)
}

type workflowValidator interface {
	ValidateInterfaceInput(context.Context, renderer.InterfaceInput) error
	PolicyValidator() hubclient.PolicyIOValidator
}

// Renderer is used to render the Capact Action workflows.
type Renderer struct {
	maxDepth      int
	renderTimeout time.Duration

	typeInstanceHandler *TypeInstanceHandler
	wfValidator         workflowValidator
	hubClient           hubclient.HubClient
}

// NewRenderer returns a new Renderer instance.
func NewRenderer(cfg renderer.Config, hubClient hubclient.HubClient, typeInstanceHandler *TypeInstanceHandler, validator workflowValidator) *Renderer {
	r := &Renderer{
		typeInstanceHandler: typeInstanceHandler,
		maxDepth:            cfg.MaxDepth,
		renderTimeout:       cfg.RenderTimeout,
		hubClient:           hubClient,
		wfValidator:         validator,
	}

	return r
}

// Render performs the rendering of an Action workflow.
func (r *Renderer) Render(ctx context.Context, input *RenderInput) (*RenderOutput, error) {
	if input == nil {
		input = &RenderInput{}
	}

	// policyEnforcedClient cannot be global because policy is calculated from global policy, action policy and workflow step policies
	policyEnforcedClient := hubclient.NewPolicyEnforcedClient(r.hubClient, r.wfValidator.PolicyValidator())

	// 0. Populate render options
	dedicatedRenderer := newDedicatedRenderer(r.maxDepth, policyEnforcedClient, r.typeInstanceHandler, input.Options...)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, r.renderTimeout)
	defer cancel()

	// 1. Find the root manifests
	interfaceRef := interfaceRefToHub(input.InterfaceRef)

	// 1.1 Get Interface
	iface, err := policyEnforcedClient.FindInterfaceRevision(ctx, interfaceRef)
	if err != nil {
		return nil, err
	}

	// 1.2 Get all ImplementationRevisions for a given Interface
	implementations, rule, err := policyEnforcedClient.ListImplementationRevisionForInterface(ctxWithTimeout, interfaceRef)
	if err != nil {
		return nil, errors.Wrapf(err, `while listing ImplementationRevisions for Interface "%s:%s"`,
			interfaceRef.Path, interfaceRef.Revision,
		)
	}

	// 1.3 Pick one of the Implementations
	implementation, err := dedicatedRenderer.PickImplementationRevision(implementations)
	if err != nil {
		return nil, errors.Wrapf(err, `while picking ImplementationRevision for Interface "%s:%s"`,
			interfaceRef.Path, interfaceRef.Revision)
	}

	// 2. Ensure that the runner was defined in imports section
	// TODO: we should check whether imported revision is valid for this render algorithm
	runnerInterface, err := dedicatedRenderer.ResolveRunnerInterface(implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while resolving runner Interface")
	}

	// 3. Extract workflow from the root Implementation
	rootWorkflow, _, err := dedicatedRenderer.UnmarshalWorkflowFromImplementation("", &implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while creating root workflow")
	}

	// 3.1 Add our own root step and replace entrypoint
	rootWorkflow, entrypointStep := dedicatedRenderer.WrapEntrypointWithRootStep(rootWorkflow)

	// 4. Add user input
	dedicatedRenderer.AddUserInputSecretRefIfProvided(rootWorkflow)

	// 5. List data based on policy and inject them if provided
	// 5.1 Required TypeInstances
	requiredTypeInstances, err := policyEnforcedClient.ListRequiredTypeInstancesToInjectBasedOnPolicy(rule, implementation)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing RequiredTypeInstances based on policy for root workflow")
	}

	// 5.2 Additional TypeInstances
	additionalTypeInstances, err := policyEnforcedClient.ListAdditionalTypeInstancesToInjectBasedOnPolicy(rule, implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while listing AdditionalTypeInstances based on policy for root workflow")
	}

	// 5.3 Inject TypeInstances
	typeInstancesToInject := append(requiredTypeInstances, additionalTypeInstances...)
	err = dedicatedRenderer.InjectDownloadStepForTypeInstancesIfProvided(rootWorkflow, typeInstancesToInject)
	if err != nil {
		return nil, errors.Wrap(err, "while injecting step for downloading additional TypeInstances based on policy")
	}
	// 5.4 Additional Input
	additionalParameters, err := policyEnforcedClient.ListAdditionalInputToInjectBasedOnPolicy(ctx, rule, implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while converting additional parameters")
	}
	dedicatedRenderer.InjectAdditionalInput(entrypointStep, additionalParameters)

	// 6. Validate workflow input against Interface:
	// Implementation-specific input is already validated on PolicyEnforcedClient level
	validateInput := renderer.InterfaceInput{
		Interface:     iface,
		Parameters:    ToInputParams(dedicatedRenderer.inputParametersRaw),
		TypeInstances: dedicatedRenderer.inputTypeInstances,
	}
	err = r.wfValidator.ValidateInterfaceInput(ctx, validateInput)

	if err != nil {
		return nil, errors.Wrap(err, "while validating required and additional input data")
	}
	// 7. Add runner context
	if err := dedicatedRenderer.AddRunnerContext(rootWorkflow, input.RunnerContextSecretRef); err != nil {
		return nil, err
	}

	// 8. Add steps to populate rootWorkflow with input TypeInstances
	if err := dedicatedRenderer.AddInputTypeInstances(rootWorkflow); err != nil {
		return nil, err
	}

	availableArtifacts := dedicatedRenderer.tplInputArguments[dedicatedRenderer.entrypointStep.Template]

	// 9. Register output TypeInstances
	if err := dedicatedRenderer.addOutputTypeInstancesToGraph(nil, "", iface, &implementation, availableArtifacts); err != nil {
		return nil, errors.Wrap(err, "while noting output artifacts")
	}

	// 10. Render rootWorkflow templates
	_, err = dedicatedRenderer.RenderTemplateSteps(ctxWithTimeout, rootWorkflow, implementation.Spec.Imports, dedicatedRenderer.inputTypeInstances, "")
	if err != nil {
		return nil, err
	}

	rootWorkflow.Templates = dedicatedRenderer.GetRootTemplates()

	if err := dedicatedRenderer.AddOutputTypeInstancesStep(rootWorkflow); err != nil {
		return nil, err
	}

	out, err := r.toMapStringInterface(rootWorkflow)
	if err != nil {
		return nil, err
	}

	return &RenderOutput{
		Action: &types.Action{
			Args:            out,
			RunnerInterface: runnerInterface,
		},
		TypeInstancesToLock: dedicatedRenderer.GetTypeInstancesToLock(),
	}, nil
}

func (r *Renderer) toMapStringInterface(w *Workflow) (map[string]interface{}, error) {
	var renderedWorkflow = struct {
		Spec Workflow `json:"workflow"`
	}{
		Spec: *w,
	}
	out := map[string]interface{}{}
	marshalled, err := json.Marshal(renderedWorkflow)
	if err != nil {
		return nil, err
	}

	if err = json.Unmarshal(marshalled, &out); err != nil {
		return nil, err
	}

	return out, nil
}
