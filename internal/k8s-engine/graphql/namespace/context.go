package namespace

import (
	"context"

	"github.com/pkg/errors"
)

type contextKey struct{}

var ErrMissingNamespaceInContext = errors.New("cannot read namespace from context")
var ErrNilContext = errors.New("context is nil")

func NewContext(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, contextKey{}, namespace)
}

func FromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrNilContext
	}

	value := ctx.Value(contextKey{})
	ns, ok := value.(string)
	if !ok {
		return "", ErrMissingNamespaceInContext
	}

	return ns, nil
}
