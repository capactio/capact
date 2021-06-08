package action

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

var ErrActionNotFound = errors.New("action not found")

var ErrActionNotReadyToRun = errors.New("action is not runnable")

var ErrActionCanceledNotRunnable = errors.New("action is not runnable, as it has been already canceled")

var ErrActionNotCancelable = errors.New("action cannot be canceled, as it is not run")

var ErrActionAdvancedRenderingDisabled = errors.New("action advanced rendering mode is disabled")

var ErrActionAdvancedRenderingIterationNotContinuable = errors.New("action advanced rendering iteration is not ready to be continued")

type ErrInvalidSetOfTypeInstancesForRenderingIteration struct {
	Names []string
}

func NewErrInvalidSetOfTypeInstancesForRenderingIteration(names []string) *ErrInvalidSetOfTypeInstancesForRenderingIteration {
	return &ErrInvalidSetOfTypeInstancesForRenderingIteration{Names: names}
}

func (e ErrInvalidSetOfTypeInstancesForRenderingIteration) Error() string {
	return fmt.Sprintf("invalid set of TypeInstances provided for a given rendering iteration: [ %s ]", strings.Join(e.Names, ", "))
}
