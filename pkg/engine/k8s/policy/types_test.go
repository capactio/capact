package policy_test

import (
	"testing"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"github.com/stretchr/testify/assert"
)

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
