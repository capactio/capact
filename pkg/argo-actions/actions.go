package argoactions

import "context"

type Action interface {
	Do(context.Context) error
}
