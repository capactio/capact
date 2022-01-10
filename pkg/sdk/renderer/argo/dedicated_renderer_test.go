package argo

import (
	"testing"

	hubclient "capact.io/capact/pkg/hub/client"
	"capact.io/capact/pkg/hub/client/fake"
	policyvalidation "capact.io/capact/pkg/sdk/validation/policy"
	"github.com/stretchr/testify/require"
)

func createFakeDedicatedRendererObject(t *testing.T) *dedicatedRenderer {
	fakeCli, err := fake.NewFromLocal("testdata/hub", true)
	require.NoError(t, err)

	genUUID := func() string { return "uuid" }
	typeInstanceHandler := NewTypeInstanceHandler("alpine:3.7")
	typeInstanceHandler.SetGenUUID(genUUID)

	policyIOValidator := policyvalidation.NewValidator(fakeCli)
	policyEnforcedClient := hubclient.NewPolicyEnforcedClient(fakeCli, policyIOValidator)
	opts := []RendererOption{}
	maxDepth := 20

	return newDedicatedRenderer(maxDepth, policyEnforcedClient, typeInstanceHandler, opts...)
}

func TestCapactWhenContainDashes(t *testing.T) {
	//given
	dedicatedRenderer := createFakeDedicatedRendererObject(t)
	params := &mapEvalParameters{}
	params.Set("postgresql-db")

	//when
	result, err := dedicatedRenderer.evaluateWhenExpression(params, "postgresql-db == nil")
	isNotExist, ok := result.(bool)
	require.True(t, ok)

	//then
	require.NoError(t, err)
	require.False(t, isNotExist)
}
