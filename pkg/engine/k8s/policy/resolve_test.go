package policy_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/cli/heredoc"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveTypeInstanceMetadata(t *testing.T) {
	// given
	tests := []struct {
		Name               string
		Input              *policy.Policy
		HubCli             policy.HubClient
		Expected           *policy.Policy
		ExpectedErrMessage *string
	}{
		{
			Name:               "Nil HubCli",
			Input:              &policy.Policy{},
			HubCli:             nil,
			ExpectedErrMessage: ptr.String("hub client cannot be nil"),
		},
		{
			Name:               "Nil Policy",
			Input:              nil,
			ExpectedErrMessage: ptr.String("policy cannot be nil"),
		},
		{
			Name:     "Unresolved TypeRefs",
			Input:    fixComplexPolicyWithoutTypeRef(),
			HubCli:   &fakeHub{ShouldRun: true, ExpectedIDLen: 4},
			Expected: fixComplexPolicyWithTypeRef(),
		},
		{
			Name:   "Partial result",
			Input:  fixComplexPolicyWithoutTypeRef(),
			HubCli: &fakeHub{ShouldRun: true, ExpectedIDLen: 4, IgnoreIDs: map[string]struct{}{"id2": {}, "id4": {}}},
			ExpectedErrMessage: ptr.String(
				heredoc.Doc(`
				while TypeInstance metadata validation after resolving TypeRefs: while validating TypeInstance metadata for Policy: 2 errors occurred:
					* missing Type reference for TypeInstance "id2" (description: "ID 2")
					* missing Type reference for TypeInstance "id4" (description: "")`,
				),
			),
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
			err := policy.ResolveTypeInstanceMetadata(context.Background(), tc.HubCli, tc.Input)

			// then
			if tc.ExpectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *tc.ExpectedErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.Expected, tc.Input)
			}
		})
	}
}
