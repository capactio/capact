package clusterpolicy

import (
	"fmt"
	"strings"
)

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

func NewUnsupportedAPIVersionError(constraintErrors []error) *UnsupportedAPIVersionError {
	return &UnsupportedAPIVersionError{constraintErrors: constraintErrors}
}
