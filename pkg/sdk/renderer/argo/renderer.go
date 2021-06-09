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
	userInputName       = "input-parameters"
	additionalInputName = "additional-parameters"
	runnerContext       = "runner-context"
)

type PolicyEnforcedHubClient interface {
	ListImplementationRevisionForInterface(ctx context.Context, interfaceRef hubpublicapi.InterfaceReference) ([]hubpublicapi.ImplementationRevision, policy.Rule, error)
	ListTypeInstancesToInjectBasedOnPolicy(policyRule policy.Rule, implRev hubpublicapi.ImplementationRevision) []types.InputTypeInstanceRef
	ListAdditionalInputToInjectBasedOnPolicy(policyRule policy.Rule, implRev hubpublicapi.ImplementationRevision) map[string]interface{}
	SetGlobalPolicy(policy policy.Policy)
	SetActionPolicy(policy policy.Policy)
	SetWorkflowStepPolicy(policy policy.Policy)
	SetPolicyOrder(policy.MergeOrder)
	FindInterfaceRevision(ctx context.Context, ref hubpublicapi.InterfaceReference) (*hubpublicapi.InterfaceRevision, error)
}

type Renderer struct {
	maxDepth      int
	renderTimeout time.Duration

	typeInstanceHandler *TypeInstanceHandler
	hubClient           hubclient.HubClient
}

func NewRenderer(cfg renderer.Config, hubClient hubclient.HubClient, typeInstanceHandler *TypeInstanceHandler) *Renderer {
	r := &Renderer{
		typeInstanceHandler: typeInstanceHandler,
		maxDepth:            cfg.MaxDepth,
		renderTimeout:       cfg.RenderTimeout,
		hubClient:           hubClient,
	}

	return r
}

func (r *Renderer) Render(ctx context.Context, input *RenderInput) (*RenderOutput, error) {
	if input == nil {
		input = &RenderInput{}
	}

	// policyEnforcedClient cannot be global because policy is calculated from global policy, action policy and worklfow step policies
	policyEnforcedClient := hubclient.NewPolicyEnforcedClient(r.hubClient)

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
	// 5.1 TypeInstances
	typeInstancesToInject := policyEnforcedClient.ListTypeInstancesToInjectBasedOnPolicy(rule, implementation)
	err = dedicatedRenderer.InjectDownloadStepForTypeInstancesIfProvided(rootWorkflow, typeInstancesToInject)
	if err != nil {
		return nil, errors.Wrap(err, "while injecting step for downloading TypeInstances based on policy")
	}
	// 5.2 Additional Input
	additionalInput := policyEnforcedClient.ListAdditionalInputToInjectBasedOnPolicy(rule, implementation)
	err = dedicatedRenderer.InjectAdditionalInput(entrypointStep, additionalInput)
	if err != nil {
		return nil, errors.Wrap(err, "while injecting additional input based on policy")
	}

	// 6. Add runner context
	if err := dedicatedRenderer.AddRunnerContext(rootWorkflow, input.RunnerContextSecretRef); err != nil {
		return nil, err
	}

	// 7. Add steps to populate rootWorkflow with input TypeInstances
	if err := dedicatedRenderer.AddInputTypeInstances(rootWorkflow); err != nil {
		return nil, err
	}

	availableArtifacts := dedicatedRenderer.tplInputArguments[dedicatedRenderer.entrypointStep.Template]

	// 8 Register output TypeInstances
	if err := dedicatedRenderer.addOutputTypeInstancesToGraph(nil, "", iface, &implementation, availableArtifacts); err != nil {
		return nil, errors.Wrap(err, "while noting output artifacts")
	}

	// 9. Render rootWorkflow templates
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
