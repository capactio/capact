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
