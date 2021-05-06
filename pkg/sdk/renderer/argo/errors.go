package argo

import (
	"github.com/pkg/errors"
)

// NOTE: Change the error to Go struct if needed, e.g. someone needs to do such assertion `errors.Is(err, MaxDepthError)`

func NewActionReferencePatternError(actionRef string) error {
	return errors.Errorf("Action reference %q doesn't follow pattern <import_alias>.<method_name>", actionRef)
}

func NewMaxDepthError(limit int) error {
	return errors.Errorf("Exceeded maximum render depth level [max depth %d]", limit)
}

func NewActionImportsError(actionRef string) error {
	return errors.Errorf("Full path not found in Implementation imports for action %q", actionRef)
}

func NewRunnerContextRefEmptyError() error {
	return errors.Errorf("Empty Runner Context Secret reference")
}

func NewWorkflowNilError() error {
	return errors.New("workflow cannot be nil")
}

func NewEntrypointWorkflowIndexNotFoundError(entrypoint string) error {
	return errors.Errorf("cannot find workflow index specified by entrypoint %q", entrypoint)
}

func NewTypeReferenceNotFoundError(typeInstanceName string) error {
	return errors.Errorf("cannot find TypeReference for TypeInstance %s", typeInstanceName)
}

func NewMissingOwnerIDError() error {
	return errors.New("missing ownerID")
}
