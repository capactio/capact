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
	ctx := context.TODO()

	// when
	ctxWithNS := namespace.SaveToContext(ctx, ns)
	readNs, err := namespace.ReadFromContext(ctxWithNS)

	// then
	require.NoError(t, err)
	assert.Equal(t, ns, readNs)
}
