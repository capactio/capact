package argoactions

import "context"

// Action is a interface with the Do method.
// The Do(context.Context) error method is used to execute an operation
// on the TypeInstances in the Local Hub.
type Action interface {
	Do(context.Context) error
}
