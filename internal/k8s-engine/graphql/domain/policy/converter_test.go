package policy_test

import (
	"testing"

	"capact.io/capact/internal/k8s-engine/graphql/domain/policy"
	"github.com/stretchr/testify/assert"
)

func TestConverter_FromGraphQLInput_HappyPath(t *testing.T) {
	// given
	gqlInput := fixGQLInput()
	expectedModel := fixModel()

	c := policy.NewConverter()

	// when
	actualModel := c.FromGraphQLInput(gqlInput)

	// then
	assert.Equal(t, expectedModel, actualModel)
}

func TestConverter_ToGraphQL_HappyPath(t *testing.T) {
	// given
	input := fixModel()
	expectedGQL := fixGQL()

	c := policy.NewConverter()

	// when
	actualGQL := c.ToGraphQL(input)

	// then
	assert.Equal(t, expectedGQL, actualGQL)
}
