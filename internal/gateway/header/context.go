package header

import (
	"context"
	"net/http"
)

type contextKey struct{}

// NewContext returns a copy of parent context with associated HTTP headers.
func NewContext(ctx context.Context, headers http.Header) context.Context {
	return context.WithValue(ctx, contextKey{}, headers)
}

// FromContext returns HTTP headers saved in a given context.
func FromContext(ctx context.Context) (http.Header, bool) {
	headers, ok := ctx.Value(contextKey{}).(http.Header)
	return headers, ok
}
