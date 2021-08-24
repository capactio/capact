package policy_test

import (
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicy_ValidateTypeInstancesMetadata(t *testing.T) {
	// given
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
			Name:               "Invalid",
			Input:              fixPolicyWithoutTypeRef(),
			ExpectedErrMessage: ptr.String("while validating TypeInstance metadata for Policy: 2 errors occurred:\n\t* missing Type reference for TypeInstance \"id\" (description: \"\")\n\t* missing Type reference for TypeInstance \"id2\" (description: \"ID 2\")\n\n"),
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			// when
			err := tc.Input.ValidateTypeInstancesMetadata()

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

func TestRule_ValidateTypeInstanceMetadata(t *testing.T) {
	// given
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
			Name:               "Invalid",
			Input:              fixPolicyWithoutTypeRef().Rules[0].OneOf[0],
			ExpectedErrMessage: ptr.String("while validating TypeInstance metadata for Policy: 2 errors occurred:\n\t* missing Type reference for TypeInstance \"id\" (description: \"\")\n\t* missing Type reference for TypeInstance \"id2\" (description: \"ID 2\")\n\n"),
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			// when
			err := tc.Input.ValidateTypeInstanceMetadata()

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
