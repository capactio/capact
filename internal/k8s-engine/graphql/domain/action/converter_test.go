package action_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"projectvoltron.dev/voltron/internal/k8s-engine/graphql/domain/action"
)

func TestConverter_FromGraphQLInput_HappyPath(t *testing.T) {
	// given
	name := "foo"
	namespace := "bar"
	gqlInput := fixGQLInput()
	expectedModel := fixModel(name, namespace)

	c := action.NewConverter()

	// when
	actualModel := c.FromGraphQLInput(gqlInput, name, namespace)

	// then
	assert.Equal(t, expectedModel, actualModel)
}

func TestConverter_ToGraphQL_HappyPath(t *testing.T) {
	// given
	name := "foo"
	expectedGQLAction := fixGQLAction(t, name)
	k8sAction := fixK8sAction(t, name)

	c := action.NewConverter()

	// when
	gqlAction := c.ToGraphQL(k8sAction)

	// then
	assert.Equal(t, expectedGQLAction, gqlAction)
}
