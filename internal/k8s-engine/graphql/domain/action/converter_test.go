package action_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/domain/action"
	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"
)

func TestConverter_FromGraphQLInput_HappyPath(t *testing.T) {
	// given
	const (
		name = "name"
	)
	gqlInput := fixGQLInput(name)
	expectedModel := fixModel(name)

	c := action.NewConverter()

	// when
	actualModel := c.FromGraphQLInput(gqlInput)

	// then
	assert.Equal(t, expectedModel, actualModel)
}

func TestConverter_ToGraphQL_HappyPath(t *testing.T) {
	// given
	const (
		name = "name"
		ns   = "ns"
	)

	expectedGQLAction := fixGQLAction(t, name)
	k8sAction := fixK8sAction(t, name, ns)

	c := action.NewConverter()

	// when
	gqlAction := c.ToGraphQL(k8sAction)

	// then
	assert.Equal(t, expectedGQLAction, gqlAction)
}

func TestConverter_FilterFromGraphQL_HappyPath(t *testing.T) {
	// given
	gqlPhase := graphql.ActionStatusPhaseAdvancedModeRenderingIteration
	gqlActionFilter := fixGQLActionFilter(&gqlPhase)

	expectedK8sPhase := v1alpha1.AdvancedModeRenderingIterationActionPhase
	expectedModelActionFilter := fixModelActionFilter(&expectedK8sPhase)

	c := action.NewConverter()

	// when
	modelActionFilter := c.FilterFromGraphQL(gqlActionFilter)

	// then
	assert.Equal(t, expectedModelActionFilter, modelActionFilter)
}

func TestConverter_AdvancedModeContinueRenderingInputFromGraphQL_HappyPath(t *testing.T) {
	// given
	gqlAdvancedModeIterationInput := fixGQLAdvancedRenderingIterationInput()

	expectedModelAdvancedModeIterationInput := fixModelAdvancedRenderingIterationInput()

	c := action.NewConverter()

	// when
	modelActionFilter := c.AdvancedModeContinueRenderingInputFromGraphQL(gqlAdvancedModeIterationInput)

	// then
	assert.Equal(t, expectedModelAdvancedModeIterationInput, modelActionFilter)
}
