package client_test

import (
	"context"
	"testing"

	policyvalidation "capact.io/capact/pkg/sdk/validation/policy"

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
							TypeInstanceReference: policy.TypeInstanceReference{
								ID:          "my-uuid",
								Description: ptr.String("My UUID"),
								TypeRef: &types.TypeRef{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: "0.1.1",
								},
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
							TypeInstanceReference: policy.TypeInstanceReference{
								ID:          "my-uuid",
								Description: ptr.String("My UUID"),
							},
						},
					},
				},
			},
			expectedErrMessage: ptr.String(
				heredoc.Doc(`
				while validating Policy rule:
				- Metadata for "RequiredTypeInstance":
				    * missing Type reference for ID: "my-uuid", description: "My UUID"`),
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
							TypeInstanceReference: policy.TypeInstanceReference{
								ID:          "my-uuid",
								Description: ptr.String("My UUID"),
								TypeRef: &types.TypeRef{
									Path:     "cap.type.gcp.auth.service-account",
									Revision: "0.1.1",
								},
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

			validator := policyvalidation.NewValidator(hubCli)
			cli := client.NewPolicyEnforcedClient(hubCli, validator)

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
				while checking if additional TypeInstances from Policy are defined in Implementation manifest:
				- AdditionalTypeInstance "not-existing":
				    * cannot find such definition with exact name and Type reference "cap.type.sample1:0.1.0" in Implementation "impl"
				- AdditionalTypeInstance "not-existing2":
				    * cannot find such definition with exact name and Type reference "cap.type.sample2:0.1.0" in Implementation "impl"`,
				),
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
				while validating Policy rule:
				- Metadata for "AdditionalTypeInstance":
				    * missing Type reference for ID: "my-uuid", name: "foo"`),
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
			validator := policyvalidation.NewValidator(hubCli)
			cli := client.NewPolicyEnforcedClient(hubCli, validator)

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
		implRev                  gqlpublicapi.ImplementationRevision
		expectedParamsCollection types.ParametersCollection
		expectedErrMessage       *string
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
			name: "Invalid and unknown parameters",
			implRev: fixImplementationRevisionWithAdditionalInputParams([]*gqlpublicapi.ImplementationAdditionalInputParameter{
				{
					Name: "foo",
					TypeRef: &gqlpublicapi.TypeReference{
						Path:     "cap.type.capactio.capact.validation.key-bool",
						Revision: "0.1.0",
					},
				},
				{
					Name: "bar",
					TypeRef: &gqlpublicapi.TypeReference{
						Path:     "cap.type.capactio.capact.validation.key-string",
						Revision: "0.1.0",
					},
				},
			}),
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					AdditionalParameters: []policy.AdditionalParametersToInject{
						{
							Name: "foo", Value: map[string]interface{}{
								"key": "string",
							},
						},
						{
							Name: "baz", Value: map[string]interface{}{
								"key": "string",
							},
						},
					},
				},
			},
			expectedErrMessage: ptr.String(
				heredoc.Doc(`
				while validating additional input parameters schemas:
				- AdditionalParameters "baz":
				    * Unknown parameter. Cannot validate it against JSONSchema.
				- AdditionalParameters "foo":
				    * key: Invalid type. Expected: boolean, given: string`),
			),
		},
		{
			name: "Multiple valid parameters",
			implRev: fixImplementationRevisionWithAdditionalInputParams([]*gqlpublicapi.ImplementationAdditionalInputParameter{
				{
					Name: "foo",
					TypeRef: &gqlpublicapi.TypeReference{
						Path:     "cap.type.capactio.capact.validation.key-bool",
						Revision: "0.1.0",
					},
				},
				{
					Name: "bar",
					TypeRef: &gqlpublicapi.TypeReference{
						Path:     "cap.type.capactio.capact.validation.key-string",
						Revision: "0.1.0",
					},
				},
			}),
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					AdditionalParameters: []policy.AdditionalParametersToInject{
						{
							Name: "foo", Value: map[string]interface{}{
								"key": true,
							},
						},
						{
							Name: "bar", Value: map[string]interface{}{
								"key": "string",
							},
						},
					},
				},
			},
			expectedParamsCollection: types.ParametersCollection{
				"foo": heredoc.Doc(`
					key: true
				`),
				"bar": heredoc.Doc(`
					key: string
				`),
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			hubCli, err := fake.NewFromLocal("testdata/hub", false)
			require.NoError(t, err)

			validator := policyvalidation.NewValidator(hubCli)
			cli := client.NewPolicyEnforcedClient(hubCli, validator)

			// when
			actual, err := cli.ListAdditionalInputToInjectBasedOnPolicy(context.Background(), tt.policyRule, tt.implRev)

			// then
			if tt.expectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *tt.expectedErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedParamsCollection, actual)
			}
		})
	}
}

func TestPolicyEnforcedClient_ListTypeInstancesBackendsBasedOnPolicy(t *testing.T) {
	tests := []struct {
		name string

		implRev            gqlpublicapi.ImplementationRevision
		policyRule         policy.Rule
		expectedBackends   map[string]policy.TypeInstanceBackend
		expectedErrMessage *string
		globalPolicy       policy.Policy
	}{
		{
			name: "Empty inject in policy rule",
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
			globalPolicy: policy.Policy{
				TypeInstance: policy.TypeInstancePolicy{
					Rules: []policy.RulesForTypeInstance{
						{
							TypeRef: types.ManifestRefWithOptRevision{
								Path: "cap.type.capactio.examples.message",
							},
							Backend: policy.TypeInstanceBackend{
								TypeInstanceReference: policy.TypeInstanceReference{
									ID: "ID1",
									TypeRef: &types.TypeRef{
										Path:     "cap.type.aws.secret-manager.storage",
										Revision: "0.1.0",
									},
									ExtendsHubStorage: true,
								},
							},
						},
					},
				},
			},
			expectedBackends: map[string]policy.TypeInstanceBackend{
				"cap.type.capactio.examples.message": {
					TypeInstanceReference: policy.TypeInstanceReference{
						ID: "ID1",
						TypeRef: &types.TypeRef{
							Path:     "cap.type.aws.secret-manager.storage",
							Revision: "0.1.0",
						},
						ExtendsHubStorage: true,
					}},
			},
		},
		{
			name: "Inject Helm storage",
			implRev: fixImplementationRevisionWithRequire(gqlpublicapi.ImplementationRequirement{
				AllOf: []*gqlpublicapi.ImplementationRequirementItem{
					{
						Alias: ptr.String("helm-storage"),
						TypeRef: &gqlpublicapi.TypeReference{
							Path:     "cap.type.helm.storage",
							Revision: "0.1.1",
						},
					},
				},
			}),
			policyRule: policy.Rule{
				Inject: &policy.InjectData{
					RequiredTypeInstances: []policy.RequiredTypeInstanceToInject{
						{
							TypeInstanceReference: policy.TypeInstanceReference{
								ID:          "ID2",
								Description: ptr.String("ID2"),
								TypeRef: &types.TypeRef{
									Path:     "cap.type.helm.storage",
									Revision: "0.1.1",
								},
								ExtendsHubStorage: true,
							},
						},
					},
				},
			},
			globalPolicy: policy.Policy{
				TypeInstance: policy.TypeInstancePolicy{
					Rules: []policy.RulesForTypeInstance{
						{
							TypeRef: types.ManifestRefWithOptRevision{
								Path: "cap.type.capactio.examples.message",
							},
							Backend: policy.TypeInstanceBackend{
								TypeInstanceReference: policy.TypeInstanceReference{
									ID: "ID1",
									TypeRef: &types.TypeRef{
										Path:     "cap.type.aws.secret-manager.storage",
										Revision: "0.1.0",
									},
									ExtendsHubStorage: true,
								},
							},
						},
					},
				},
			},
			expectedBackends: map[string]policy.TypeInstanceBackend{
				"cap.type.capactio.examples.message": {
					TypeInstanceReference: policy.TypeInstanceReference{
						ID: "ID1",
						TypeRef: &types.TypeRef{
							Path:     "cap.type.aws.secret-manager.storage",
							Revision: "0.1.0",
						},
						ExtendsHubStorage: true,
					},
				},
				"helm-storage": {
					TypeInstanceReference: policy.TypeInstanceReference{
						ID:          "ID2",
						Description: ptr.String("ID2"),
						TypeRef: &types.TypeRef{
							Path:     "cap.type.helm.storage",
							Revision: "0.1.1",
						},
						ExtendsHubStorage: true,
					},
				},
			},
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			// given
			hubCli := &fake.FileSystemClient{}

			validator := policyvalidation.NewValidator(hubCli)
			cli := client.NewPolicyEnforcedClient(hubCli, validator)
			cli.SetGlobalPolicy(tt.globalPolicy)

			// when
			actual, err := cli.ListTypeInstancesBackendsBasedOnPolicy(context.Background(), tt.policyRule, tt.implRev)

			// then
			if tt.expectedErrMessage != nil {
				require.Error(t, err)
				assert.EqualError(t, err, *tt.expectedErrMessage)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBackends, actual.GetAll())
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

func fixImplementationRevisionWithAdditionalInputTypeInstances(additionalTI []*gqlpublicapi.InputTypeInstance) gqlpublicapi.ImplementationRevision {
	impl := fixImplementationRevision("impl", "0.0.1")
	impl.Spec.AdditionalInput = &gqlpublicapi.ImplementationAdditionalInput{
		TypeInstances: additionalTI,
	}

	return impl
}

func fixImplementationRevisionWithAdditionalInputParams(additionalParams []*gqlpublicapi.ImplementationAdditionalInputParameter) gqlpublicapi.ImplementationRevision {
	impl := fixImplementationRevision("impl", "0.0.1")
	impl.Spec.AdditionalInput = &gqlpublicapi.ImplementationAdditionalInput{
		Parameters: additionalParams,
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
