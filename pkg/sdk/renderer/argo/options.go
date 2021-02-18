package argo

import (
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
