package upgrade

import "github.com/pkg/errors"

// Defines Capact Action related errors
var (
	ErrActionNotFinished   = errors.New("Action still not finished")
	ErrActionWithoutStatus = errors.New("Action doesn't have status")
	ErrActionDeleted       = errors.New("Action has been deleted, final state is unknown")
)

// NewErrAnotherUpgradeIsRunning returns an error indicating that another Capact upgrade is in progress.
func NewErrAnotherUpgradeIsRunning(actionName string) error {
	return errors.Errorf("Another upgrade action %s is currently running", actionName)
}
