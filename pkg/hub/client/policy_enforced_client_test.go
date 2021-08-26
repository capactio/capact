package client_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/cli/heredoc"

	"capact.io/capact/pkg/hub/client/fake"
	"github.com/stretchr/testify/require"

	"capact.io/capact/pkg/engine/k8s/policy"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/hub/client"
	"github.com/stretchr/testify/assert"

	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
)

// ListImplementationRevisionForInterface tested in pkg/sdk/renderer/argo

func TestPolicyEnforcedClient_ListTypeInstancesToInjectBasedOnPolicy(t *testing.T) {
	tests := []struct {
		name string

		implRev               gqlpublicapi.ImplementationRevision
		policyRule            policy.Rule
		expectedTypeInstances []types.InputTypeInstanceRef
		expectedErrMessage    *string
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
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{},
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
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
						{
							RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
								ID:          "my-uuid",
								Description: ptr.String("My UUID"),
							},
							TypeRef: &types.ManifestRef{
								Path:     "cap.type.gcp.auth.service-account",
								Revision: "0.1.1",
							},
						},
					},
				},
			},
			expectedTypeInstances: nil,
		},
		{
			name: "Inject GCP SA",
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
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
						{
							RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
								ID:          "my-uuid",
								Description: ptr.String("My UUID"),
							},
							TypeRef: &types.ManifestRef{
								Path:     "cap.type.gcp.auth.service-account",
								Revision: "0.1.1",
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
			name: "No TypeRef for injected required TypeInstance",
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
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
						{
							RequiredTypeInstanceReference: policy.RequiredTypeInstanceReference{
								ID:          "my-uuid",
								Description: ptr.String("My UUID"),
							},
						},
					},
				},
			},
			expectedErrMessage: ptr.String(
				heredoc.Doc(`
				while validating Policy rule: while validating TypeInstance metadata for Policy: 1 error occurred:
					* missing Type reference for TypeInstance "my-uuid" (description: "My UUID")`),
			),
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			hubCli := &fake.FileSystemClient{}
			cli := client.NewPolicyEnforcedClient(hubCli)

			// when
			actual, err := cli.ListRequiredTypeInstancesToInjectBasedOnPolicy(context.Background(), tt.policyRule, tt.implRev)

			// then
			if tt.expectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *tt.expectedErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedTypeInstances, actual)
			}
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
