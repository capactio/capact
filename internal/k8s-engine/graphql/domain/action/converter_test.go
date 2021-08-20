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
	v1 "k8s.io/api/core/v1"
)

func TestConverter_FromGraphQLInput_HappyPath1(t *testing.T) {
	// given
	const (
		name = "name"
	)

	tests := map[string]struct {
		paramsDefined        bool
		typeInstancesDefined bool
		policyDefined        bool
	}{
		"Should convert all inputs": {
			paramsDefined:        true,
			typeInstancesDefined: true,
			policyDefined:        true,
		},
		"Should convert only input parameters": {
			paramsDefined:        true,
			typeInstancesDefined: false,
			policyDefined:        false,
		},
		"Should convert only input TypeInstances": {
			paramsDefined:        false,
			typeInstancesDefined: true,
			policyDefined:        false,
		},
		"Should convert only Action policy": {
			paramsDefined:        false,
			typeInstancesDefined: false,
			policyDefined:        true,
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			c := action.NewConverter()

			var (
				givenGQLParams *graphql.JSON
				givenGQLTI     []*graphql.InputTypeInstanceData
				givenGQLPolicy *graphql.PolicyInput

				expectedModelParams *v1alpha1.InputParameters
				expectedModelPolicy *v1alpha1.ActionPolicy
				expectedModelSecret *v1.Secret
				expectedModelTI     *[]v1alpha1.InputTypeInstance
			)
			if tc.paramsDefined {
				givenGQLParams = fixGQLInputParameters()
				expectedModelParams = fixModelInputParameters(name)
			}
			if tc.policyDefined {
				givenGQLPolicy = fixGQLInputActionPolicy()
				expectedModelPolicy = fixModelInputPolicy(name)
			}
			if tc.typeInstancesDefined {
				givenGQLTI = fixGQLInputTypeInstances()
				expectedModelTI = fixModelInputTypeInstances()
			}
			expectedModelSecret = fixModelInputSecret(name, tc.paramsDefined, tc.policyDefined)

			givenGQLInput := fixGQLActionInput(name, givenGQLParams, givenGQLTI, givenGQLPolicy)
			expectedModel := fixActionModel(name, expectedModelParams, expectedModelTI, expectedModelPolicy, expectedModelSecret)

			// when
			actualModel, err := c.FromGraphQLInput(givenGQLInput)

			// then
			require.NoError(t, err)
			assert.Equal(t, expectedModel, actualModel)
		})
	}
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
	gqlAction, err := c.ToGraphQL(k8sAction)

	// then
	require.NoError(t, err)
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
		InterfaceRef: &graphql.ManifestReferenceInput{
			Path:     "cap.interface.test",
			Revision: nil,
		},
	}

	expectedK8sPhase := v1alpha1.AdvancedModeRenderingIterationActionPhase
	expectedModelActionFilter := model.ActionFilter{
		Phase:     &expectedK8sPhase,
		NameRegex: regexp.MustCompile(gqlNameRegex),
		InterfaceRef: &v1alpha1.ManifestReference{
			Path:     "cap.interface.test",
			Revision: nil,
		},
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
