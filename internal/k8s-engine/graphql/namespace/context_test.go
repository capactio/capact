package namespace_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gotest.tools/assert"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/namespace"
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
