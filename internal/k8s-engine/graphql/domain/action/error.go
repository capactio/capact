package action

import (
	"github.com/pkg/errors"
)

var ErrActionNotFound = errors.New("action not found")

var ErrActionCancelledNotRunnable = errors.New("action is not runnable, as it has been already cancelled")

var ErrActionNotCancellable = errors.New("action cannot be cancelled, as it is not run")

var ErrActionAdvancedRenderingDisabled = errors.New("action advanced rendering mode is disabled")

var ErrActionAdvancedRenderingIterationNotContinuable = errors.New("action advanced rendering iteration is not ready to be continued")

var ErrInvalidTypeInstanceSetProvidedForRenderingIteration = errors.New("invalid set of TypeInstances provided for a given rendering iteration")
