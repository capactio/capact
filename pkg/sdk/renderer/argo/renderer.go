package argo

import (
	"context"
	"encoding/json"
	"time"

	"projectvoltron.dev/voltron/pkg/engine/k8s/clusterpolicy"
	ochpublicapi "projectvoltron.dev/voltron/pkg/och/api/graphql/public"

	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"projectvoltron.dev/voltron/pkg/sdk/renderer"

	"github.com/pkg/errors"
)

const (
	userInputName = "input-parameters"
	runnerContext = "runner-context"
)

type PolicyEnforcedOCHClient interface {
	ListImplementationRevisionForInterface(ctx context.Context, interfaceRef ochpublicapi.InterfaceReference) ([]ochpublicapi.ImplementationRevision, clusterpolicy.Rule, error)
	ListTypeInstancesToInjectBasedOnPolicy(policyRule clusterpolicy.Rule, implRev ochpublicapi.ImplementationRevision) []types.InputTypeInstanceRef
	SetPolicy(policy clusterpolicy.ClusterPolicy)
	GetInterfaceRevision(ctx context.Context, ref ochpublicapi.InterfaceReference) (*ochpublicapi.InterfaceRevision, error)
}

type Renderer struct {
	maxDepth      int
	renderTimeout time.Duration

	policyEnforcedCli   PolicyEnforcedOCHClient
	typeInstanceHandler *TypeInstanceHandler
}

func NewRenderer(cfg renderer.Config, policyEnforcedCli PolicyEnforcedOCHClient, typeInstanceHandler *TypeInstanceHandler) *Renderer {
	r := &Renderer{
		typeInstanceHandler: typeInstanceHandler,
		policyEnforcedCli:   policyEnforcedCli,
		maxDepth:            cfg.MaxDepth,
		renderTimeout:       cfg.RenderTimeout,
	}

	return r
}

func (r *Renderer) Render(ctx context.Context, runnerCtxSecretRef RunnerContextSecretRef, ref types.InterfaceRef, opts ...RendererOption) (*types.Action, error) {
	// 0. Populate render options
	dedicatedRenderer := newDedicatedRenderer(r.maxDepth, r.policyEnforcedCli, r.typeInstanceHandler, opts...)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, r.renderTimeout)
	defer cancel()

	// 1. Find the root manifests
	interfaceRef := interfaceRefToOCH(ref)

	// 1.1 Get Interface
	iface, err := r.policyEnforcedCli.GetInterfaceRevision(ctx, interfaceRef)
	if err != nil {
		return nil, err
	}

	// 1.2 Get all ImplementationRevisions for a given Interface
	implementations, rule, err := r.policyEnforcedCli.ListImplementationRevisionForInterface(ctxWithTimeout, interfaceRef)
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

	// 2 Register output TypeInstances
	if err := dedicatedRenderer.addOutputTypeInstancesToGraph(nil, "", iface, &implementation, []InputArtifact{}); err != nil {
		return nil, errors.Wrap(err, "while noting output artifacts")
	}

	// 3. Ensure that the runner was defined in imports section
	// TODO: we should check whether imported revision is valid for this render algorithm
	runnerInterface, err := dedicatedRenderer.ResolveRunnerInterface(implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while resolving runner Interface")
	}

	// 4. Extract workflow from the root Implementation
	rootWorkflow, _, err := dedicatedRenderer.UnmarshalWorkflowFromImplementation("", &implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while creating root workflow")
	}

	// 4.1 Add our own root step and replace entrypoint
	rootWorkflow = dedicatedRenderer.WrapEntrypointWithRootStep(rootWorkflow)

	// 5. Add user input
	dedicatedRenderer.AddUserInputSecretRefIfProvided(rootWorkflow)

	// 6. List TypeInstances to inject based on policy and inject them if provided
	typeInstancesToInject := r.policyEnforcedCli.ListTypeInstancesToInjectBasedOnPolicy(rule, implementation)
	err = dedicatedRenderer.InjectDownloadStepForTypeInstancesIfProvided(rootWorkflow, typeInstancesToInject)
	if err != nil {
		return nil, errors.Wrap(err, "while injecting step for downloading TypeInstances based on policy")
	}

	// 7. Add runner context
	if err := dedicatedRenderer.AddRunnerContext(rootWorkflow, runnerCtxSecretRef); err != nil {
		return nil, err
	}

	// 8. Add steps to populate rootWorkflow with input TypeInstances
	if err := dedicatedRenderer.AddInputTypeInstances(rootWorkflow); err != nil {
		return nil, err
	}

	// 9. Render rootWorkflow templates
	err = dedicatedRenderer.RenderTemplateSteps(ctxWithTimeout, rootWorkflow, implementation.Spec.Imports, dedicatedRenderer.inputTypeInstances)
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

	return &types.Action{
		Args:            out,
		RunnerInterface: runnerInterface,
	}, nil
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
