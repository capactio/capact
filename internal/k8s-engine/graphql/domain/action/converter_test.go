package action_test

import (
	"regexp"
	"testing"

	"capact.io/capact/internal/k8s-engine/graphql/domain/action"
	"capact.io/capact/internal/k8s-engine/graphql/model"
	"capact.io/capact/pkg/engine/api/graphql"
	"capact.io/capact/pkg/engine/k8s/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	var (
		gqlPhase     = graphql.ActionStatusPhaseAdvancedModeRenderingIteration
		gqlNameRegex = "foo-*"
	)

	gqlActionFilter := graphql.ActionFilter{
		Phase:     &gqlPhase,
		NameRegex: &gqlNameRegex,
	}

	expectedK8sPhase := v1alpha1.AdvancedModeRenderingIterationActionPhase
	expectedModelActionFilter := model.ActionFilter{
		Phase:     &expectedK8sPhase,
		NameRegex: regexp.MustCompile(gqlNameRegex),
	}

	c := action.NewConverter()

	// when
	modelActionFilter, err := c.FilterFromGraphQL(&gqlActionFilter)

	// then
	require.NoError(t, err)
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
