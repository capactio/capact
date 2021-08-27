package client_test

import (
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

// ListImplementationRevisionForInterface tested in pkg/sdk/renderer/argo/renderer_test.go

func TestPolicyEnforcedClient_ListRequiredTypeInstancesToInjectBasedOnPolicy(t *testing.T) {
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
			name: "No TypeRef for injected required TypeInstance",
			implRev: fixImplementationRevisionWithRequire(gqlpublicapi.ImplementationRequirement{
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
					* missing Type reference for RequiredTypeInstance "my-uuid" (description: "My UUID")`),
			),
		},
		{
			name: "Inject GCP SA",
			implRev: fixImplementationRevisionWithRequire(gqlpublicapi.ImplementationRequirement{
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
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			hubCli := &fake.FileSystemClient{}
			cli := client.NewPolicyEnforcedClient(hubCli)

			// when
			actual, err := cli.ListRequiredTypeInstancesToInjectBasedOnPolicy(tt.policyRule, tt.implRev)

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

func TestPolicyEnforcedClient_ListAdditionalTypeInstancesToInjectBasedOnPolicy(t *testing.T) {
	tests := []struct {
		name string

		implRev               gqlpublicapi.ImplementationRevision
		policyRule            policy.Rule
		expectedTypeInstances []types.InputTypeInstanceRef
		expectedErrMessage    *string
	}{
		{
			name:    "Empty inject in policy",
			implRev: fixImplementationRevisionWithAdditionalInputTypeInstances(nil),
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{},
				},
			},
			expectedTypeInstances: nil,
		},
		{
			name:    "Non-existing additional TypeInstance to inject",
			implRev: fixImplementationRevisionWithAdditionalInputTypeInstances(nil),
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
						{
							AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
								ID:   "my-uuid",
								Name: "not-existing",
							},
							TypeRef: &types.ManifestRef{
								Path:     "cap.type.sample1",
								Revision: "0.1.0",
							},
						},
						{
							AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
								ID:   "my-uuid2",
								Name: "not-existing2",
							},
							TypeRef: &types.ManifestRef{
								Path:     "cap.type.sample2",
								Revision: "0.1.0",
							},
						},
					},
				},
			},
			expectedTypeInstances: nil,
			expectedErrMessage: ptr.String(
				heredoc.Doc(`
				while checking if additional TypeInstances from Policy are defined in Implementation manifest: 2 errors occurred:
					* cannot find additional TypeInstance with name "not-existing" (Type reference: "cap.type.sample1:0.1.0") in Implementation "impl"
					* cannot find additional TypeInstance with name "not-existing2" (Type reference: "cap.type.sample2:0.1.0") in Implementation "impl"`),
			),
		},
		{
			name: "No TypeRef for injected additional TypeInstance",
			implRev: fixImplementationRevisionWithAdditionalInputTypeInstances(
				[]*gqlpublicapi.InputTypeInstance{
					{
						Name: "foo",
						TypeRef: &gqlpublicapi.TypeReference{
							Path:     "cap.type.sample1",
							Revision: "0.1.0",
						},
						Verbs: []gqlpublicapi.TypeInstanceOperationVerb{gqlpublicapi.TypeInstanceOperationVerbGet},
					},
				},
			),
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
						{
							AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
								ID:   "my-uuid",
								Name: "foo",
							},
						},
					},
				},
			},
			expectedErrMessage: ptr.String(
				heredoc.Doc(`
				while validating Policy rule: while validating TypeInstance metadata for Policy: 1 error occurred:
					* missing Type reference for AdditionalTypeInstance "my-uuid" (name: "foo")`),
			),
		},
		{
			name: "Inject database",
			implRev: fixImplementationRevisionWithAdditionalInputTypeInstances(
				[]*gqlpublicapi.InputTypeInstance{
					{
						Name: "database",
						TypeRef: &gqlpublicapi.TypeReference{
							Path:     "cap.type.database",
							Revision: "0.1.0",
						},
						Verbs: []gqlpublicapi.TypeInstanceOperationVerb{gqlpublicapi.TypeInstanceOperationVerbGet},
					},
				},
			),
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					AdditionalTypeInstances: []policy.AdditionalTypeInstanceToInject{
						{
							AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{
								Name: "database",
								ID:   "my-uuid",
							},
							TypeRef: &types.ManifestRef{
								Path:     "cap.type.database",
								Revision: "0.1.0",
							},
						},
					},
				},
			},
			expectedTypeInstances: []types.InputTypeInstanceRef{
				{
					Name: "database",
					ID:   "my-uuid",
				},
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			hubCli := &fake.FileSystemClient{}
			cli := client.NewPolicyEnforcedClient(hubCli)

			// when
			actual, err := cli.ListAdditionalTypeInstancesToInjectBasedOnPolicy(tt.policyRule, tt.implRev)

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

func TestPolicyEnforcedClient_ListAdditionalInputToInjectBasedOnPolicy(t *testing.T) {
	// given
	tests := []struct {
		name string

		policyRule               policy.Rule
		expectedParamsCollection types.ParametersCollection
	}{
		{
			name: "Empty inject in policy",
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					AdditionalParameters: nil,
				},
			},
			expectedParamsCollection: nil,
		},
		{
			name: "Multiple parameters",
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					AdditionalParameters: []policy.AdditionalParametersToInject{
						{
							Name: "foo", Value: map[string]interface{}{
								"foo": map[string]interface{}{
									"string": "value",
								},
							},
						},
						{
							Name: "bar", Value: map[string]interface{}{
								"bar": true,
							},
						},
					},
				},
			},
			expectedParamsCollection: types.ParametersCollection{
				"foo": heredoc.Doc(`
					foo:
					  string: value
				`),
				"bar": heredoc.Doc(`
					bar: true
				`),
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			cli := client.NewPolicyEnforcedClient(nil)

			// when
			actual, err := cli.ListAdditionalInputToInjectBasedOnPolicy(tt.policyRule)

			// then
			require.NoError(t, err)
			assert.Equal(t, tt.expectedParamsCollection, actual)
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

func fixImplementationRevisionWithAdditionalInputTypeInstances(additionalTI []*gqlpublicapi.InputTypeInstance) gqlpublicapi.ImplementationRevision {
	impl := fixImplementationRevision("impl", "0.0.1")
	impl.Spec.AdditionalInput = &gqlpublicapi.ImplementationAdditionalInput{
		TypeInstances: additionalTI,
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
