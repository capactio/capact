package model

import (
	v1 "k8s.io/api/core/v1"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

// ActionToCreateOrUpdate holds data to create or update all Action details.
type ActionToCreateOrUpdate struct {
	Action            v1alpha1.Action
	InputParamsSecret *v1.Secret
}

func (m *ActionToCreateOrUpdate) SetNamespace(namespace string) {
	m.Action.Namespace = namespace
	m.InputParamsSecret.Namespace = namespace
}

// ActionFilter defines filtering options for Actions
type ActionFilter struct {
	Phase *v1alpha1.ActionPhase
}

// Input used for continuing Action rendering in advanced mode
type AdvancedModeContinueRenderingInput struct {
	// TypeInstances that are optional for a given rendering iteration
	TypeInstances *[]v1alpha1.InputTypeInstance
}
