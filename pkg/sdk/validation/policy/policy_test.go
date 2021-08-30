package policy_test

import (
	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	policyvalidation "capact.io/capact/pkg/sdk/validation/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestValidator_ValidateTypeInstancesMetadata(t *testing.T) {
	// given
	validator := policyvalidation.NewValidator()
	tests := []struct {
		Name               string
		Input              policy.Policy
		ExpectedErrMessage *string
	}{
		{
			Name:  "Empty",
			Input: policy.Policy{},
		},
		{
			Name:  "Valid",
			Input: fixPolicyWithTypeRef(),
		},
		{
			Name:  "Invalid",
			Input: fixPolicyWithoutTypeRef(),
			ExpectedErrMessage: ptr.String(
				heredoc.Docf(`
				while validating TypeInstance metadata for Policy: 3 errors occurred:
					* missing Type reference for RequiredTypeInstance "id"
					* missing Type reference for RequiredTypeInstance "id2" (description: "ID 2")
					* missing Type reference for AdditionalTypeInstance "id3" (name: "id-3")`,
				),
			),
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			
			// when
			err := validator.ValidateTypeInstancesMetadata(tc.Input)

			// then
			if tc.ExpectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *tc.ExpectedErrMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidator_ValidateTypeInstanceMetadata(t *testing.T) {
	// given
	validator := policyvalidation.NewValidator()
	tests := []struct {
		Name               string
		Input              policy.Rule
		ExpectedErrMessage *string
	}{
		{
			Name:  "Empty",
			Input: policy.Rule{},
		},
		{
			Name:  "Valid",
			Input: fixPolicyWithTypeRef().Rules[0].OneOf[0],
		},
		{
			Name:  "Invalid",
			Input: fixPolicyWithoutTypeRef().Rules[0].OneOf[0],
			ExpectedErrMessage: ptr.String(
				heredoc.Doc(`
				while validating TypeInstance metadata for Policy: 3 errors occurred:
					* missing Type reference for RequiredTypeInstance "id"
					* missing Type reference for RequiredTypeInstance "id2" (description: "ID 2")
					* missing Type reference for AdditionalTypeInstance "id3" (name: "id-3")`,
				),
			),
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// when
			err := validator.ValidateTypeInstancesMetadataForRule(tc.Input)

			// then
			if tc.ExpectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *tc.ExpectedErrMessage)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPolicy_AreTypeInstancesMetadataResolved(t *testing.T) {
	// given
	validator := policyvalidation.NewValidator()
	tests := []struct {
		Name           string
		Input          policy.Policy
		ExpectedResult bool
	}{
		{
			Name:           "Empty",
			Input:          policy.Policy{},
			ExpectedResult: true,
		},
		{
			Name:           "False",
			Input:          fixPolicyWithoutTypeRef(),
			ExpectedResult: false,
		},
		{
			Name:           "True",
			Input:          fixPolicyWithTypeRef(),
			ExpectedResult: true,
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			// when
			res := validator.AreTypeInstancesMetadataResolved(tc.Input)

			// then
			assert.Equal(t, tc.ExpectedResult, res)
		})
	}
}
