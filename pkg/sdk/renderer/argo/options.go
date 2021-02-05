package argo

import (
	"projectvoltron.dev/voltron/pkg/sdk/apis/0.0.1/types"
)

type renderOptions struct {
	plainTextUserInput      map[string]interface{}
	runnerContextFromSecret runnerContextFromSecret
	inputTypeInstances      []types.InputTypeInstanceRef
}

type runnerContextFromSecret struct {
	Name string
	Key string
}

type RendererOption func(*renderOptions)

func WithPlainTextUserInput(data map[string]interface{}) RendererOption {
	return func(r *renderOptions) {
		r.plainTextUserInput = data
	}
}

func WithRunnerContextFromSecret(secretName, keyName string) RendererOption {
	return func(r *renderOptions) {
		r.runnerContextFromSecret = runnerContextFromSecret{
			Name: secretName,
			Key: keyName,
		}
	}
}

func WithTypeInstances(typeInstances []types.InputTypeInstanceRef) RendererOption {
	return func(r *renderOptions) {
		r.inputTypeInstances = typeInstances
	}
}
