package action

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// Defines GraphQL Action related errors.
var (
	ErrActionNotFound = errors.New("action not found")

	ErrActionNotReadyToRun = errors.New("action is not runnable")

	ErrActionCanceledNotRunnable = errors.New("action is not runnable, as it has been already canceled")

	ErrActionNotCancelable = errors.New("action cannot be canceled, as it is not run")

	ErrActionAdvancedRenderingDisabled = errors.New("action advanced rendering mode is disabled")

	ErrActionAdvancedRenderingIterationNotContinuable = errors.New("action advanced rendering iteration is not ready to be continued")
)

// InvalidSetOfTypeInstancesForRenderingIterationError defines an error indicating that some TypeInstances are
// not in the set of optional TypeInstances to provide.
type InvalidSetOfTypeInstancesForRenderingIterationError struct {
	Names []string
}

// NewInvalidSetOfTypeInstancesForRenderingIterationError returns a new InvalidSetOfTypeInstancesForRenderingIterationError instance.
func NewInvalidSetOfTypeInstancesForRenderingIterationError(names []string) *InvalidSetOfTypeInstancesForRenderingIterationError {
	return &InvalidSetOfTypeInstancesForRenderingIterationError{Names: names}
}

// Error returns error message.
func (e InvalidSetOfTypeInstancesForRenderingIterationError) Error() string {
	return fmt.Sprintf("invalid set of TypeInstances provided for a given rendering iteration: [ %s ]", strings.Join(e.Names, ", "))
}
