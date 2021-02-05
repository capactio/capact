package argo

import (
	"context"
	"encoding/json"
	"time"

	"projectvoltron.dev/voltron/pkg/och/client/public"

	ochlocalgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/local"
	ochpublicgraphql "projectvoltron.dev/voltron/pkg/och/api/graphql/public"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
	"projectvoltron.dev/voltron/pkg/sdk/renderer"

	"github.com/pkg/errors"
)

const (
	userInputName = "input-parameters"
	runnerContext = "runner-context"
)

type OCHClient interface {
	GetImplementationRevisionsForInterface(ctx context.Context, ref ochpublicgraphql.InterfaceReference, opts ...public.GetImplementationOption) ([]ochpublicgraphql.ImplementationRevision, error)
	GetTypeInstance(ctx context.Context, id string) (*ochlocalgraphql.TypeInstance, error)
}

type Renderer struct {
	ochCli        OCHClient
	maxDepth      int
	renderTimeout time.Duration

	typeInstanceHandler TypeInstanceHandler
}

func NewRenderer(cfg renderer.Config, ochCli OCHClient) *Renderer {
	r := &Renderer{
		ochCli:              ochCli,
		typeInstanceHandler: TypeInstanceHandler{ochCli: ochCli},
		maxDepth:            cfg.MaxDepth,
		renderTimeout:       cfg.RenderTimeout,
	}

	return r
}

func (r *Renderer) Render(ctx context.Context, runnerCtxSecretRef RunnerContextSecretRef, ref types.InterfaceRef, opts ...RendererOption) (*types.Action, error) {
	// 0. Populate render options
	dedicateRenderer := newDedicatedRenderer(r.maxDepth, r.ochCli, &r.typeInstanceHandler, opts...)

	ctxWithTimeout, cancel := context.WithTimeout(ctx, r.renderTimeout)
	defer cancel()

	// 1. Find the root implementation
	implementations, err := r.ochCli.GetImplementationRevisionsForInterface(ctxWithTimeout, interfaceRefToOCH(ref), dedicateRenderer.ochImplementationFilters...)
	if err != nil {
		return nil, err
	}

	// business decision select the first one
	implementation := implementations[0]

	// 2. Ensure that the runner was defined in imports section
	// TODO: we should check whether imported revision is valid for this render algorithm
	runnerInterface, err := dedicateRenderer.ResolveRunnerInterface(implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while resolving runner interface")
	}

	// 3. Extract workflow from the root Implementation
	rootWorkflow, _, err := dedicateRenderer.UnmarshalWorkflowFromImplementation("", &implementation)
	if err != nil {
		return nil, errors.Wrap(err, "while creating root workflow")
	}

	// 4. Add user input if provided
	if err := dedicateRenderer.AddPlainTextUserInput(rootWorkflow); err != nil {
		return nil, err
	}

	// 5. Add runner context
	if err := dedicateRenderer.AddRunnerContext(rootWorkflow, runnerCtxSecretRef); err != nil {
		return nil, err
	}

	// 6. Add steps to populate rootWorkflow with input TypeInstances
	if err := dedicateRenderer.AddInputTypeInstance(rootWorkflow); err != nil {
		return nil, err
	}

	// 7. Render rootWorkflow templates
	err = dedicateRenderer.RenderTemplateSteps(ctxWithTimeout, rootWorkflow, implementation.Spec.Imports, dedicateRenderer.inputTypeInstances)
	if err != nil {
		return nil, err
	}

	rootWorkflow.Templates = dedicateRenderer.GetRootTemplates()

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
