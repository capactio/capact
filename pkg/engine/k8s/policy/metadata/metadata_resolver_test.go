package metadata_test

import (
	"capact.io/capact/pkg/engine/k8s/policy/metadata"
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
		Input    *policy.Policy
		HubCli   metadata.HubClient
		Expected *policy.Policy
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
			HubCli:   &fakeHub{ShouldRun: true, ExpectedIDLen: 8},
			Expected: fixComplexPolicyWithTypeRef(),
		},
		{
			Name:   "Partial result",
			Input:  fixComplexPolicyWithoutTypeRef(),
			HubCli: &fakeHub{ShouldRun: true, ExpectedIDLen: 8, IgnoreIDs: map[string]struct{}{"id2": {}, "id4": {}, "id8": {}}},
			ExpectedErrMessage: ptr.String(
				heredoc.Doc(`
				while TypeInstance metadata validation after resolving TypeRefs: while validating TypeInstance metadata for Policy: 3 errors occurred:
					* missing Type reference for RequiredTypeInstance "id2" (description: "ID 2")
					* missing Type reference for RequiredTypeInstance "id4"
					* missing Type reference for AdditionalTypeInstance "id8" (name: "ID8")`,
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
			resolver := metadata.NewMetadataResolver(tc.HubCli)

			// when
			err := resolver.ResolveTypeInstanceMetadata(context.Background(), tc.Input)

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
