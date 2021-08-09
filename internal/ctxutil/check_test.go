package ctxutil_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/ctxutil"

	"github.com/stretchr/testify/assert"
)

func TestShouldExit(t *testing.T) {
	t.Run("Should notify about exit if context is canceled", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())

		// when
		cancel()
		shouldExit := ctxutil.ShouldExit(ctx)

		// then
		assert.True(t, shouldExit)
	})

	t.Run("Should return false if context is not canceled", func(t *testing.T) {
		// given
		ctx := context.Background()

		// when
		shouldExit := ctxutil.ShouldExit(ctx)

		// then
		assert.False(t, shouldExit)
	})
}
