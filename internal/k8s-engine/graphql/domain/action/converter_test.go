package action_test

import (
	"testing"

	"projectvoltron.dev/voltron/pkg/engine/api/graphql"
	"projectvoltron.dev/voltron/pkg/engine/k8s/api/v1alpha1"

	"github.com/stretchr/testify/assert"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/domain/action"
)

func TestConverter_FromGraphQLInput_HappyPath(t *testing.T) {
	// given
	name := "foo"
	namespace := "bar"
	gqlInput := fixGQLInput(name)
	expectedModel := fixModel(name, namespace)

	c := action.NewConverter()

	// when
	actualModel := c.FromGraphQLInput(gqlInput, namespace)

	// then
	assert.Equal(t, expectedModel, actualModel)
}

func TestConverter_ToGraphQL_HappyPath(t *testing.T) {
	// given
	name := "foo"
	expectedGQLAction := fixGQLAction(t, name)
	k8sAction := fixK8sAction(t, name, "ns")

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
