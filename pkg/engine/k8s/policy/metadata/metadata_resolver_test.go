package metadata_test

import (
	"context"
	"testing"

	"capact.io/capact/pkg/engine/k8s/policy/metadata"

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
		HubCli             metadata.HubClient
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
			HubCli:   &fakeHub{ShouldRun: true, ExpectedIDLen: 12},
			Expected: fixComplexPolicyWithTypeRef(),
		},
		{
			Name:  "Partial result",
			Input: fixComplexPolicyWithoutTypeRef(),
			HubCli: &fakeHub{ShouldRun: true, ExpectedIDLen: 12, IgnoreIDs: map[string]struct{}{
				"id2": {}, "id4": {}, // required
				"id8": {},             // additional
				"id9": {}, "id11": {}, // backend
			},
			},
			ExpectedErrMessage: ptr.String(
				heredoc.Doc(`
				5 errors occurred:
					* missing Type reference for RequiredTypeInstance (ID: "id2", description: "ID 2")
					* missing Type reference for RequiredTypeInstance (ID: "id4")
					* missing Type reference for AdditionalTypeInstance (ID: "id8", name: "ID8")
					* missing Type reference for BackendTypeInstance (ID: "id9", description: "ID 9")
					* missing Type reference for BackendTypeInstance (ID: "id11", description: "ID 11")`,
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
			resolver := metadata.NewResolver(tc.HubCli)

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
