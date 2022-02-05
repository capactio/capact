package policy_test

import (
	"testing"

	"capact.io/capact/internal/k8s-engine/graphql/domain/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_FromGraphQLInput_HappyPath(t *testing.T) {
	// given
	gqlInput := fixGQLInput()
	expectedModel := fixModel()

	c := policy.NewConverter()

	// when
	actualModel, err := c.FromGraphQLInput(gqlInput)

	// then
	require.NoError(t, err)
	assert.Equal(t, expectedModel, actualModel)
}

func TestConverter_FromGraphQLInput_Failure(t *testing.T) {
	// given
	gqlInput := fixGQLInput()

	// simulate wrong additional params format
	gqlInput.Interface.Rules[0].OneOf[0].Inject.AdditionalParameters[0].Value = "string"

	expErrMsg := `while converting 'OneOf' rules for "cap.interface.database.postgresql.install:0.1.0": while getting Policy inject data: while converting additional parameters: additional input cannot be converted to map[string]interface{}`
	c := policy.NewConverter()

	// when
	_, err := c.FromGraphQLInput(gqlInput)

	// then
	assert.EqualError(t, err, expErrMsg)
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
