package policy_test

import (
	"context"
	"testing"

	"capact.io/capact/internal/cli/heredoc"
	"capact.io/capact/internal/ptr"
	"capact.io/capact/pkg/engine/k8s/policy"
	gqlpublicapi "capact.io/capact/pkg/hub/api/graphql/public"
	"capact.io/capact/pkg/sdk/apis/0.0.1/types"
	"capact.io/capact/pkg/sdk/validation"
	policyvalidation "capact.io/capact/pkg/sdk/validation/policy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
)

func TestValidator_ValidateTypeInstancesMetadata(t *testing.T) {
	// given
	validator := policyvalidation.NewValidator(nil)
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
			Name:  "Policy Without TypeRefs",
			Input: fixPolicyWithoutTypeRef(),
			ExpectedErrMessage: ptr.String(
				heredoc.Docf(`
				- Metadata for "AdditionalTypeInstance":
				    * missing Type reference for ID: "id3", name: "id-3"
				- Metadata for "BackendTypeInstance":
				    * missing Type reference for ID: "00fd161c-01bd-47a6-9872-47490e11f996", description: "Vault TI"
				    * missing Type reference for ID: "31bb8355-10d7-49ce-a739-4554d8a40b63"
				    * missing Type reference for ID: "a36ed738-dfe7-45ec-acd1-8e44e8db893b", description: "Default Capact PostgreSQL backend"
				- Metadata for "RequiredTypeInstance":
				    * missing Type reference for ID: "id"
				    * missing Type reference for ID: "id2", description: "ID 2"`,
				),
			),
		},
		{
			Name:  "Policy with wrong backends for Types",
			Input: fixPolicyWithWrongBackendForTypeRef(),
			ExpectedErrMessage: ptr.String(
				heredoc.Docf(`
				- Metadata for "BackendTypeInstance":
				    * Type reference ID: "00fd161c-01bd-47a6-9872-47490e11f996", description: "Vault TI" is not a Hub storage
				    * Type reference ID: "31bb8355-10d7-49ce-a739-4554d8a40b63" is not a Hub storage
				    * Type reference ID: "a36ed738-dfe7-45ec-acd1-8e44e8db893b", description: "Default Capact PostgreSQL backend" is not a Hub storage`,
				),
			),
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// when
			res := validator.ValidateTypeInstancesMetadata(tc.Input)
			err := res.ErrorOrNil()

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

func TestValidator_IsTypeRefInjectableAndEqualToImplReq(t *testing.T) {
	// given
	validator := policyvalidation.NewValidator(nil)
	tests := []struct {
		Name           string
		TypeRef        *types.TypeRef
		ReqItem        *gqlpublicapi.ImplementationRequirementItem
		ExpectedResult bool
	}{
		{
			Name:    "Empty TypeRef",
			TypeRef: nil,
			ReqItem: &gqlpublicapi.ImplementationRequirementItem{
				TypeRef: &gqlpublicapi.TypeReference{
					Path:     "path",
					Revision: "revision",
				},
				Alias: ptr.String("alias"),
			},
			ExpectedResult: false,
		},
		{
			Name: "Empty ReqItem",
			TypeRef: &types.TypeRef{
				Path:     "path",
				Revision: "revision",
			},
			ReqItem:        nil,
			ExpectedResult: false,
		},
		{
			Name: "Different path",
			TypeRef: &types.TypeRef{
				Path:     "path1",
				Revision: "1.0.0",
			},
			ReqItem: &gqlpublicapi.ImplementationRequirementItem{
				TypeRef: &gqlpublicapi.TypeReference{
					Path:     "path2",
					Revision: "1.0.0",
				},
				Alias: ptr.String("alias"),
			},
			ExpectedResult: false,
		},
		{
			Name: "Different revision",
			TypeRef: &types.TypeRef{
				Path:     "path",
				Revision: "1.0.0",
			},
			ReqItem: &gqlpublicapi.ImplementationRequirementItem{
				TypeRef: &gqlpublicapi.TypeReference{
					Path:     "path",
					Revision: "0.1.1",
				},
				Alias: ptr.String("alias"),
			},
			ExpectedResult: false,
		},
		{
			Name: "Equal but empty alias",
			TypeRef: &types.TypeRef{
				Path:     "path",
				Revision: "revision",
			},
			ReqItem: &gqlpublicapi.ImplementationRequirementItem{
				TypeRef: &gqlpublicapi.TypeReference{
					Path:     "path",
					Revision: "revision",
				},
			},
			ExpectedResult: false,
		},
		{
			Name: "Equal",
			TypeRef: &types.TypeRef{
				Path:     "path",
				Revision: "revision",
			},
			ReqItem: &gqlpublicapi.ImplementationRequirementItem{
				TypeRef: &gqlpublicapi.TypeReference{
					Path:     "path",
					Revision: "revision",
				},
				Alias: ptr.String("foo"),
			},
			ExpectedResult: true,
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// when
			res := validator.IsTypeRefInjectableAndEqualToImplReq(tc.TypeRef, tc.ReqItem)

			// then
			assert.Equal(t, tc.ExpectedResult, res)
		})
	}
}

func TestValidator_ValidateTypeInstanceMetadata(t *testing.T) {
	// given
	validator := policyvalidation.NewValidator(nil)
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
			Input: fixPolicyWithTypeRef().Interface.Rules[0].OneOf[0],
		},
		{
			Name:  "Invalid",
			Input: fixPolicyWithoutTypeRef().Interface.Rules[0].OneOf[0],
			ExpectedErrMessage: ptr.String(
				heredoc.Doc(`
				- Metadata for "AdditionalTypeInstance":
				    * missing Type reference for ID: "id3", name: "id-3"
				- Metadata for "RequiredTypeInstance":
				    * missing Type reference for ID: "id"
				    * missing Type reference for ID: "id2", description: "ID 2"`,
				),
			),
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()

			// when
			res := validator.ValidateTypeInstancesMetadataForRule(tc.Input)
			err := res.ErrorOrNil()

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

func TestValidator_ValidateAdditionalInputParameters(t *testing.T) {
	// given
	impl := gqlpublicapi.ImplementationRevision{}
	require.NoError(t, yaml.Unmarshal(implementationRevisionRaw, &impl))

	tests := map[string]struct {
		givenHubTypeInstances []*gqlpublicapi.Type
		givenParameters       types.ParametersCollection
		expectedIssues        string
	}{
		"Happy path JSON": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
				validation.AWSElasticsearchTypeRevFixture(),
			},
			givenParameters: types.ParametersCollection{
				"additional-parameters": `{"key": "true"}`,
				"impl-specific-config":  `{"replicas": "3"}`,
			},
		},
		"Happy path YAML": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
				validation.AWSElasticsearchTypeRevFixture(),
			},
			givenParameters: types.ParametersCollection{
				"additional-parameters": `key: "true"`,
				"impl-specific-config":  `replicas: "3"`,
			},
		},
		"Not found `db-settings`": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
				validation.AWSElasticsearchTypeRevFixture(),
			},
			givenParameters: types.ParametersCollection{
				"db-settings": `{"key": true}`,
			},
			expectedIssues: heredoc.Doc(`
			    - AdditionalParameters "db-settings":
			        * Unknown parameter. Cannot validate it against JSONSchema.`),
		},
		"Invalid parameters": {
			givenHubTypeInstances: []*gqlpublicapi.Type{
				validation.AWSCredsTypeRevFixture(),
				validation.AWSElasticsearchTypeRevFixture(),
			},
			givenParameters: types.ParametersCollection{
				"additional-parameters": `{"key": true}`,
				"impl-specific-config":  `{"key": true}`,
			},
			expectedIssues: heredoc.Doc(`
			            	- AdditionalParameters "additional-parameters":
			            	    * key: Invalid type. Expected: string, given: boolean
			            	- AdditionalParameters "impl-specific-config":
			            	    * (root): replicas is required
			            	    * (root): Additional property key is not allowed`),
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			ctx := context.Background()
			fakeCli := &validation.FakeHubCli{
				Types: tc.givenHubTypeInstances,
			}

			validator := policyvalidation.NewValidator(fakeCli)

			// when
			implSchemas, err := validator.LoadAdditionalInputParametersSchemas(ctx, impl)
			// then
			require.NoError(t, err)
			require.Len(t, implSchemas, 2)

			// when
			result, err := validator.ValidateAdditionalInputParameters(ctx, implSchemas, tc.givenParameters)
			// then
			require.NoError(t, err)

			if tc.expectedIssues == "" {
				assert.NoError(t, result.ErrorOrNil())
			} else {
				assert.EqualError(t, result.ErrorOrNil(), tc.expectedIssues)
			}
		})
	}
}

func TestValidator_ValidateAdditionalTypeInstances(t *testing.T) {
	// given
	impl := gqlpublicapi.ImplementationRevision{}
	require.NoError(t, yaml.Unmarshal(implementationRevisionRaw, &impl))

	tests := map[string]struct {
		additionalTIsInPolicy []policy.AdditionalTypeInstanceToInject
		implRev               gqlpublicapi.ImplementationRevision
		expectedIssues        string
	}{
		"Empty": {
			additionalTIsInPolicy: []policy.AdditionalTypeInstanceToInject{},
			implRev:               fixImplementationRevisionWithAdditionalInputParams(nil),
		},
		"Undefined": {
			additionalTIsInPolicy: []policy.AdditionalTypeInstanceToInject{
				{
					AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{Name: "bar", ID: "uuid"},
					TypeRef:                         &types.ManifestRef{Path: "path", Revision: "revision2"},
				},
				{
					AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{Name: "baz", ID: "uuid"},
					TypeRef:                         &types.ManifestRef{Path: "path", Revision: "revision"},
				},
			},
			implRev: fixImplementationRevisionWithAdditionalInputParams([]*gqlpublicapi.InputTypeInstance{
				{
					Name:    "foo",
					TypeRef: &gqlpublicapi.TypeReference{Path: "path", Revision: "revision"},
					Verbs:   []gqlpublicapi.TypeInstanceOperationVerb{gqlpublicapi.TypeInstanceOperationVerbGet},
				}, {
					Name:    "bar",
					TypeRef: &gqlpublicapi.TypeReference{Path: "path", Revision: "revision"},
					Verbs:   []gqlpublicapi.TypeInstanceOperationVerb{gqlpublicapi.TypeInstanceOperationVerbGet},
				},
			}),
			expectedIssues: heredoc.Doc(`
			- AdditionalTypeInstance "bar":
			    * cannot find such definition with exact name and Type reference "path:revision2" in Implementation "impl"
			- AdditionalTypeInstance "baz":
			    * cannot find such definition with exact name and Type reference "path:revision" in Implementation "impl"`),
		},
		"Success": {
			additionalTIsInPolicy: []policy.AdditionalTypeInstanceToInject{
				{
					AdditionalTypeInstanceReference: policy.AdditionalTypeInstanceReference{Name: "foo", ID: "uuid"},
					TypeRef:                         &types.ManifestRef{Path: "path", Revision: "revision"},
				},
			},
			implRev: fixImplementationRevisionWithAdditionalInputParams([]*gqlpublicapi.InputTypeInstance{
				{
					Name:    "foo",
					TypeRef: &gqlpublicapi.TypeReference{Path: "path", Revision: "revision"},
					Verbs:   []gqlpublicapi.TypeInstanceOperationVerb{gqlpublicapi.TypeInstanceOperationVerbGet},
				},
			}),
			expectedIssues: "",
		},
	}
	for tn, tc := range tests {
		t.Run(tn, func(t *testing.T) {
			// given
			fakeCli := &validation.FakeHubCli{}
			validator := policyvalidation.NewValidator(fakeCli)

			// when
			result := validator.ValidateAdditionalTypeInstances(tc.additionalTIsInPolicy, tc.implRev)

			// then
			if tc.expectedIssues == "" {
				assert.NoError(t, result.ErrorOrNil())
			} else {
				assert.EqualError(t, result.ErrorOrNil(), tc.expectedIssues)
			}
		})
	}
}

func TestValidator_AreTypeInstancesMetadataResolved(t *testing.T) {
	// given
	validator := policyvalidation.NewValidator(nil)
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
			t.Parallel()
			// when
			res := validator.AreTypeInstancesMetadataResolved(tc.Input)

			// then
			assert.Equal(t, tc.ExpectedResult, res)
		})
	}
}

func fixImplementationRevisionWithAdditionalInputParams(additionalTI []*gqlpublicapi.InputTypeInstance) gqlpublicapi.ImplementationRevision {
	return gqlpublicapi.ImplementationRevision{
		Metadata: &gqlpublicapi.ImplementationMetadata{
			Path:   "impl",
			Prefix: ptr.String("rev"),
		},
		Spec: &gqlpublicapi.ImplementationSpec{
			AdditionalInput: &gqlpublicapi.ImplementationAdditionalInput{
				TypeInstances: additionalTI,
			},
		},
		Revision: "rev",
	}
}
