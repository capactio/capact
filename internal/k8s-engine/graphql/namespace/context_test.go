package namespace_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/k8s-engine/graphql/namespace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSaveAndReadFromContext_HappyPath(t *testing.T) {
	// given
	ns := "namespace"
	ctx := context.Background()

	// when
	ctxWithNS := namespace.NewContext(ctx, ns)
	readNs, err := namespace.FromContext(ctxWithNS)

	// then
	require.NoError(t, err)
	assert.Equal(t, ns, readNs)
}
