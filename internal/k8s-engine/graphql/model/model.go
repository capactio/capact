package model

import (
	"regexp"

	"capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
)

// ActionToCreateOrUpdate holds data to create or update all Action details.
type ActionToCreateOrUpdate struct {
	Action            v1alpha1.Action
	InputParamsSecret *v1.Secret
}

// SetNamespace sets a given namespace to all required poperties.
func (m *ActionToCreateOrUpdate) SetNamespace(namespace string) {
	m.Action.Namespace = namespace

	if m.InputParamsSecret == nil {
		return
	}
	m.InputParamsSecret.Namespace = namespace
}

// ActionFilter defines filtering options for Actions
type ActionFilter struct {
	Phase        *v1alpha1.ActionPhase
	NameRegex    *regexp.Regexp
	InterfaceRef *v1alpha1.ManifestReference
}

// AllAllowed returns true if all Actions are allowed.
func (f *ActionFilter) AllAllowed() bool {
	return f == nil || (f.Phase == nil && f.NameRegex == nil && f.InterfaceRef == nil)
}

// ZapFields returns zap logger filed used for debug purposes.
func (f *ActionFilter) ZapFields() []zap.Field {
	var out []zap.Field
	if f.Phase != nil {
		out = append(out, zap.String("status.phase", string(*f.Phase)))
	}
	if f.NameRegex != nil {
		out = append(out, zap.String("metadata.name", f.NameRegex.String()))
	}
	return out
}

// Match returns true if a given Action match filter criteria.
func (f *ActionFilter) Match(item v1alpha1.Action) bool {
	if f.Phase != nil && *f.Phase != item.Status.Phase {
		return false
	}

	if f.NameRegex != nil && !f.NameRegex.MatchString(item.Name) {
		return false
	}

	if f.InterfaceRef != nil {
		if f.InterfaceRef.Path != item.Spec.ActionRef.Path {
			return false
		}

		if f.InterfaceRef.Revision != nil && f.InterfaceRef.Revision != item.Spec.ActionRef.Revision {
			return false
		}
	}

	return true
}

// AdvancedModeContinueRenderingInput is used for continuing Action rendering in advanced mode.
type AdvancedModeContinueRenderingInput struct {
	// TypeInstances that are optional for a given rendering iteration
	TypeInstances *[]v1alpha1.InputTypeInstance
}
