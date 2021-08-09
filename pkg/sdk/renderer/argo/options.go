package argo

import (
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// RendererOption is used to provide additional configuration options to the rendering process.
type RendererOption func(*dedicatedRenderer)

// WithTypeInstances returns a RendererOption, which adds input TypeInstances to the workflow.
func WithTypeInstances(typeInstances []types.InputTypeInstanceRef) RendererOption {
	return func(r *dedicatedRenderer) {
		r.inputTypeInstances = typeInstances
	}
}

// WithSecretUserInput returns a RendererOption, which adds user input to the workflow.
func WithSecretUserInput(ref *UserInputSecretRef, inputRaw []byte) RendererOption {
	return func(r *dedicatedRenderer) {
		r.inputParametersSecretRef = ref
		r.inputParametersRaw = string(inputRaw)
	}
}

// WithGlobalPolicy returns a RendererOption, which sets the Global policy for the rendering process.
func WithGlobalPolicy(policy policy.Policy) RendererOption {
	return func(r *dedicatedRenderer) {
		r.policyEnforcedCli.SetGlobalPolicy(policy)
	}
}

// WithActionPolicy returns a RendererOption, which sets Action policy for the rendering process.
func WithActionPolicy(policy policy.ActionPolicy) RendererOption {
	return func(r *dedicatedRenderer) {
		r.policyEnforcedCli.SetActionPolicy(policy)
	}
}

// WithPolicyOrder returns a RendererOption, which sets the priority order for the policy.
// The priorty order is from the most important policy to the least important.
func WithPolicyOrder(order policy.MergeOrder) RendererOption {
	return func(r *dedicatedRenderer) {
		r.policyEnforcedCli.SetPolicyOrder(order)
	}
}

// WithOwnerID returns a RendererOption, which sets OwnerID for the workflow.
// The OwnerID is used to lock the TypeInstances.
func WithOwnerID(ownerID string) RendererOption {
	return func(r *dedicatedRenderer) {
		r.ownerID = &ownerID
	}
}
