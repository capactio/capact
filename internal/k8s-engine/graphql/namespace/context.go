package namespace

import (
	"context"

	"github.com/pkg/errors"
)

type contextKey struct{}

var ErrMissingNamespaceInContext = errors.New("cannot read namespace from context")
var ErrEmptyContext = errors.New("context is empty")

func SaveToContext(ctx context.Context, namespace string) context.Context {
	return context.WithValue(ctx, contextKey{}, namespace)
}

func ReadFromContext(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", ErrEmptyContext
	}

	value := ctx.Value(contextKey{})
	ns, ok := value.(string)
	if !ok {
		return "", ErrMissingNamespaceInContext
	}

	return ns, nil
}
