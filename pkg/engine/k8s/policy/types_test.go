package policy_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPolicy_AreTypeInstancesMetadataResolved(t *testing.T) {
	// given
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
			// when
			res := tc.Input.AreTypeInstancesMetadataResolved()

			// then
			assert.Equal(t, tc.ExpectedResult, res)
		})
	}
}

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

func TestPolicy_ResolveTypeInstanceMetadata(t *testing.T) {
	// given
	tests := []struct {
		Name               string
		Input              policy.Policy
		HubCli             policy.HubClient
		Expected           policy.Policy
		ExpectedErrMessage *string
	}{
		{
			Name:               "Nil HubCli",
			Input:              policy.Policy{},
			HubCli:             nil,
			ExpectedErrMessage: ptr.String("hub client cannot be nil"),
		},
		{
			Name:     "Unresolved TypeRefs",
			Input:    fixComplexPolicyWithoutTypeRef(),
			HubCli:   &fakeHub{ShouldRun: true, ExpectedIDLen: 4},
			Expected: fixComplexPolicyWithTypeRef(),
		},
		{
			Name:               "Partial result",
			Input:              fixComplexPolicyWithoutTypeRef(),
			HubCli:             &fakeHub{ShouldRun: true, ExpectedIDLen: 4, IgnoreIDs: map[string]struct{}{"id2": {}, "id4": {}}},
			ExpectedErrMessage: ptr.String("while TypeInstance metadata validation after resolving TypeRefs: while validating TypeInstance metadata for Policy: 2 errors occurred:\n\t* missing Type reference for TypeInstance \"id2\" (description: \"ID 2\")\n\t* missing Type reference for TypeInstance \"id4\" (description: \"\")\n\n"),
		},
		{
			Name:     "Already resolved",
			HubCli:   &fakeHub{ShouldRun: false},
			Input:    fixComplexPolicyWithTypeRef(),
			Expected: fixComplexPolicyWithTypeRef(),
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			// when
			policy := tc.Input
			err := policy.ResolveTypeInstanceMetadata(context.Background(), tc.HubCli)

			// then
			if tc.ExpectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *tc.ExpectedErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.Expected, policy)
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

func TestRule_RequiredTypeInstancesToInject(t *testing.T) {
	// given
	tests := []struct {
		Name     string
		Input    policy.Rule
		Expected []policy.RequiredTypeInstanceToInject
	}{
		{
			Name:     "Nil Inject",
			Input:    policy.Rule{},
			Expected: nil,
		},
		{
			Name:  "Inject with RequiredTypeInstances",
			Input: fixPolicyWithTypeRef().Rules[0].OneOf[0],
			Expected: []policy.RequiredTypeInstanceToInject{
				{
					RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
						ID: "id",
					},
					TypeRef: &types.ManifestRef{
						Path:     "cap.type.sample",
						Revision: "0.1.0",
					},
				},
				{
					RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
						ID:          "id2",
						Description: ptr.String("ID 2"),
					},
					TypeRef: &types.ManifestRef{
						Path:     "cap.type.sample2",
						Revision: "0.2.0",
					},
				},
			},
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			// when
			actual := tc.Input.RequiredTypeInstancesToInject()

			// then
			assert.Equal(t, tc.Expected, actual)
		})
	}
}
