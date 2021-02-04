package argo

import (
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type renderOptions struct {
	plainTextUserInput map[string]interface{}
	inputTypeInstances []types.InputTypeInstanceRef
}

type RendererOption func(*renderOptions)

func WithPlainTextUserInput(data map[string]interface{}) RendererOption {
	return func(r *renderOptions) {
		r.plainTextUserInput = data
	}
}

func WithTypeInstances(typeInstances []types.InputTypeInstanceRef) RendererOption {
	return func(r *renderOptions) {
		r.inputTypeInstances = typeInstances
	}
}
