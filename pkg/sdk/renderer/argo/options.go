package argo

import (
	"projectvoltron.dev/voltron/pkg/engine/k8s/clusterpolicy"
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
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

func WithPolicy(policy clusterpolicy.ClusterPolicy) RendererOption {
	return func(r *dedicatedRenderer) {
		r.policyEnforcedCli.SetPolicy(policy)
	}
}
