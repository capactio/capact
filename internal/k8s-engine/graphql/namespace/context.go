package namespace

import (
	"context"

	"github.com/pkg/errors"
)

type contextKey struct{}

var (
	// ErrMissingNamespaceInContext defines an error indicating that namespace was not found in a given context.
	ErrMissingNamespaceInContext = errors.New("cannot read namespace from context")
	// ErrNilContext defines an error indicating that a given context is nil.
	ErrNilContext = errors.New("context is nil")
)

// NewContext returns a copy of parent context with associated namespace.
func NewContext(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, contextKey{}, namespace)
}

// FromContext returns namespace saved in a given context.
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
