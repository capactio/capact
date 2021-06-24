package policy

import (
	"fmt"
	"strings"
)

// UnsupportedAPIVersionError indicates that the Policy APIVersion is not supported.
type UnsupportedAPIVersionError struct {
	constraintErrors []error
}

func (e UnsupportedAPIVersionError) Error() string {
	var errMsgs []string
	for _, err := range e.constraintErrors {
		errMsgs = append(errMsgs, err.Error())
	}
	return fmt.Sprintf("unsupported API version: %s", strings.Join(errMsgs, ", "))
}

// NewUnsupportedAPIVersionError returns a new UnsupportedAPIVersionError error.
func NewUnsupportedAPIVersionError(constraintErrors []error) *UnsupportedAPIVersionError {
	return &UnsupportedAPIVersionError{constraintErrors: constraintErrors}
}
