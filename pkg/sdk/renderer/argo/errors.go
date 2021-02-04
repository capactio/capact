package argo

import (
	"github.com/pkg/errors"
)

func NewActionReferencePatternError(actionRef string) error {
	return errors.Errorf("Action reference %q doesn't follow pattern <import_alias>.<method_name>", actionRef)
}

func NewMaxDepthError(limit int) error {
	return errors.Errorf("Exceeded maximum render depth level [max depth %d]", limit)
}

func NewActionImportsError(actionRef string) error {
	return errors.Errorf("Full path not found in Implementation imports for action %q", actionRef)
}
