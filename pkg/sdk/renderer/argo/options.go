package argo

import (
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

type RendererOption func(*dedicatedRenderer)

func WithTypeInstances(typeInstances []types.InputTypeInstanceRef) RendererOption {
	return func(r *dedicatedRenderer) {
		r.inputTypeInstances = typeInstances
	}
}

func WithSecretUserInput(ref *UserInputSecretRef) RendererOption {
	return func(r *dedicatedRenderer) {
		r.userInputSecretRef = ref
	}
}

func WithGlobalPolicy(policy policy.Policy) RendererOption {
	return func(r *dedicatedRenderer) {
		r.policyEnforcedCli.SetGlobalPolicy(policy)
	}
}

func WithActionPolicy(policy policy.Policy) RendererOption {
	return func(r *dedicatedRenderer) {
		r.policyEnforcedCli.SetActionPolicy(policy)
	}
}

func WithPolicyOrder(order policy.MergeOrder) RendererOption {
	return func(r *dedicatedRenderer) {
		r.policyEnforcedCli.SetPolicyOrder(order)
	}
}

func WithOwnerID(ownerID string) RendererOption {
	return func(r *dedicatedRenderer) {
		r.ownerID = &ownerID
	}
}
