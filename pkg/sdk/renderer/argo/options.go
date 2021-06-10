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

func WithPolicy(policy policy.Policy) RendererOption {
	return func(r *dedicatedRenderer) {
		r.policyEnforcedCli.SetPolicy(policy)
	}
}

func WithOwnerID(ownerID string) RendererOption {
	return func(r *dedicatedRenderer) {
		r.ownerID = &ownerID
	}
}
