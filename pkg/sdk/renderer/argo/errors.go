package argo

import (
	"fmt"
)

type MaxDepthError struct {
	limit int
}

func NewMaxDepthError(limit int) *MaxDepthError {
	return &MaxDepthError{limit: limit}
}

func (e *MaxDepthError) Error() string {
	return fmt.Sprintf("Exceeded maximum render depth level [max depth %d]", e.limit)
}

type ActionImportsError struct {
	actionRef string
}

func NewActionImportsError(actionRef string) *ActionImportsError {
	return &ActionImportsError{actionRef: actionRef}
}

func (e *ActionImportsError) Error() string {
	return fmt.Sprintf("Full path not found in Implementation imports for action %q", e.actionRef)
}
