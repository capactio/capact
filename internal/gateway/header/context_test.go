package header_test

import (
	"context"
	"net/http"
	"testing"

	header_forwarder "capact.io/capact/internal/gateway/header"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndReadFromContext_HappyPath(t *testing.T) {
	// given
	headers := http.Header{
		"foo": []string{"bar"},
		"bar": []string{"baz", "qux"},
	}
	ctx := context.Background()

	// when
	ctxWithNS := header_forwarder.NewContext(ctx, headers)
	readHeaders, ok := header_forwarder.FromContext(ctxWithNS)

	// then
	require.True(t, ok)
	assert.Equal(t, headers, readHeaders)
}
