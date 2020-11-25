package action

import (
	"github.com/pkg/errors"
)

var ErrActionNotFound = errors.New("action not found")

var ErrActionCancelledNotRunnable = errors.New("action is not runnable, as it has been already cancelled")

var ErrActionNotCancellable = errors.New("action is not run, so it cannot be cancelled")
