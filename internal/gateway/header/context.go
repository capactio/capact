package header

import (
	"context"
	"net/http"
)

type contextKey struct{}

func NewContext(ctx context.Context, headers http.Header) context.Context {
	return context.WithValue(ctx, contextKey{}, headers)
}

func FromContext(ctx context.Context) (http.Header, bool) {
	headers, ok := ctx.Value(contextKey{}).(http.Header)
	return headers, ok
}
