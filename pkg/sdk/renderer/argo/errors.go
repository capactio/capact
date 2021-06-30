package argo

import (
	"github.com/pkg/errors"
)

// NOTE: Change the error to Go struct if needed, e.g. someone needs to do such assertion `errors.Is(err, MaxDepthError)`

// NewMaxDepthError indicates that the maximum depth of the nested actions was reached.
func NewMaxDepthError(limit int) error {
	return errors.Errorf("Exceeded maximum render depth level [max depth %d]", limit)
}

// NewRunnerContextRefEmptyError indicates that the reference for the runner context is empty.
func NewRunnerContextRefEmptyError() error {
	return errors.Errorf("Empty Runner Context Secret reference")
}

// NewWorkflowNilError indicates that there is no workflow defined in the manifest.
func NewWorkflowNilError() error {
	return errors.New("workflow cannot be nil")
}

// NewEntrypointWorkflowIndexNotFoundError indicates that the entrypoint template
// cannot be found in the workflow.
func NewEntrypointWorkflowIndexNotFoundError(entrypoint string) error {
	return errors.Errorf("cannot find workflow index specified by entrypoint %q", entrypoint)
}

// NewTypeReferenceNotFoundError indicates that the TypeReference for an TypeInstance
// cannot be found in the manifests.
func NewTypeReferenceNotFoundError(typeInstanceName string) error {
	return errors.Errorf("cannot find TypeReference for TypeInstance %s", typeInstanceName)
}

// NewMissingOwnerIDError indicates that the OwnerID
// for the workflow has not been set.
func NewMissingOwnerIDError() error {
	return errors.New("missing ownerID used to update TypeInstances")
}
