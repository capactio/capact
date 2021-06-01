package client_test

import (
	"testing"

	gqlpublicapi "capact.io/capact/pkg/och/api/graphql/public"
	"capact.io/capact/pkg/och/client"
	"github.com/stretchr/testify/assert"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/clusterpolicy"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// ListImplementationRevisionForInterface tested in pkg/sdk/renderer/argo

func TestPolicyEnforcedClient_ListTypeInstancesToInjectBasedOnPolicy(t *testing.T) {
	tests := []struct {
		name string

		implRev               gqlpublicapi.ImplementationRevision
		policyRule            clusterpolicy.Rule
		expectedTypeInstances []types.InputTypeInstanceRef
	}{
		{
			name: "Empty inject in policy",
			implRev: fixImplementationRevisionWithRequire(gqlpublicapi.ImplementationRequirement{
				Prefix: "cap.core.type.platform",
				AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
					{
						TypeRef: &gqlpublicapi.TypeReference{
							Path:     "cap.type.gcp.sa",
							Revision: "0.1.1",
						},
					},
				},
			}),
			policyRule: clusterpolicy.Rule{
				Inject: &clusterpolicy.InjectData{
					TypeInstances: []clusterpolicy.TypeInstanceToInject{},
				},
			},
			expectedTypeInstances: nil,
		},
		{
			name: "Unused TypeInstance to inject",
			implRev: fixImplementationRevisionWithRequire(gqlpublicapi.ImplementationRequirement{
				Prefix: "cap.core.type.platform",
				OneOf: []*gqlpublicapi.ImplementationRequirementItem{
					{
						TypeRef: &gqlpublicapi.TypeReference{
							Path: "kubernetes",
						},
					},
				},
				AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
					{
						TypeRef: &gqlpublicapi.TypeReference{
							Path: "kubernetes",
						},
					},
				},
				AllOf: []*gqlpublicapi.ImplementationRequirementItem{
					{
						TypeRef: &gqlpublicapi.TypeReference{
							Path: "kubernetes",
						},
					},
				},
			}),
			policyRule: clusterpolicy.Rule{
				Inject: &clusterpolicy.InjectData{
					TypeInstances: []clusterpolicy.TypeInstanceToInject{
						{
							ID: "my-uuid",
							TypeRef: types.ManifestRef{
								Path:     "cap.type.gcp.auth.service-account",
								Revision: ptr.String("0.1.1"),
							},
						},
					},
				},
			},
			expectedTypeInstances: nil,
		},
		{
			name: "Inject GCP SA with specific revision",
			implRev: fixImplementationRevisionWithRequire(gqlpublicapi.ImplementationRequirement{
				Prefix: "cap.type.gcp.auth",
				AllOf: []*gqlpublicapi.ImplementationRequirementItem{
					{
						Alias: ptr.String("gcp-sa"),
						TypeRef: &gqlpublicapi.TypeReference{
							Path:     "cap.type.gcp.auth.service-account",
							Revision: "0.1.1",
						},
					},
				},
			}),
			policyRule: clusterpolicy.Rule{
				Inject: &clusterpolicy.InjectData{
					TypeInstances: []clusterpolicy.TypeInstanceToInject{
						{
							ID: "my-uuid",
							TypeRef: types.ManifestRef{
								Path:     "cap.type.gcp.auth.service-account",
								Revision: ptr.String("0.1.1"),
							},
						},
					},
				},
			},
			expectedTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "gcp-sa",
					ID:   "my-uuid",
				},
			},
		},
		{
			name: "Inject GCP SA with any revision",
			implRev: fixImplementationRevisionWithRequire(gqlpublicapi.ImplementationRequirement{
				Prefix: "cap.type.gcp.auth",
				AnyOf: []*gqlpublicapi.ImplementationRequirementItem{
					{
						Alias: ptr.String("gcp-sa"),
						TypeRef: &gqlpublicapi.TypeReference{
							Path:     "cap.type.gcp.auth.service-account",
							Revision: "0.1.1",
						},
					},
				},
			}),
			policyRule: clusterpolicy.Rule{
				Inject: &clusterpolicy.InjectData{
					TypeInstances: []clusterpolicy.TypeInstanceToInject{
						{
							ID: "my-uuid",
							TypeRef: types.ManifestRef{
								Path: "cap.type.gcp.auth.service-account",
							},
						},
					},
				},
			},
			expectedTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "gcp-sa",
					ID:   "my-uuid",
				},
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			cli := client.NewPolicyEnforcedClient(nil)

			// when
			actual := cli.ListTypeInstancesToInjectBasedOnPolicy(tt.policyRule, tt.implRev)

			// then
			assert.Equal(t, tt.expectedTypeInstances, actual)
		})
	}
}

func fixImplementationRevisionWithRequire(req gqlpublicapi.ImplementationRequirement) gqlpublicapi.ImplementationRevision {
	impl := fixImplementationRevision("impl", "0.0.1")
	impl.Spec.Requires = []*gqlpublicapi.ImplementationRequirement{
		&req,
	}

	return impl
}

func fixImplementationRevision(path, rev string) gqlpublicapi.ImplementationRevision {
	return gqlpublicapi.ImplementationRevision{
		Metadata: &gqlpublicapi.ImplementationMetadata{
			Path:   path,
			Prefix: ptr.String(path),
		},
		Spec:     &gqlpublicapi.ImplementationSpec{},
		Revision: rev,
	}
}
