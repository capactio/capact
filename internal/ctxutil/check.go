package ctxutil

import "context"

// ShouldExit returns true if a given context was canceled.
func ShouldExit(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
